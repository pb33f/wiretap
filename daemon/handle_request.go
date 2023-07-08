// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package daemon

import (
	_ "embed"
	"fmt"
	"github.com/pb33f/libopenapi-validator/parameters"
	"github.com/pb33f/libopenapi-validator/requests"
	"github.com/pb33f/libopenapi-validator/responses"
	"github.com/pb33f/ranch/model"
	"github.com/pb33f/ranch/plank/utils"
	"github.com/pb33f/wiretap/shared"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"text/template"
	"time"
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
				tmpl.Execute(tmpFile, m)

				// serve it.
				http.ServeFile(request.HttpResponseWriter, request.HttpRequest, tmpFile.Name())
				return
			}

			if !localStat.IsDir() {
				http.ServeFile(request.HttpResponseWriter, request.HttpRequest, fp)
				return
			}
		}
	}
	var returnedResponse *http.Response
	var returnedError error

	// create validators.
	requestValidator := requests.NewRequestBodyValidator(ws.docModel)
	paramValidator := parameters.NewParameterValidator(ws.docModel)
	responseValidator := responses.NewResponseBodyValidator(ws.docModel)

	configStore, _ := ws.controlsStore.Get(shared.ConfigKey)

	config := configStore.(*shared.WiretapConfiguration)
	newReq := cloneRequest(CloneRequest{
		Request:     request.HttpRequest,
		Protocol:    config.RedirectProtocol,
		Host:        config.RedirectHost,
		Port:        config.RedirectPort,
		DropHeaders: config.Headers.DropHeaders,
	})

	apiRequest := cloneRequest(CloneRequest{
		Request:     request.HttpRequest,
		Protocol:    config.RedirectProtocol,
		Host:        config.RedirectHost,
		Port:        config.RedirectPort,
		DropHeaders: config.Headers.DropHeaders,
	})

	// validate the request
	go ws.validateRequest(request, newReq, requestValidator, paramValidator, responseValidator)

	// call the API being requested.
	returnedResponse, returnedError = ws.callAPI(apiRequest)

	if returnedResponse == nil && returnedError != nil {
		utils.Log.Infof("[wiretap] request %s: Failed (%d)", request.HttpRequest.URL.String(), 500)
		go ws.broadcastResponseError(request, cloneResponse(returnedResponse), returnedError)
		request.HttpResponseWriter.WriteHeader(500)
		wtError := shared.GenerateError("Unable to call API", 500, returnedError.Error(), "")
		_, _ = request.HttpResponseWriter.Write(shared.MarshalError(wtError))
		return
	} else {
		// validate response
		go ws.validateResponse(request, responseValidator, cloneResponse(returnedResponse))
	}

	// send response back to client.

	if config.GlobalAPIDelay > 0 {
		time.Sleep(time.Duration(config.GlobalAPIDelay) * time.Millisecond) // simulate a slow response.
	}
	body, _ := io.ReadAll(returnedResponse.Body)
	headers := extractHeaders(returnedResponse)

	// wiretap needs to work from anywhere, so allow everything.
	headers["Access-Control-Allow-Headers"] = "*"
	headers["Access-Control-Allow-Origin"] = "*"
	headers["Access-Control-Allow-Methods"] = "OPTIONS,POST,GET,DELETE,PATCH,PUT"

	// write headers
	for k, v := range headers {
		request.HttpResponseWriter.Header().Set(k, fmt.Sprint(v))
	}
	utils.Log.Infof("[wiretap] request %s: completed (%d)", request.HttpRequest.URL.String(), returnedResponse.StatusCode)
	// write the status code and body
	request.HttpResponseWriter.WriteHeader(returnedResponse.StatusCode)
	_, _ = request.HttpResponseWriter.Write(body)
}
