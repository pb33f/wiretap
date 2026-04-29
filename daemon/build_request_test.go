// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package daemon

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/pb33f/wiretap/shared"
	"github.com/pb33f/wiretap/transaction"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildHttpTransactionMultipartLargeForm(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	require.NoError(t, writer.WriteField("payload", strings.Repeat("a", 1<<20)))
	fileWriter, err := writer.CreateFormFile("upload", "sample.txt")
	require.NoError(t, err)
	_, err = fileWriter.Write([]byte("sample file"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	bodyBytes := append([]byte(nil), body.Bytes()...)
	req, err := http.NewRequest(http.MethodPost, "http://localhost/upload", bytes.NewReader(bodyBytes))
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	config := &shared.WiretapConfiguration{
		RedirectProtocol:   "http",
		RedirectHost:       "upstream.local",
		PathConfigurations: orderedmap.New[string, *shared.WiretapPathConfig](),
	}
	config.CompilePaths()

	id := uuid.New()
	txn := BuildHttpTransaction(HttpTransactionConfig{
		OriginalRequest:   req,
		NewRequest:        req,
		ID:                &id,
		TransactionConfig: config,
		BodyBytes:         bodyBytes,
	})

	require.NotNil(t, txn)
	require.NotNil(t, txn.Request)

	var parts []transaction.FormPart
	require.NoError(t, json.Unmarshal([]byte(txn.Request.Body), &parts))

	var sawPayload, sawFile bool
	for _, part := range parts {
		switch part.Name {
		case "payload":
			if assert.Len(t, part.Value, 1) {
				assert.Len(t, part.Value[0], 1<<20)
			}
			sawPayload = true
		case "upload":
			if assert.Len(t, part.Files, 1) {
				assert.Equal(t, "sample.txt", part.Files[0].Name)
			}
			sawFile = true
		}
	}
	assert.True(t, sawPayload, "expected large multipart field to be preserved")
	assert.True(t, sawFile, "expected multipart file metadata to be preserved")
}

func TestBuildHttpTransactionUsesPreparedDisplayURL(t *testing.T) {
	bodyBytes := []byte("payload")
	originalReq, err := http.NewRequest(
		http.MethodPost,
		"http://wiretap.local/api/items?debug=true",
		bytes.NewReader(bodyBytes),
	)
	require.NoError(t, err)

	newReq, err := http.NewRequest(
		http.MethodPost,
		"http://api.example.com/api/items?debug=true",
		bytes.NewReader(bodyBytes),
	)
	require.NoError(t, err)

	apiReq, err := http.NewRequest(
		http.MethodPost,
		"http://wrong.local/should-not-win",
		bytes.NewReader(bodyBytes),
	)
	require.NoError(t, err)

	displayURL, err := url.Parse("http://target.local/already?debug=true")
	require.NoError(t, err)

	config := &shared.WiretapConfiguration{
		RedirectProtocol:   "http",
		RedirectHost:       "fallback.local",
		PathConfigurations: orderedmap.New[string, *shared.WiretapPathConfig](),
	}
	config.PathConfigurations.Set("/api/**", &shared.WiretapPathConfig{
		Target: "rewritten.local",
		PathRewrite: map[string]string{
			"^/api/": "/rewritten/",
		},
	})
	config.CompilePaths()

	id := uuid.New()
	txn := BuildHttpTransaction(HttpTransactionConfig{
		OriginalRequest:   originalReq,
		NewRequest:        newReq,
		APIRequest:        apiReq,
		DisplayURL:        displayURL,
		ID:                &id,
		TransactionConfig: config,
		BodyBytes:         bodyBytes,
	})

	require.NotNil(t, txn)
	require.NotNil(t, txn.Request)
	assert.Equal(t, "http://target.local/already?debug=true", txn.Request.URL)
	assert.Equal(t, "target.local", txn.Request.Host)
	assert.Equal(t, "/already", txn.Request.Path)
	assert.Equal(t, "/api/items", txn.Request.OriginalPath)
	assert.Equal(t, "payload", txn.Request.Body)

	bodyAfterBuild, err := io.ReadAll(newReq.Body)
	require.NoError(t, err)
	assert.Equal(t, bodyBytes, bodyAfterBuild)
}

func TestBuildHttpTransactionAppliesHeaderConfigWithoutClone(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://wiretap.local/api/items", nil)
	require.NoError(t, err)
	req.Header.Set("X-Drop", "drop-me")
	req.Header.Set("X-Keep", "keep-me")

	config := &shared.WiretapConfiguration{
		Headers: &shared.WiretapHeaderConfig{
			DropHeaders: []string{"X-Drop"},
			InjectHeaders: map[string]string{
				"X-Injected": "${token}",
			},
		},
		Variables: map[string]string{
			"password": "secret",
			"token":    "abc123",
		},
		PathConfigurations: orderedmap.New[string, *shared.WiretapPathConfig](),
	}
	config.CompileVariables()
	config.CompilePaths()

	id := uuid.New()
	txn := BuildHttpTransaction(HttpTransactionConfig{
		OriginalRequest:   req,
		NewRequest:        req,
		ID:                &id,
		TransactionConfig: config,
		Auth:              "user:${password}",
	})

	require.NotNil(t, txn)
	require.NotNil(t, txn.Request)
	assert.NotContains(t, txn.Request.Headers, "X-Drop")
	assert.Equal(t, "keep-me", txn.Request.Headers["X-Keep"])
	assert.Equal(t, "abc123", txn.Request.Headers["X-Injected"])
	assert.Equal(t, "Basic dXNlcjpzZWNyZXQ=", txn.Request.Headers["Authorization"])
}
