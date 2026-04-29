// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package mockproxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pb33f/ranch/model"
	configModel "github.com/pb33f/wiretap/config"
	"github.com/pb33f/wiretap/daemon/problems"
	"github.com/pb33f/wiretap/shared"
)

type RequestValidator func() []*shared.WiretapValidationError
type ResponseBroadcaster func(*http.Response)
type MockGenerator func(*http.Request) ([]byte, int, error)

type PreparedRequest struct {
	Config            *shared.WiretapConfiguration
	NewReq            *http.Request
	IsHardError       bool
	ValidateRequest   RequestValidator
	GenerateMock      MockGenerator
	BroadcastResponse ResponseBroadcaster
}

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Handle(request *model.Request, prep *PreparedRequest) {
	config := prep.Config

	delay := configModel.FindPathDelay(request.HttpRequest.URL.Path, config)
	if delay > 0 {
		time.Sleep(time.Duration(delay) * time.Millisecond)
	} else if config.GlobalAPIDelay > 0 {
		time.Sleep(time.Duration(config.GlobalAPIDelay) * time.Millisecond)
	}

	var requestErrors []*shared.WiretapValidationError
	if prep.IsHardError {
		requestErrors = prep.ValidateRequest()
	} else {
		prep.ValidateRequest()
	}

	// Preserve existing ordering behavior for mock responses.
	time.Sleep(5 * time.Millisecond)

	headers := make(map[string][]string)
	shared.SetCORSHeaders(headers)
	headers["Content-Type"] = []string{"application/json"}

	for k, v := range headers {
		for _, j := range v {
			request.HttpResponseWriter.Header().Set(k, fmt.Sprint(j))
		}
	}

	if prep.IsHardError && problems.ShouldReturnValidationProblem(config, requestErrors, nil) {
		statusCode := problems.PickHardErrorStatus(true, requestErrors, nil, config, http.StatusOK)
		problem := shared.BuildValidationProblem(statusCode, request.HttpRequest.URL.Path, requestErrors, nil)
		body := shared.MarshalValidationProblem(problem)

		problems.WriteValidationProblemResponse(
			request.HttpResponseWriter,
			statusCode,
			request.HttpRequest.URL.Path,
			requestErrors,
			nil,
		)

		go prep.BroadcastResponse(&http.Response{
			StatusCode: statusCode,
			Header:     request.HttpResponseWriter.Header().Clone(),
			Body:       io.NopCloser(bytes.NewBuffer(body)),
		})
		return
	}

	mockRequest := prep.NewReq
	if mockRequest == nil {
		mockRequest = request.HttpRequest
	}
	mock, mockStatus, mockErr := prep.GenerateMock(mockRequest)
	resp := newMockResponse(mockStatus, headers, mock)

	if mockErr != nil && len(mock) == 0 {
		config.Logger.Error("[wiretap] mock mode request error", "url", mockRequest.URL.String(), "code", 404, "error", mockErr.Error())
		request.HttpResponseWriter.WriteHeader(http.StatusNotFound)
		wtError := shared.GenerateError("[mock error] unable to generate mock for request", http.StatusNotFound, mockErr.Error(), "", mock)
		_, _ = request.HttpResponseWriter.Write(shared.MarshalError(wtError))

		go prep.BroadcastResponse(resp)
		return
	}

	if mockErr != nil && len(mock) > 0 {
		config.Logger.Warn("[wiretap] mock mode request problem", "url", mockRequest.URL.String(), "code", mockStatus, "violation", mockErr.Error())
		request.HttpResponseWriter.WriteHeader(mockStatus)
		wtError := shared.GenerateError("unable to serve mocked response", mockStatus, mockErr.Error(), "", nil)
		_, _ = request.HttpResponseWriter.Write(shared.MarshalError(wtError))

		go prep.BroadcastResponse(resp)
		return
	}

	go prep.BroadcastResponse(resp)

	request.HttpResponseWriter.WriteHeader(mockStatus)
	if mock == nil {
		return
	}

	_, err := request.HttpResponseWriter.Write(mock)
	if err != nil {
		panic(err)
	}
}

func newMockResponse(status int, headers map[string][]string, body []byte) *http.Response {
	resp := &http.Response{
		StatusCode: status,
		Header:     http.Header{},
		Body:       io.NopCloser(bytes.NewBuffer(body)),
	}
	for k, v := range headers {
		for _, j := range v {
			resp.Header.Add(k, fmt.Sprint(j))
		}
	}
	return resp
}
