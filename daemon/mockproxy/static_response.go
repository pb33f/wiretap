// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package mockproxy

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/pb33f/ranch/model"
)

type StaticResponseBroadcaster func(*http.Response)

func (h *Handler) HandleStaticResponse(
	request *model.Request,
	response *http.Response,
	broadcast StaticResponseBroadcaster,
) {
	if request == nil || request.HttpResponseWriter == nil {
		slog.Default().Warn("[wiretap] dropping static mock response; request writer is missing")
		return
	}
	if response == nil {
		slog.Default().Warn("[wiretap] static mock response is missing")
		request.HttpResponseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}

	var byteArrayBody []byte
	hasBody := response.Body != nil
	if hasBody {
		var err error
		byteArrayBody, err = io.ReadAll(response.Body)
		if closeErr := response.Body.Close(); closeErr != nil {
			slog.Default().Warn("[wiretap] static mock response body close failed", "error", closeErr)
		}
		if err != nil {
			slog.Default().Error("[wiretap] static mock response body read failed", "error", err)
			request.HttpResponseWriter.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	broadcastResponse := &http.Response{
		StatusCode: response.StatusCode,
		Header:     response.Header.Clone(),
	}
	if hasBody {
		broadcastResponse.Body = io.NopCloser(bytes.NewBuffer(byteArrayBody))
	}
	if broadcast != nil {
		go broadcast(broadcastResponse)
	}

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

	if !hasBody {
		return
	}

	_, err := request.HttpResponseWriter.Write(byteArrayBody)
	if err != nil {
		slog.Default().Error("[wiretap] static mock response body write failed", "error", err)
	}
}
