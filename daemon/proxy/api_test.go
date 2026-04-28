// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/pb33f/wiretap/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCallAPIRewritesReferer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "http://old.example/products", r.Header.Get("X-Original-Referer"))
		assert.Equal(t, "http://wiretap.local:9090/v1/products?debug=true", r.Header.Get("Referer"))
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	req, err := http.NewRequest(http.MethodGet, server.URL+"/products?debug=true", nil)
	require.NoError(t, err)
	req.Header.Set("Referer", "http://old.example/products")

	resp, err := NewHandler().callAPI(req, testWiretapConfig())
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
}

func TestCallAPIPreservesRedirectCookieOnFinalResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/redirect":
			http.SetCookie(w, &http.Cookie{Name: "session", Value: "abc123"})
			http.Redirect(w, r, "/final", http.StatusFound)
		case "/final":
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	req, err := http.NewRequest(http.MethodGet, server.URL+"/redirect", nil)
	require.NoError(t, err)

	resp, err := NewHandler().callAPI(req, testWiretapConfig())
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Set-Cookie"), "session=abc123")
}

func testWiretapConfig() *shared.WiretapConfiguration {
	config := &shared.WiretapConfiguration{
		RedirectProtocol:   "http",
		RedirectHost:       "wiretap.local",
		RedirectPort:       "9090",
		RedirectBasePath:   "/v1",
		PathConfigurations: orderedmap.New[string, *shared.WiretapPathConfig](),
	}
	config.CompilePaths()
	return config
}
