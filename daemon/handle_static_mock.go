// Copyright 2023-2024 Princess Beef Heavy Industries, LLC / Dave Shanley
// https://pb33f.io
// SPDX-License-Identifier: AGPL

package daemon

import (
	"io"
	"net/http"

	"github.com/pb33f/ranch/model"
)

func (ws *WiretapService) handleStaticMockResponse(request *model.Request, response *http.Response) {
	// validate response async
	go ws.broadcastResponse(request, response)

	// if the mock is empty
	request.HttpResponseWriter.WriteHeader(response.StatusCode)

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
