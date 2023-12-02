// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package daemon

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"net/http"
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

	var requestErrors []*errors.ValidationError
	var responseErrors []*errors.ValidationError

	ws.config.Logger.Info("[wiretap] handling API request", "url", request.HttpRequest.URL.String())

	// check if we're going to fail hard on validation errors. (default is to skip this)
	if ws.config.HardErrors && !ws.config.MockMode {

		// validate the request synchronously
		requestErrors = ws.validateRequest(request, newReq)

	} else {
		// validate the request asynchronously
		if !ws.config.MockMode {
			go ws.validateRequest(request, newReq)
		}
	}

	// short-circuit if we're using mock mode, there is no API call to make.
	if ws.config.MockMode {
		ws.handleMockRequest(request, config, newReq)
		return
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
			responseErrors = ws.validateResponse(request, CloneExistingResponse(returnedResponse))
		} else {
			// validate response async
			go ws.validateResponse(request, CloneExistingResponse(returnedResponse))
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

	// write headers
	for k, v := range headers {
		request.HttpResponseWriter.Header().Set(k, fmt.Sprint(v))
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

func (ws *WiretapService) handleMockRequest(
	request *model.Request, config *shared.WiretapConfiguration, newReq *http.Request) {
	// dip out early if we're in mock mode.
	delay := configModel.FindPathDelay(request.HttpRequest.URL.Path, config)
	if delay > 0 {
		time.Sleep(time.Duration(delay) * time.Millisecond) // simulate a slow response, configured for path.
	} else {
		if config.GlobalAPIDelay > 0 {
			time.Sleep(time.Duration(config.GlobalAPIDelay) * time.Millisecond) // simulate a slow response, all paths.
		}
	}

	ws.validateRequest(request, newReq)

	// build a mock based on the request.
	mock, mockStatus, mockErr := ws.mockEngine.GenerateResponse(request.HttpRequest)

	headers := make(map[string]any)

	// wiretap needs to work from anywhere, so allow everything.
	setCORSHeaders(headers)
	headers["Content-Type"] = "application/json"

	buff := bytes.NewBuffer(mock)

	// create a simulated response to send up to the monitor UI.
	resp := &http.Response{
		StatusCode: mockStatus,
		Body:       io.NopCloser(buff),
	}
	header := http.Header{}
	resp.Header = header
	// write headers
	for k, v := range headers {
		request.HttpResponseWriter.Header().Set(k, fmt.Sprint(v))
		header.Add(k, fmt.Sprint(v))
	}

	// if there was an error building the mock, return a 404
	if mockErr != nil && len(mock) == 0 {
		config.Logger.Error("[wiretap] mock mode request error", "url", newReq.URL.String(), "code", 404, "error", mockErr.Error())
		request.HttpResponseWriter.WriteHeader(404)
		wtError := shared.GenerateError("[mock error] unable to generate mock for request", 404, mockErr.Error(), "", mock)
		_, _ = request.HttpResponseWriter.Write(shared.MarshalError(wtError))

		// validate response async
		resp.StatusCode = mockStatus
		go ws.broadcastResponse(request, resp)
		return
	}

	// if the mock exists, but there was an error, return the error
	if mockErr != nil && len(mock) > 0 {
		config.Logger.Warn("[wiretap] mock mode request problem", "url", newReq.URL.String(), "code", mockStatus, "violation", mockErr.Error())
		request.HttpResponseWriter.WriteHeader(mockStatus)
		wtError := shared.GenerateError("unable to serve mocked response", mockStatus, mockErr.Error(), "", nil)
		_, _ = request.HttpResponseWriter.Write(shared.MarshalError(wtError))

		// validate response async
		resp.StatusCode = mockStatus
		go ws.broadcastResponse(request, resp)
		return
	}

	// validate response async
	resp.StatusCode = mockStatus
	go ws.broadcastResponse(request, resp)

	// if the mock is empty
	request.HttpResponseWriter.WriteHeader(mockStatus)
	_, errs := request.HttpResponseWriter.Write(mock)
	if errs != nil {
		panic(errs)
	}
	return
}

func setCORSHeaders(headers map[string]any) {
	headers["Access-Control-Allow-Headers"] = "*"
	headers["Access-Control-Allow-Origin"] = "*"
	headers["Access-Control-Allow-Methods"] = "OPTIONS,POST,GET,DELETE,PATCH,PUT"
}
