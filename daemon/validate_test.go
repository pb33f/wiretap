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
	"github.com/pb33f/wiretap/specs"
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

func TestValidateRequestAttachesSpecConflict(t *testing.T) {
	docs := []shared.ApiDocument{
		buildDaemonSpec(t, "users.yaml", "/users/{id}", "id"),
		buildDaemonSpec(t, "accounts.yaml", "/users/{name}", "name"),
	}
	report := specs.Analyze(docs)
	require.NotEmpty(t, report.Conflicts)

	config := &shared.WiretapConfiguration{
		ReportFile: t.TempDir() + "/violations.jsonl",
		Logger:     slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	eventBus := bus.NewEventBus()
	storeManager := store.NewManager(eventBus)
	ws := NewWiretapService(docs, config, storeManager, report)
	ws.setBroadcastChannel(eventBus.GetChannelManager().CreateChannel(WiretapBroadcastChan))
	ws.controlsStore.Put(shared.ConfigKey, config, nil)

	id := uuid.New()
	httpReq := httptest.NewRequest(http.MethodGet, "http://wiretap.local/users/123", nil)
	modelRequest := &model.Request{
		Id:          &id,
		HttpRequest: httpReq,
	}

	ws.ValidateRequest(modelRequest, httpReq)

	stored, ok := ws.transactionStore.Get(id.String())
	require.True(t, ok)
	txn, ok := stored.(*transaction.HttpTransaction)
	require.True(t, ok)
	require.NotNil(t, txn.SpecConflict)
	assert.Equal(t, "users.yaml", txn.SpecConflict.MatchedSpec)
	assert.Equal(t, "/users/{id}", txn.SpecConflict.Path)
	assert.Equal(t, "/users/{id}", txn.SpecConflict.RoutePath)
	assert.Equal(t, []string{"accounts.yaml"}, txn.SpecConflict.ConflictSpecs)
}

func TestValidateRequestFiltersSpecConflictByCurrentPath(t *testing.T) {
	docs := []shared.ApiDocument{
		buildDaemonLiteralSpec(t, "profile.yaml", "/users/me"),
		buildDaemonSpec(t, "users.yaml", "/users/{id}", "id"),
	}
	report := specs.Analyze(docs)
	require.NotEmpty(t, report.Conflicts)

	config := &shared.WiretapConfiguration{
		ReportFile: t.TempDir() + "/violations.jsonl",
		Logger:     slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	eventBus := bus.NewEventBus()
	storeManager := store.NewManager(eventBus)
	ws := NewWiretapService(docs, config, storeManager, report)
	ws.setBroadcastChannel(eventBus.GetChannelManager().CreateChannel(WiretapBroadcastChan))
	ws.controlsStore.Put(shared.ConfigKey, config, nil)

	id := uuid.New()
	httpReq := httptest.NewRequest(http.MethodGet, "http://wiretap.local/users/123", nil)
	modelRequest := &model.Request{
		Id:          &id,
		HttpRequest: httpReq,
	}

	ws.ValidateRequest(modelRequest, httpReq)

	stored, ok := ws.transactionStore.Get(id.String())
	require.True(t, ok)
	txn, ok := stored.(*transaction.HttpTransaction)
	require.True(t, ok)
	assert.Nil(t, txn.SpecConflict)

	id = uuid.New()
	httpReq = httptest.NewRequest(http.MethodGet, "http://wiretap.local/users/me", nil)
	modelRequest = &model.Request{
		Id:          &id,
		HttpRequest: httpReq,
	}

	ws.ValidateRequest(modelRequest, httpReq)

	stored, ok = ws.transactionStore.Get(id.String())
	require.True(t, ok)
	txn, ok = stored.(*transaction.HttpTransaction)
	require.True(t, ok)
	require.NotNil(t, txn.SpecConflict)
	assert.Equal(t, "profile.yaml", txn.SpecConflict.MatchedSpec)
	assert.Equal(t, "/users/me", txn.SpecConflict.Path)
	assert.Equal(t, "/users/me", txn.SpecConflict.RoutePath)
	assert.Equal(t, []string{"users.yaml"}, txn.SpecConflict.ConflictSpecs)
}

func buildDaemonSpec(t *testing.T, name, path, paramName string) shared.ApiDocument {
	t.Helper()
	doc, err := libopenapi.NewDocument([]byte(`openapi: 3.1.0
info:
  title: route conflict test
  version: "1.0"
paths:
  "` + path + `":
    get:
      parameters:
        - name: ` + paramName + `
          in: path
          required: true
          schema:
            type: string
      responses:
        "200":
          description: ok
`))
	require.NoError(t, err)
	model, err := doc.BuildV3Model()
	require.NoError(t, err)
	return shared.ApiDocument{
		DocumentName:  name,
		Document:      doc,
		DocumentModel: model,
	}
}

func buildDaemonLiteralSpec(t *testing.T, name, path string) shared.ApiDocument {
	t.Helper()
	doc, err := libopenapi.NewDocument([]byte(`openapi: 3.1.0
info:
  title: route conflict test
  version: "1.0"
paths:
  "` + path + `":
    get:
      responses:
        "200":
          description: ok
`))
	require.NoError(t, err)
	model, err := doc.BuildV3Model()
	require.NoError(t, err)
	return shared.ApiDocument{
		DocumentName:  name,
		Document:      doc,
		DocumentModel: model,
	}
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
