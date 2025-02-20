// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package cmd

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/pb33f/ranch/model"
	"github.com/pb33f/wiretap/daemon"
	"github.com/pb33f/wiretap/shared"
	staticMock "github.com/pb33f/wiretap/static-mock"
	"github.com/pterm/pterm"
)

type HandleHttpTraffic struct {
	WiretapConfig     *shared.WiretapConfiguration
	WiretapService    *daemon.WiretapService
	StaticMockService *staticMock.StaticMockService
}

func handleHttpTraffic(hht *HandleHttpTraffic) {
	wiretapConfig := hht.WiretapConfig
	wtService := hht.WiretapService
	staticMockService := hht.StaticMockService

	go func() {
		handleTraffic := func(w http.ResponseWriter, r *http.Request) {
			id, _ := uuid.NewUUID()
			// create a new request that can be passed over to the service.
			requestModel := &model.Request{
				Id:                 &id,
				HttpRequest:        r,
				HttpResponseWriter: w,
			}

			// if static-mock-dir is set, then we call the handler of staticMockService
			if len(wiretapConfig.StaticMockDir) != 0 {
				staticMockService.HandleStaticMockRequest(requestModel)
			} else { // else call the wiretap service handler
				wtService.HandleHttpRequest(requestModel)
			}
		}

		handleWebsocket := func(w http.ResponseWriter, r *http.Request) {
			id, _ := uuid.NewUUID()
			requestModel := &model.Request{
				Id:                 &id,
				HttpRequest:        r,
				HttpResponseWriter: w,
			}
			wtService.HandleWebsocketRequest(requestModel)
		}

		// create a new mux.
		mux := http.NewServeMux()

		// handle the index
		mux.HandleFunc("/", handleTraffic)

		// Handle Websockets
		for websocket := range wiretapConfig.WebsocketConfigs {
			mux.HandleFunc(websocket, handleWebsocket)
		}

		pterm.Info.Println(pterm.LightMagenta(fmt.Sprintf("API Gateway UI booting on port %s...", wiretapConfig.Port)))

		var httpErr error
		if wiretapConfig.CertificateKey != "" && wiretapConfig.Certificate != "" {
			httpErr = http.ListenAndServeTLS(fmt.Sprintf(":%s", wiretapConfig.Port),
				wiretapConfig.Certificate,
				wiretapConfig.CertificateKey,
				handlers.CompressHandler(mux))
		} else {
			httpErr = http.ListenAndServe(fmt.Sprintf(":%s", wiretapConfig.Port), handlers.CompressHandler(mux))
		}

		if httpErr != nil {
			pterm.Error.Println(httpErr)
		}
	}()
}
