// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package proxy

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
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
		ValidateRequest: func() []*shared.WiretapValidationError {
			validatedRequest = true
			return nil
		},
		ValidateResponse: func(_ *http.Response, _ []byte) []*shared.WiretapValidationError {
			validatedResponse = true
			return nil
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
		ValidateRequest: func() []*shared.WiretapValidationError {
			return nil
		},
		ValidateResponse: func(_ *http.Response, _ []byte) []*shared.WiretapValidationError {
			return []*shared.WiretapValidationError{{
				ValidationError: validationerrors.ValidationError{Message: "bad response"},
				SpecName:        "spec.yaml",
			}}
		},
		BroadcastResponseError: func(_ *http.Response, _ error) {},
	})

	rec := request.HttpResponseWriter.(*httptest.ResponseRecorder)
	assert.Equal(t, http.StatusBadGateway, rec.Code)
	assert.Equal(t, problems.JSONContentType, rec.Header().Get("Content-Type"))
	assert.Contains(t, rec.Body.String(), "Response validation failed")
}

func TestHandlerRunsValidationInlineWhenAsyncLimitIsFull(t *testing.T) {
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
	handler.Handle(request, &PreparedRequest{
		Config:      testConfig(),
		APIRequest:  httptest.NewRequest(http.MethodGet, "http://upstream.local/products", nil),
		IsHardError: false,
		CallAPI: func(_ *http.Request, _ ...*shared.WiretapConfiguration) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(bytes.NewBufferString(`{"ok":true}`)),
			}, nil
		},
		ValidateRequest: func() []*shared.WiretapValidationError {
			validatedRequest++
			return nil
		},
		ValidateResponse: func(_ *http.Response, _ []byte) []*shared.WiretapValidationError {
			validatedResponse++
			return nil
		},
		BroadcastResponseError: func(_ *http.Response, _ error) {},
	})

	assert.Equal(t, 1, validatedRequest)
	assert.Equal(t, 1, validatedResponse)
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
		ValidateRequest: func() []*shared.WiretapValidationError {
			validatedRequest = true
			return nil
		},
		ValidateResponse: func(_ *http.Response, _ []byte) []*shared.WiretapValidationError {
			return nil
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
