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

	store := bus.GetBus().GetStoreManager().CreateStore("prep-test-" + uuid.NewString())
	store.Put(shared.ConfigKey, config, nil)
	ws := &WiretapService{
		controlsStore: store,
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

func readBodyString(t *testing.T, body io.Reader) string {
	t.Helper()
	b, err := io.ReadAll(body)
	require.NoError(t, err)
	return string(b)
}
