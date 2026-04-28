// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package daemon

import (
	"crypto/tls"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/pb33f/ranch/model"
	"github.com/pb33f/wiretap/shared"
)

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

	clientSentinel := make(chan struct{})
	serverSentinel := make(chan struct{})

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

	for {
		select {
		case <-clientSentinel:
			return
		case <-serverSentinel:
			return
		}
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

func logWebsocketClose(config *shared.WiretapConfiguration, closeCode int, isUnexpected bool) {
	if isUnexpected {
		config.Logger.Warn(fmt.Sprintf("Websocket closed unexepectedly with code: %d", closeCode))
	} else {
		config.Logger.Info(fmt.Sprintf("Websocket closed expectedly with code: %d", closeCode))
	}
}
