// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package daemon

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
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
