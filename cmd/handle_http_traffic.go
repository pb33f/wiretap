// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/pb33f/ranch/model"
	"github.com/pb33f/wiretap/daemon"
	"github.com/pb33f/wiretap/shared"
	"github.com/pterm/pterm"
	"net/http"
)

func handleHttpTraffic(wiretapConfig *shared.WiretapConfiguration, wtService *daemon.WiretapService) {
	go func() {
		handleTraffic := func(w http.ResponseWriter, r *http.Request) {
			id, _ := uuid.NewUUID()
			// create a new request that can be passed over to the service.
			requestModel := &model.Request{
				Id:                 &id,
				HttpRequest:        r,
				HttpResponseWriter: w,
			}
			wtService.HandleHttpRequest(requestModel)
		}

		// create a new mux.
		mux := http.NewServeMux()

		// handle the index
		mux.HandleFunc("/", handleTraffic)

		pterm.Info.Println(pterm.LightMagenta(fmt.Sprintf("API Gateway UI booting on port %s...", wiretapConfig.Port)))
		err := http.ListenAndServe(fmt.Sprintf(":%s", wiretapConfig.Port), handlers.CompressHandler(mux))

		if err != nil {

			pterm.Fatal.Println(err)
		}
	}()
}
