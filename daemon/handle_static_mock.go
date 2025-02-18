// Copyright 2023-2024 Princess Beef Heavy Industries, LLC / Dave Shanley
// https://pb33f.io
// SPDX-License-Identifier: AGPL

package daemon

import (
	"fmt"
	"io"
	"net/http"

	"github.com/pb33f/ranch/model"
)

func (ws *WiretapService) handleStaticMockResponse(request *model.Request, response *http.Response) {
	// validate response async
	go ws.broadcastResponse(request, response)

	for k, v := range response.Header {
		for _, j := range v {
			request.HttpResponseWriter.Header().Set(k, fmt.Sprint(j))
		}
	}

	responseCodeToReturn := 200
	if response.StatusCode != 0 {
		responseCodeToReturn = response.StatusCode
	}
	request.HttpResponseWriter.WriteHeader(responseCodeToReturn)

	if response.Body == nil {
		return
	}

	byteArrayBody, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	_, errs := request.HttpResponseWriter.Write(byteArrayBody)
	if errs != nil {
		panic(errs)
	}
}
