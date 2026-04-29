// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package mockproxy

import (
	"bytes"
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
	var byteArrayBody []byte
	if response.Body != nil {
		var err error
		byteArrayBody, err = io.ReadAll(response.Body)
		if err != nil {
			panic(err)
		}
		_ = response.Body.Close()
	}

	broadcastResponse := &http.Response{
		StatusCode: response.StatusCode,
		Header:     response.Header.Clone(),
	}
	if response.Body != nil {
		broadcastResponse.Body = io.NopCloser(bytes.NewBuffer(byteArrayBody))
	}
	go broadcast(broadcastResponse)

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

	_, err := request.HttpResponseWriter.Write(byteArrayBody)
	if err != nil {
		panic(err)
	}
}
