// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package daemon

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/pb33f/ranch/bus"
	"github.com/pb33f/ranch/model"
	"github.com/pb33f/ranch/store"
	"github.com/pb33f/wiretap/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrepareRequestBuildsDisplayAndAPIRequests(t *testing.T) {
	config := &shared.WiretapConfiguration{
		RedirectProtocol: "https",
		RedirectHost:     "api.example.com",
		RedirectPort:     "8443",
		RedirectBasePath: "/v1",
		Headers: &shared.WiretapHeaderConfig{
			DropHeaders: []string{"X-Drop"},
			InjectHeaders: map[string]string{
				"X-Injected": "${token}",
			},
		},
		Variables: map[string]string{
			"token": "abc123",
		},
		PathConfigurations: orderedmap.New[string, *shared.WiretapPathConfig](),
	}
	config.CompileVariables()
	config.CompilePaths()

	eventBus := bus.NewEventBus()
	storeManager := store.NewManager(eventBus)
	configStore := storeManager.CreateStore("prep-test-" + uuid.NewString())
	configStore.Put(shared.ConfigKey, config, nil)
	ws := &WiretapService{
		controlsStore: configStore,
		config:        config,
	}

	httpReq, err := http.NewRequest(
		http.MethodPost,
		"http://wiretap.local/products?debug=true",
		strings.NewReader("payload"),
	)
	require.NoError(t, err)
	httpReq.Header.Set("X-Drop", "drop-me")
	httpReq.Header.Set("X-Keep", "keep-me")

	id := uuid.New()
	prep := ws.prepareRequest(&model.Request{
		Id:          &id,
		HttpRequest: httpReq,
	})

	require.NotNil(t, prep)
	assert.Equal(t, "https://api.example.com:8443/products?debug=true", prep.NewReq.URL.String())
	assert.Equal(t, "https://api.example.com:8443/v1/products?debug=true", prep.APIRequest.URL.String())
	assert.Equal(t, "https://api.example.com:8443/products?debug=true", prep.TxnConfig.DisplayURL.String())
	assert.Equal(t, []byte("payload"), prep.BodyBytes)
	assert.Equal(t, "payload", readBodyString(t, httpReq.Body))
	assert.Equal(t, "keep-me", prep.NewReq.Header.Get("X-Keep"))
	assert.Empty(t, prep.NewReq.Header.Get("X-Drop"))
	assert.Equal(t, "abc123", prep.NewReq.Header.Get("X-Injected"))
	assert.Equal(t, config.RedirectBasePath, prep.TxnConfig.BasePath)
	assert.Equal(t, prep.NewReq, prep.TxnConfig.NewRequest)
	assert.Equal(t, prep.BodyBytes, prep.TxnConfig.BodyBytes)
	assert.False(t, prep.UseMock)
}

func TestPrepareRequestEvaluatesControlsAgainstOriginalPath(t *testing.T) {
	config := &shared.WiretapConfiguration{
		RedirectProtocol:   "http",
		RedirectHost:       "api.example.com",
		MockModeList:       []string{"/api/**"},
		HardErrorsList:     []string{"/api/**"},
		PathConfigurations: orderedmap.New[string, *shared.WiretapPathConfig](),
	}
	config.PathConfigurations.Set("/api/**", &shared.WiretapPathConfig{
		Target: "api.example.com",
		PathRewrite: map[string]string{
			"^/api/": "/internal/",
		},
	})
	config.CompilePaths()
	config.CompileMockModeList()
	config.CompileHardErrorList()

	eventBus := bus.NewEventBus()
	storeManager := store.NewManager(eventBus)
	configStore := storeManager.CreateStore("prep-test-" + uuid.NewString())
	configStore.Put(shared.ConfigKey, config, nil)
	ws := &WiretapService{
		controlsStore: configStore,
		config:        config,
	}

	httpReq, err := http.NewRequest(http.MethodGet, "http://wiretap.local/api/products", nil)
	require.NoError(t, err)

	id := uuid.New()
	prep := ws.prepareRequest(&model.Request{
		Id:          &id,
		HttpRequest: httpReq,
	})

	require.NotNil(t, prep)
	assert.Equal(t, "/api/products", prep.ControlPath)
	assert.Equal(t, "/internal/products", prep.APIRequest.URL.Path)
	assert.True(t, prep.UseMock)
	assert.True(t, prep.IsHardError)
}

func TestPrepareRequestURLsAppliesPathRewriteToDisplayAndAPIRequests(t *testing.T) {
	config := &shared.WiretapConfiguration{
		PathConfigurations: orderedmap.New[string, *shared.WiretapPathConfig](),
	}
	config.PathConfigurations.Set("/**", &shared.WiretapPathConfig{
		Target: "target.local",
		PathRewrite: map[string]string{
			"^/api/":      "/display/",
			"^/base/api/": "/upstream/",
		},
	})
	config.CompilePaths()

	newReq, err := http.NewRequest(http.MethodGet, "http://wiretap.local/api/items?debug=true", nil)
	require.NoError(t, err)
	apiRequest, err := http.NewRequest(http.MethodGet, "http://upstream.local/base/api/items?debug=true", nil)
	require.NoError(t, err)

	displayURL := prepareRequestURLs(newReq, apiRequest, config)

	require.NotNil(t, displayURL)
	assert.Equal(t, "http://target.local/display/items?debug=true", displayURL.String())
	assert.Equal(t, "http://target.local/display/items?debug=true", newReq.URL.String())
	assert.Equal(t, "http://target.local/upstream/items?debug=true", apiRequest.URL.String())
}

func readBodyString(t *testing.T, body io.Reader) string {
	t.Helper()
	b, err := io.ReadAll(body)
	require.NoError(t, err)
	return string(b)
}
