// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package mockproxy

import (
	"fmt"
	"io"
	"net/http"

	"github.com/pb33f/ranch/model"
)

type StaticResponseBroadcaster func(*http.Response)

func (h *Handler) HandleStaticResponse(
	request *model.Request,
	response *http.Response,
	broadcast StaticResponseBroadcaster,
) {
	go broadcast(response)

	for k, v := range response.Header {
		for _, j := range v {
			request.HttpResponseWriter.Header().Set(k, fmt.Sprint(j))
		}
	}

	responseCodeToReturn := http.StatusOK
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

	_, err = request.HttpResponseWriter.Write(byteArrayBody)
	if err != nil {
		panic(err)
	}
}
