// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package daemon

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/pb33f/ranch/bus"
	"github.com/pb33f/ranch/model"
	"github.com/pb33f/ranch/store"
	"github.com/pb33f/wiretap/daemon/proxy"
	"github.com/pb33f/wiretap/shared"
	"github.com/pb33f/wiretap/transaction"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleHttpRequestValidatesResponseAgainstPreparedRequestPath(t *testing.T) {
	doc, err := libopenapi.NewDocument([]byte(`openapi: 3.1.0
info:
  title: rewritten response validation
  version: "1.0"
paths:
  /products:
    get:
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema:
                type: object
                required:
                  - name
                properties:
                  name:
                    type: string
`))
	require.NoError(t, err)

	config := &shared.WiretapConfiguration{
		RedirectProtocol:    "http",
		RedirectHost:        "upstream.local",
		HardErrors:          true,
		HardErrorCode:       http.StatusBadRequest,
		HardErrorReturnCode: http.StatusBadGateway,
		ReportFile:          t.TempDir() + "/violations.jsonl",
		Logger:              slog.New(slog.NewTextHandler(io.Discard, nil)),
		PathConfigurations:  orderedmap.New[string, *shared.WiretapPathConfig](),
	}
	config.PathConfigurations.Set("/external/**", &shared.WiretapPathConfig{
		Target: "upstream.local",
		PathRewrite: map[string]string{
			"^/external": "",
		},
	})
	config.CompilePaths()

	eventBus := bus.NewEventBus()
	storeManager := store.NewManager(eventBus)
	ws := NewWiretapService([]shared.ApiDocument{{
		DocumentName: "rewritten.yaml",
		Document:     doc,
	}}, config, storeManager)
	ws.controlsStore.Put(shared.ConfigKey, config, nil)

	broadcaster := &recordingBroadcaster{}
	ws.broadcaster.Set(broadcaster)
	ws.proxy = proxy.NewHandler(roundTripFunc(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, "/products", req.URL.Path)
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"wrong":true}`)),
		}, nil
	}))

	id := uuid.New()
	request := &model.Request{
		Id:                 &id,
		HttpRequest:        httptest.NewRequest(http.MethodGet, "http://wiretap.local/external/products", nil),
		HttpResponseWriter: httptest.NewRecorder(),
	}

	ws.handleHttpRequest(request)

	errs := broadcaster.responseValidationErrors()
	require.NotEmpty(t, errs)
	assert.False(t, errs[0].IsPathMissingError())
	assert.Contains(t, errs[0].Message, "response body")
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type recordingBroadcaster struct {
	mu                               sync.Mutex
	capturedResponseValidationErrors []*shared.WiretapValidationError
}

func (r *recordingBroadcaster) RequestValidationErrors(
	*model.Request,
	[]*shared.WiretapValidationError,
	*transaction.HttpTransaction,
) {
}

func (r *recordingBroadcaster) Request(*model.Request, *transaction.HttpTransaction) {}

func (r *recordingBroadcaster) Response(*model.Request, *transaction.HttpTransaction) {}

func (r *recordingBroadcaster) ResponseError(*model.Request, *transaction.HttpTransaction, error) {}

func (r *recordingBroadcaster) ResponseValidationErrors(
	_ *model.Request,
	_ *transaction.HttpTransaction,
	errs []*shared.WiretapValidationError,
) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.capturedResponseValidationErrors = errs
}

func (r *recordingBroadcaster) responseValidationErrors() []*shared.WiretapValidationError {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]*shared.WiretapValidationError(nil), r.capturedResponseValidationErrors...)
}
