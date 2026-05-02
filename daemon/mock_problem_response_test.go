// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package daemon

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/pb33f/ranch/bus"
	"github.com/pb33f/ranch/model"
	"github.com/pb33f/ranch/store"
	"github.com/pb33f/wiretap/daemon/problems"
	"github.com/pb33f/wiretap/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockModeWiretapService(t *testing.T, config *shared.WiretapConfiguration) *WiretapService {
	t.Helper()

	if config.PathConfigurations == nil {
		config.PathConfigurations = orderedmap.New[string, *shared.WiretapPathConfig]()
	}
	config.CompilePaths()
	if config.ReportFile == "" {
		config.ReportFile = t.TempDir() + "/violations.ndjson"
	}

	if config.Logger == nil {
		config.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	spec, err := os.ReadFile("../testdata/giftshop-openapi.yaml")
	require.NoError(t, err)

	doc, err := libopenapi.NewDocument(spec)
	require.NoError(t, err)

	eventBus := bus.NewEventBus()
	storeManager := store.NewManager(eventBus)
	ws := NewWiretapService([]shared.ApiDocument{{
		DocumentName: "giftshop-openapi.yaml",
		Document:     doc,
	}}, config, storeManager)
	ws.setBroadcastChannel(eventBus.GetChannelManager().CreateChannel(WiretapBroadcastChan))
	ws.controlsStore.Put(shared.ConfigKey, config, nil)
	return ws
}

func newInvalidGiftshopCreateProductRequest(t *testing.T) (*model.Request, *httptest.ResponseRecorder) {
	t.Helper()

	payload := []byte(`{"price":400.23,"shortCode":"pb0001"}`)
	httpReq, err := http.NewRequest(
		http.MethodPost,
		"https://api.pb33f.io/wiretap/giftshop/products",
		bytes.NewBuffer(payload),
	)
	require.NoError(t, err)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", "doesnotmatter")

	rec := httptest.NewRecorder()
	id := uuid.New()
	return &model.Request{
		Id:                 &id,
		HttpRequest:        httpReq,
		HttpResponseWriter: rec,
	}, rec
}

func TestHandleHttpRequest_MockModeHardErrorReturnProblem(t *testing.T) {
	config := &shared.WiretapConfiguration{
		MockMode:               true,
		HardErrors:             true,
		HardErrorCode:          http.StatusBadRequest,
		HardErrorReturnCode:    http.StatusBadGateway,
		HardErrorReturnProblem: true,
	}
	ws := newMockModeWiretapService(t, config)
	request, rec := newInvalidGiftshopCreateProductRequest(t)

	ws.handleHttpRequest(request)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, problems.JSONContentType, rec.Header().Get("Content-Type"))
	assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &decoded))

	assert.Equal(t, shared.ValidationProblemType, decoded["type"])
	assert.Equal(t, "Request validation failed", decoded["title"])
	assert.Len(t, decoded["requestErrors"], 1)
}

func TestHandleHttpRequest_MockModeHardErrorKeepsLegacyJSONWhenProblemDisabled(t *testing.T) {
	config := &shared.WiretapConfiguration{
		MockMode:            true,
		HardErrors:          true,
		HardErrorCode:       http.StatusBadRequest,
		HardErrorReturnCode: http.StatusBadGateway,
	}
	ws := newMockModeWiretapService(t, config)
	request, rec := newInvalidGiftshopCreateProductRequest(t)

	ws.handleHttpRequest(request)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &decoded))

	assert.Equal(t, "unable to serve mocked response", decoded["title"])
	_, hasPayload := decoded["payload"]
	assert.False(t, hasPayload)
}

func TestHandleHttpRequest_MockModeUsesRouteAdjustedRequest(t *testing.T) {
	spec := []byte(`openapi: 3.1.0
info:
  title: scoped mock
  version: "1.0"
paths:
  "/health":
    servers:
      - url: https://api.example.com/accounts
    get:
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema:
                type: object
                properties:
                  status:
                    type: string
                    example: ok
`)
	doc, err := libopenapi.NewDocument(spec)
	require.NoError(t, err)

	config := &shared.WiretapConfiguration{
		MockMode:           true,
		ReportFile:         t.TempDir() + "/violations.ndjson",
		Logger:             slog.New(slog.NewTextHandler(io.Discard, nil)),
		PathConfigurations: orderedmap.New[string, *shared.WiretapPathConfig](),
	}
	config.CompilePaths()

	eventBus := bus.NewEventBus()
	storeManager := store.NewManager(eventBus)
	ws := NewWiretapService([]shared.ApiDocument{{
		DocumentName: "scoped.yaml",
		Document:     doc,
	}}, config, storeManager)
	ws.setBroadcastChannel(eventBus.GetChannelManager().CreateChannel(WiretapBroadcastChan))
	ws.controlsStore.Put(shared.ConfigKey, config, nil)

	id := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "http://wiretap.local/accounts/health", nil)
	rec := httptest.NewRecorder()
	ws.handleHttpRequest(&model.Request{
		Id:                 &id,
		HttpRequest:        req,
		HttpResponseWriter: rec,
	})

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotContains(t, rec.Body.String(), "Path / operation not found")
}

func TestHandleHttpRequest_MockModeRejectsAnotherOperationsServerBase(t *testing.T) {
	spec := []byte(`openapi: 3.1.0
info:
  title: split mock
  version: "1.0"
paths:
  "/health":
    get:
      servers:
        - url: https://api.example.com/get
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema:
                type: object
                properties:
                  status:
                    type: string
                    example: get
    post:
      servers:
        - url: https://api.example.com/post
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema:
                type: object
                properties:
                  status:
                    type: string
                    example: post
`)
	doc, err := libopenapi.NewDocument(spec)
	require.NoError(t, err)

	config := &shared.WiretapConfiguration{
		MockMode:           true,
		ReportFile:         t.TempDir() + "/violations.ndjson",
		Logger:             slog.New(slog.NewTextHandler(io.Discard, nil)),
		PathConfigurations: orderedmap.New[string, *shared.WiretapPathConfig](),
	}
	config.CompilePaths()

	eventBus := bus.NewEventBus()
	storeManager := store.NewManager(eventBus)
	ws := NewWiretapService([]shared.ApiDocument{{
		DocumentName: "split.yaml",
		Document:     doc,
	}}, config, storeManager)
	ws.setBroadcastChannel(eventBus.GetChannelManager().CreateChannel(WiretapBroadcastChan))
	ws.controlsStore.Put(shared.ConfigKey, config, nil)

	id := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "http://wiretap.local/post/health", nil)
	rec := httptest.NewRecorder()
	ws.handleHttpRequest(&model.Request{
		Id:                 &id,
		HttpRequest:        req,
		HttpResponseWriter: rec,
	})

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "unable to serve mocked response")
}
