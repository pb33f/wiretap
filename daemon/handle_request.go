// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package daemon

import (
	"crypto/tls"
	_ "embed"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/pb33f/libopenapi-validator/errors"
	"github.com/pb33f/ranch/model"
	configModel "github.com/pb33f/wiretap/config"
	"github.com/pb33f/wiretap/shared"
)

//go:embed templates/socket-include.html
var staticTemplate string

type staticTemplateModel struct {
	OriginalContent string
	WebSocketPort   string
}

func (ws *WiretapService) handleHttpRequest(request *model.Request) {

	// determine if this is a request for a file or not.
	if ws.config.StaticDir != "" {
		fp := filepath.Join(ws.config.StaticDir, request.HttpRequest.URL.Path)

		isRoot := false
		// check if this is a static path catch-all
		if len(ws.config.StaticPathsCompiled) > 0 {
			for key := range ws.config.StaticPathsCompiled {
				if ws.config.StaticPathsCompiled[key].Match(request.HttpRequest.URL.Path) {
					fp = filepath.Join(ws.config.StaticDir, ws.config.StaticIndex)
					isRoot = true
					break
				}
			}
		}

		// check if this is a root request
		if fp == ws.config.StaticDir {
			isRoot = true
			fp = filepath.Join(ws.config.StaticDir, "index.html")
		}
		localStat, _ := os.Stat(fp)
		if localStat != nil {

			if isRoot {

				// if this root, we need to modify the index to inject some JS.
				tmpFile, _ := os.CreateTemp("", "index.html")
				defer os.Remove(tmpFile.Name())

				tmpl, _ := template.New("index").Parse(staticTemplate)
				indexBytes, _ := os.ReadFile(fp)

				// prep a model
				m := staticTemplateModel{
					OriginalContent: string(indexBytes),
					WebSocketPort:   ws.config.WebSocketPort,
				}

				// execute the new template
				_ = tmpl.Execute(tmpFile, m)

				ws.config.Logger.Info("[wiretap] static file request", "url", request.HttpRequest.URL.String(), "code", 200)

				// serve it.
				http.ServeFile(request.HttpResponseWriter, request.HttpRequest, tmpFile.Name())
				return
			}

			if !localStat.IsDir() {

				ws.config.Logger.Info("[wiretap] static file request", "url", request.HttpRequest.URL.String(), "code", 200)

				http.ServeFile(request.HttpResponseWriter, request.HttpRequest, fp)
				return
			}
		}
	}
	var returnedResponse *http.Response
	var returnedError error

	configStore, _ := ws.controlsStore.Get(shared.ConfigKey)
	config := configStore.(*shared.WiretapConfiguration)

	if config.Headers == nil || len(config.Headers.DropHeaders) == 0 {
		config.Headers = &shared.WiretapHeaderConfig{
			DropHeaders: []string{},
		}
	}

	dropHeaders, injectHeaders, auth := ws.getHeadersAndAuth(config, request)

	newReq := CloneExistingRequest(CloneRequest{
		Request:       request.HttpRequest,
		Protocol:      config.RedirectProtocol,
		Host:          config.RedirectHost,
		Port:          config.RedirectPort,
		DropHeaders:   dropHeaders,
		InjectHeaders: injectHeaders,
		Auth:          auth,
		Variables:     config.CompiledVariables,
	})

	apiRequest := CloneExistingRequest(CloneRequest{
		Request:       request.HttpRequest,
		Protocol:      config.RedirectProtocol,
		Host:          config.RedirectHost,
		BasePath:      config.RedirectBasePath,
		Port:          config.RedirectPort,
		DropHeaders:   dropHeaders,
		InjectHeaders: injectHeaders,
		Auth:          auth,
		Variables:     config.CompiledVariables,
	})

	if newReq == nil || apiRequest == nil {
		ws.config.Logger.Error("[wiretap] unable to clone API request, failed", "url", request.HttpRequest.URL.String())
		return
	}

	var requestErrors []*errors.ValidationError
	var responseErrors []*errors.ValidationError

	ws.config.Logger.Info("[wiretap] handling API request", "url", request.HttpRequest.URL.String())

	// short-circuit if we're using mock mode, there is no API call to make.
	if ws.config.MockMode {
		ws.config.Logger.Info("MockMode enabled; skipping validation")
		ws.handleMockRequest(request, config, newReq)
		return
	} else if configModel.IgnoreValidationOnPath(apiRequest.URL.Path, ws.config) && !configModel.PathValidationAllowListed(apiRequest.URL.Path, ws.config) {
		ws.config.Logger.Info(
			fmt.Sprintf("Request on validation ignored path: %s ; skipping validation", apiRequest.URL.Path))
	} else if ws.config.HardErrors { // check if we're going to fail hard on validation errors. (default is to skip this)
		// validate the request synchronously
		requestErrors = ws.ValidateRequest(request, newReq)
	} else {
		// validate the request asynchronously
		go ws.ValidateRequest(request, newReq)
	}

	// call the API being requested.
	returnedResponse, returnedError = ws.callAPI(apiRequest)

	if returnedResponse == nil && returnedError != nil {
		config.Logger.Info("[wiretap] request failed", "url", apiRequest.URL.String(), "code", 500,
			"error", returnedError.Error())
		go ws.broadcastResponseError(request, CloneExistingResponse(returnedResponse), returnedError)
		request.HttpResponseWriter.WriteHeader(500)
		wtError := shared.GenerateError("Unable to call API", 500, returnedError.Error(), "", returnedResponse)
		_, _ = request.HttpResponseWriter.Write(shared.MarshalError(wtError))
		return

	} else {

		// check if we're going to fail hard on validation errors. (default is to skip this)
		if ws.config.HardErrors {
			// validate response
			responseErrors = ws.ValidateResponse(request, CloneExistingResponse(returnedResponse))
		} else {
			// validate response async
			go ws.ValidateResponse(request, CloneExistingResponse(returnedResponse))
		}
	}

	// check if this path has a delay set.
	delay := configModel.FindPathDelay(request.HttpRequest.URL.Path, config)
	if delay > 0 {
		time.Sleep(time.Duration(delay) * time.Millisecond) // simulate a slow response, configured for path.
	} else {
		if config.GlobalAPIDelay > 0 {
			time.Sleep(time.Duration(config.GlobalAPIDelay) * time.Millisecond) // simulate a slow response.
		}
	}

	body, _ := io.ReadAll(returnedResponse.Body)
	headers := ExtractHeaders(returnedResponse)

	// wiretap needs to work from anywhere, so allow everything.
	setCORSHeaders(headers)

	if config.StrictRedirectLocation && is3xxStatusCode(returnedResponse.StatusCode) {
		setStrictLocationHeader(config, headers)
	}

	// write headers
	for k, v := range headers {
		for _, j := range v {
			request.HttpResponseWriter.Header().Set(k, fmt.Sprint(j))
		}
	}
	config.Logger.Info("[wiretap] request completed", "url", request.HttpRequest.URL.String(), "code", returnedResponse.StatusCode)

	// if there are validation errors, set an error code
	requestCode := config.HardErrorCode
	returnCode := config.HardErrorReturnCode

	switch {
	case config.HardErrors && len(requestErrors) > 0 && len(responseErrors) <= 0:
		request.HttpResponseWriter.WriteHeader(requestCode)
	case config.HardErrors && len(requestErrors) <= 0 && len(responseErrors) > 0:
		request.HttpResponseWriter.WriteHeader(returnCode)
	case config.HardErrors && len(requestErrors) > 0 && len(responseErrors) > 0:
		request.HttpResponseWriter.WriteHeader(returnCode)
	default:
		request.HttpResponseWriter.WriteHeader(returnedResponse.StatusCode)
	}
	_, _ = request.HttpResponseWriter.Write(body)
}

var gorillaDropHeaders = []string{
	// Gorilla fills in the following headers, and complains if they are already present
	"Upgrade",
	"Connection",
	"Sec-Websocket-Key",
	"Sec-Websocket-Version",
	"Sec-Websocket-Protocol",
	"Sec-Websocket-Extensions",
}

func (ws *WiretapService) handleWebsocketRequest(request *model.Request) {

	configStore, _ := ws.controlsStore.Get(shared.ConfigKey)
	config := configStore.(*shared.WiretapConfiguration)

	// Get the Websocket Configuration
	websocketUrl := request.HttpRequest.URL.String()
	websocketConfig, ok := config.WebsocketConfigs[websocketUrl]
	if !ok {
		ws.config.Logger.Error(fmt.Sprintf("Unable to find websocket config for URL: %s", websocketUrl))
	}

	// There's nothing to do if we're in mock mode
	if config.MockMode {
		return
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	// Upgrade the connection from the client to open a websocket connection
	clientConn, err := upgrader.Upgrade(request.HttpResponseWriter, request.HttpRequest, nil)
	if err != nil {
		ws.config.Logger.Error("Unable to upgrade websocket connection")
		return
	}
	defer func(clientConn *websocket.Conn) {
		_ = clientConn.Close()
	}(clientConn)

	if config.Headers == nil || len(config.Headers.DropHeaders) == 0 {
		config.Headers = &shared.WiretapHeaderConfig{
			DropHeaders: []string{},
		}
	}

	// Get the updated headers and auth
	dropHeaders, injectHeaders, auth := ws.getHeadersAndAuth(config, request)

	dropHeaders = append(dropHeaders, gorillaDropHeaders...)
	dropHeaders = append(dropHeaders, websocketConfig.DropHeaders...)

	// Determine the correct websocket protocol based on redirect protocol
	var protocol string
	if config.RedirectProtocol == "https" {
		protocol = "wss"
	} else if config.RedirectProtocol == "http" {
		protocol = "ws"
	} else if config.RedirectProtocol != "wss" && config.RedirectProtocol != "ws" {
		config.Logger.Error(fmt.Sprintf("Unsupported Redirect Protocol: %s", config.RedirectProtocol))
		return
	}

	// Create a new request, which fills in the URL and other information
	newRequest := CloneExistingRequest(CloneRequest{
		Request:       request.HttpRequest,
		Protocol:      protocol,
		Host:          config.RedirectHost,
		BasePath:      config.RedirectBasePath,
		Port:          config.RedirectPort,
		DropHeaders:   dropHeaders,
		InjectHeaders: injectHeaders,
		Auth:          auth,
		Variables:     config.CompiledVariables,
	})

	// Open a new websocket connection with the server
	dialer := *websocket.DefaultDialer
	dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: !*websocketConfig.VerifyCert}
	serverConn, _, err := dialer.Dial(newRequest.URL.String(), newRequest.Header)
	if err != nil {
		ws.config.Logger.Error(fmt.Sprintf("Unable to connect to remote server; websocket connection failed: %s", err))
		return
	}
	defer func(serverConn *websocket.Conn) {
		_ = serverConn.Close()
	}(serverConn)

	// Create sentinel channels
	clientSentinel := make(chan struct{})
	serverSentinel := make(chan struct{})

	// Go-Routine for communication between Client -> Server
	go func() {
		defer close(clientSentinel)

		for {
			messageType, message, err := clientConn.ReadMessage()
			if err != nil {
				closeCode, isUnexpected := getCloseCode(err)
				logWebsocketClose(config, closeCode, isUnexpected)
				_ = clientConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				return
			}

			err = serverConn.WriteMessage(messageType, message)
			if err != nil {
				closeCode, isUnexpected := getCloseCode(err)
				logWebsocketClose(config, closeCode, isUnexpected)
				return
			}
		}
	}()

	// Go-Routine for communication between Server -> Client
	go func() {
		defer close(serverSentinel)

		for {
			messageType, message, err := serverConn.ReadMessage()
			if err != nil {
				closeCode, isUnexpected := getCloseCode(err)
				logWebsocketClose(config, closeCode, isUnexpected)
				_ = clientConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				return
			}

			err = clientConn.WriteMessage(messageType, message)
			if err != nil {
				closeCode, isUnexpected := getCloseCode(err)
				logWebsocketClose(config, closeCode, isUnexpected)
				return
			}
		}
	}()

	// Loop until at least one of our sentinel channels have been closed
	for {
		select {
		case <-clientSentinel:
			return
		case <-serverSentinel:
			return
		}
	}
}

func setCORSHeaders(headers map[string][]string) {
	headers["Access-Control-Allow-Headers"] = []string{"*"}
	headers["Access-Control-Allow-Origin"] = []string{"*"}
	headers["Access-Control-Allow-Methods"] = []string{"OPTIONS,POST,GET,DELETE,PATCH,PUT"}
}

// setStrictLocationHeader rewrites any `Location` headers to wiretap's ApiGatewayHost. Some web servers specify
// the full URL when redirecting the browser, so we need to ensure that the browser isn't redirected away from the
// wiretap Host. We achieve this by rewriting the `Location` header host and port to wiretap's host and port on all
// redirect responses.
func setStrictLocationHeader(config *shared.WiretapConfiguration, headers map[string][]string) {
	if locations, ok := headers["Location"]; ok {
		newLocations := make([]string, 0)

		apiGatewayHost := config.GetApiGatewayHost()

		for _, location := range locations {
			parsedLocation, parseErr := url.Parse(location)

			// Unable to parse the location url, let's just re-add the location to ensure that there is at least one
			// redirect target
			if parseErr != nil {
				config.Logger.Warn(fmt.Sprintf("Unable to parse `Location` header URL: %s", location))
				newLocations = append(newLocations, location)
			} else if parsedLocation.Host != "" && parsedLocation.Host != apiGatewayHost { // Check if the target location's host differs from wiretap's host
				parsedLocation.Host = apiGatewayHost

				newLocation := parsedLocation.String()
				config.Logger.Info(fmt.Sprintf("Rewrote `Location` header from %s to %s", location, newLocation))

				newLocations = append(newLocations, newLocation)
			} else { // Otherwise, we need to re-add the old location
				newLocations = append(newLocations, location)
			}

		}
		headers["Location"] = newLocations
	}

}

func getCloseCode(err error) (int, bool) {
	unexpectedClose := websocket.IsUnexpectedCloseError(err,
		websocket.CloseNormalClosure,
		websocket.CloseGoingAway,
		websocket.CloseNoStatusReceived,
		websocket.CloseAbnormalClosure,
	)

	if ce, ok := err.(*websocket.CloseError); ok {
		return ce.Code, unexpectedClose
	}
	return -1, unexpectedClose
}

func is3xxStatusCode(statusCode int) bool {
	return 300 <= statusCode && statusCode < 400
}

func logWebsocketClose(config *shared.WiretapConfiguration, closeCode int, isUnexpected bool) {
	if isUnexpected {
		config.Logger.Warn(fmt.Sprintf("Websocket closed unexepectedly with code: %d", closeCode))
	} else {
		config.Logger.Info(fmt.Sprintf("Websocket closed expectedly with code: %d", closeCode))
	}
}

func (ws *WiretapService) getHeadersAndAuth(config *shared.WiretapConfiguration, request *model.Request) ([]string, map[string]string, string) {
	var dropHeaders []string
	var injectHeaders map[string]string

	// add global headers with injection.
	if config.Headers != nil {
		dropHeaders = config.Headers.DropHeaders
		injectHeaders = config.Headers.InjectHeaders
	}

	// now add path specific headers.
	matchedPaths := configModel.FindPaths(request.HttpRequest.URL.Path, config)
	auth := ""
	if len(matchedPaths) > 0 {
		for _, path := range matchedPaths {
			auth = path.Auth
			if path.Headers != nil {
				dropHeaders = append(dropHeaders, path.Headers.DropHeaders...)
				newInjectHeaders := path.Headers.InjectHeaders
				for key := range injectHeaders {
					newInjectHeaders[key] = injectHeaders[key]
				}
				injectHeaders = newInjectHeaders
			}
			break
		}
	}

	return dropHeaders, injectHeaders, auth
}
