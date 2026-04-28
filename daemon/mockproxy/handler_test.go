// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package mockproxy

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	validationerrors "github.com/pb33f/libopenapi-validator/errors"
	"github.com/pb33f/ranch/model"
	"github.com/pb33f/wiretap/daemon/problems"
	"github.com/pb33f/wiretap/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlerWritesMockResponse(t *testing.T) {
	id := uuid.New()
	request := &model.Request{
		Id:                 &id,
		HttpRequest:        httptest.NewRequest(http.MethodGet, "http://wiretap.local/products", nil),
		HttpResponseWriter: httptest.NewRecorder(),
	}

	broadcastedC := make(chan *http.Response, 1)
	NewHandler().Handle(request, &PreparedRequest{
		Config:      testConfig(),
		NewReq:      httptest.NewRequest(http.MethodGet, "http://wiretap.local/products", nil),
		IsHardError: false,
		ValidateRequest: func() []*shared.WiretapValidationError {
			return nil
		},
		GenerateMock: func(_ *http.Request) ([]byte, int, error) {
			return []byte(`{"ok":true}`), http.StatusCreated, nil
		},
		BroadcastResponse: func(resp *http.Response) {
			broadcastedC <- resp
		},
	})

	rec := request.HttpResponseWriter.(*httptest.ResponseRecorder)
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.JSONEq(t, `{"ok":true}`, rec.Body.String())
	select {
	case broadcasted := <-broadcastedC:
		assert.Equal(t, http.StatusCreated, broadcasted.StatusCode)
	case <-time.After(500 * time.Millisecond):
		require.Fail(t, "expected mock response to be broadcast")
	}
}

func TestHandlerWritesValidationProblemForHardRequestErrors(t *testing.T) {
	id := uuid.New()
	request := &model.Request{
		Id:                 &id,
		HttpRequest:        httptest.NewRequest(http.MethodPost, "http://wiretap.local/products", nil),
		HttpResponseWriter: httptest.NewRecorder(),
	}

	config := testConfig()
	config.HardErrorReturnProblem = true
	config.HardErrorCode = http.StatusBadRequest

	var generatedMock bool
	NewHandler().Handle(request, &PreparedRequest{
		Config:      config,
		NewReq:      httptest.NewRequest(http.MethodPost, "http://wiretap.local/products", nil),
		IsHardError: true,
		ValidateRequest: func() []*shared.WiretapValidationError {
			return []*shared.WiretapValidationError{{
				ValidationError: validationerrors.ValidationError{Message: "bad request"},
				SpecName:        "spec.yaml",
			}}
		},
		GenerateMock: func(_ *http.Request) ([]byte, int, error) {
			generatedMock = true
			return []byte(`{"ok":true}`), http.StatusOK, nil
		},
		BroadcastResponse: func(_ *http.Response) {},
	})

	rec := request.HttpResponseWriter.(*httptest.ResponseRecorder)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, problems.JSONContentType, rec.Header().Get("Content-Type"))
	assert.False(t, generatedMock, "hard-error problem path must short-circuit mock generation")
	assert.Contains(t, rec.Body.String(), "Request validation failed")
}

func testConfig() *shared.WiretapConfiguration {
	return &shared.WiretapConfiguration{
		HardErrorCode:       http.StatusBadRequest,
		HardErrorReturnCode: http.StatusBadGateway,
		Logger:              slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}
