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

	ws := NewWiretapService([]shared.ApiDocument{{
		DocumentName: "giftshop-openapi.yaml",
		Document:     doc,
	}}, config)
	ws.broadcastChan = bus.GetBus().GetChannelManager().CreateChannel(WiretapBroadcastChan)
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
	assert.Equal(t, problemJSONContentType, rec.Header().Get("Content-Type"))
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
