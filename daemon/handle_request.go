// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package daemon

import (
	"fmt"
	"github.com/pb33f/libopenapi-validator/parameters"
	"github.com/pb33f/libopenapi-validator/requests"
	"github.com/pb33f/libopenapi-validator/responses"
	"github.com/pb33f/ranch/model"
	"github.com/pb33f/ranch/plank/utils"
	"github.com/pb33f/wiretap/shared"
	"io"
	"net/http"
	"time"
)

func (ws *WiretapService) handleHttpRequest(request *model.Request) {

	lowResponseChan := make(chan *http.Response)
	lowErrorChan := make(chan error)
	var returnedResponse *http.Response
	var returnedError error

	// create validators.
	requestValidator := requests.NewRequestBodyValidator(ws.docModel)
	paramValidator := parameters.NewParameterValidator(ws.docModel)
	responseValidator := responses.NewResponseBodyValidator(ws.docModel)

	// validate the request
	go ws.validateRequest(request, requestValidator, paramValidator, responseValidator)

	// call the API being requested.
	go ws.callAPI(request.HttpRequest, lowResponseChan, lowErrorChan)

doneWaitingForResponse:
	for {
		select {
		case resp, ok := <-lowResponseChan:
			if ok {
				returnedResponse = resp
			}
			break doneWaitingForResponse
		case err := <-lowErrorChan:
			returnedError = err
			break doneWaitingForResponse
		}
	}

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
	//go func() {
	config := ws.controlsStore.GetValue(shared.ConfigKey).(*shared.WiretapConfiguration)
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
