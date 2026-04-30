// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package proxy

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	validationerrors "github.com/pb33f/libopenapi-validator/errors"
	"github.com/pb33f/ranch/model"
	"github.com/pb33f/wiretap/daemon/problems"
	"github.com/pb33f/wiretap/shared"
	"github.com/stretchr/testify/assert"
)

func TestHandlerWritesUpstreamResponse(t *testing.T) {
	id := uuid.New()
	request := &model.Request{
		Id: &id,
		HttpRequest: httptest.NewRequest(
			http.MethodGet,
			"http://wiretap.local/products",
			nil,
		),
		HttpResponseWriter: httptest.NewRecorder(),
	}

	var validatedRequest, validatedResponse bool
	NewHandler().Handle(request, &PreparedRequest{
		Config:      testConfig(),
		APIRequest:  httptest.NewRequest(http.MethodGet, "http://upstream.local/products", nil),
		IsHardError: true,
		CallAPI: func(_ *http.Request, _ ...*shared.WiretapConfiguration) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusAccepted,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(bytes.NewBufferString(`{"ok":true}`)),
			}, nil
		},
		Validator: testValidator{
			validateRequest: func() []*shared.WiretapValidationError {
				validatedRequest = true
				return nil
			},
			validateResponse: func(_ *http.Response, _ []byte) []*shared.WiretapValidationError {
				validatedResponse = true
				return nil
			},
		},
		BroadcastResponseError: func(_ *http.Response, _ error) {},
	})

	rec := request.HttpResponseWriter.(*httptest.ResponseRecorder)
	assert.Equal(t, http.StatusAccepted, rec.Code)
	assert.JSONEq(t, `{"ok":true}`, rec.Body.String())
	assert.True(t, validatedRequest)
	assert.True(t, validatedResponse)
}

func TestHandlerWritesValidationProblemForHardResponseErrors(t *testing.T) {
	id := uuid.New()
	request := &model.Request{
		Id: &id,
		HttpRequest: httptest.NewRequest(
			http.MethodGet,
			"http://wiretap.local/products",
			nil,
		),
		HttpResponseWriter: httptest.NewRecorder(),
	}

	config := testConfig()
	config.HardErrorReturnProblem = true
	config.HardErrorReturnCode = http.StatusBadGateway

	NewHandler().Handle(request, &PreparedRequest{
		Config:      config,
		APIRequest:  httptest.NewRequest(http.MethodGet, "http://upstream.local/products", nil),
		IsHardError: true,
		CallAPI: func(_ *http.Request, _ ...*shared.WiretapConfiguration) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(bytes.NewBufferString(`{"ok":true}`)),
			}, nil
		},
		Validator: testValidator{
			validateRequest: func() []*shared.WiretapValidationError {
				return nil
			},
			validateResponse: func(_ *http.Response, _ []byte) []*shared.WiretapValidationError {
				return []*shared.WiretapValidationError{{
					ValidationError: validationerrors.ValidationError{Message: "bad response"},
					SpecName:        "spec.yaml",
				}}
			},
		},
		BroadcastResponseError: func(_ *http.Response, _ error) {},
	})

	rec := request.HttpResponseWriter.(*httptest.ResponseRecorder)
	assert.Equal(t, http.StatusBadGateway, rec.Code)
	assert.Equal(t, problems.JSONContentType, rec.Header().Get("Content-Type"))
	assert.Contains(t, rec.Body.String(), "Response validation failed")
}

func TestHandlerDropsSoftValidationWhenAsyncLimitIsFull(t *testing.T) {
	id := uuid.New()
	request := &model.Request{
		Id: &id,
		HttpRequest: httptest.NewRequest(
			http.MethodGet,
			"http://wiretap.local/products",
			nil,
		),
		HttpResponseWriter: httptest.NewRecorder(),
	}

	handler := NewHandler()
	for i := 0; i < cap(handler.validationSem); i++ {
		handler.validationSem <- struct{}{}
	}
	defer func() {
		for i := 0; i < cap(handler.validationSem); i++ {
			<-handler.validationSem
		}
	}()

	var validatedRequest, validatedResponse int
	var logs bytes.Buffer
	config := testConfig()
	config.Logger = slog.New(slog.NewTextHandler(&logs, nil))

	handler.Handle(request, &PreparedRequest{
		Config:      config,
		APIRequest:  httptest.NewRequest(http.MethodGet, "http://upstream.local/products", nil),
		IsHardError: false,
		CallAPI: func(_ *http.Request, _ ...*shared.WiretapConfiguration) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(bytes.NewBufferString(`{"ok":true}`)),
			}, nil
		},
		Validator: testValidator{
			validateRequest: func() []*shared.WiretapValidationError {
				validatedRequest++
				return nil
			},
			validateResponse: func(_ *http.Response, _ []byte) []*shared.WiretapValidationError {
				validatedResponse++
				return nil
			},
		},
		BroadcastResponseError: func(_ *http.Response, _ error) {},
	})

	rec := request.HttpResponseWriter.(*httptest.ResponseRecorder)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 0, validatedRequest)
	assert.Equal(t, 0, validatedResponse)
	assert.True(t, strings.Contains(logs.String(), "phase=request"), logs.String())
	assert.True(t, strings.Contains(logs.String(), "phase=response"), logs.String())
}

func TestHandlerIgnoreValidationUsesOriginalControlPath(t *testing.T) {
	id := uuid.New()
	request := &model.Request{
		Id: &id,
		HttpRequest: httptest.NewRequest(
			http.MethodGet,
			"http://wiretap.local/api/products",
			nil,
		),
		HttpResponseWriter: httptest.NewRecorder(),
	}

	config := testConfig()
	config.IgnoreValidation = []string{"/api/**"}
	config.CompileIgnoreValidations()

	var validatedRequest bool
	NewHandler().Handle(request, &PreparedRequest{
		Config:      config,
		ControlPath: "/api/products",
		APIRequest:  httptest.NewRequest(http.MethodGet, "http://upstream.local/internal/products", nil),
		IsHardError: true,
		CallAPI: func(_ *http.Request, _ ...*shared.WiretapConfiguration) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(bytes.NewBufferString(`{"ok":true}`)),
			}, nil
		},
		Validator: testValidator{
			validateRequest: func() []*shared.WiretapValidationError {
				validatedRequest = true
				return nil
			},
			validateResponse: func(_ *http.Response, _ []byte) []*shared.WiretapValidationError {
				return nil
			},
		},
		BroadcastResponseError: func(_ *http.Response, _ error) {},
	})

	assert.False(t, validatedRequest)
}

func testConfig() *shared.WiretapConfiguration {
	return &shared.WiretapConfiguration{
		HardErrorCode:       http.StatusBadRequest,
		HardErrorReturnCode: http.StatusBadGateway,
		Logger:              slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}

type testValidator struct {
	validateRequest  func() []*shared.WiretapValidationError
	validateResponse func(*http.Response, []byte) []*shared.WiretapValidationError
}

func (v testValidator) ValidateRequest() []*shared.WiretapValidationError {
	if v.validateRequest == nil {
		return nil
	}
	return v.validateRequest()
}

func (v testValidator) ValidateResponse(response *http.Response, body []byte) []*shared.WiretapValidationError {
	if v.validateResponse == nil {
		return nil
	}
	return v.validateResponse(response, body)
}
