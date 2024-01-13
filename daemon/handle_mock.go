// Copyright 2023-2024 Princess Beef Heavy Industries, LLC / Dave Shanley
// https://pb33f.io
// SPDX-License-Identifier: AGPL

package daemon

import (
	"bytes"
	"fmt"
	"github.com/pb33f/ranch/model"
	configModel "github.com/pb33f/wiretap/config"
	"github.com/pb33f/wiretap/shared"
	"io"
	"net/http"
	"time"
)

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

	// build a mock based on the request.
	mock, mockStatus, mockErr := ws.mockEngine.GenerateResponse(request.HttpRequest)

	// validate http request.
	ws.ValidateRequest(request, newReq)

	// sleep for a few ms, this prevents responses from being sent out of order.
	time.Sleep(5 * time.Millisecond)

	// wiretap needs to work from anywhere, so allow everything.
	headers := make(map[string]any)
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
