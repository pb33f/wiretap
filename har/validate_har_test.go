// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package har

import (
	"io"
	"log/slog"
	"testing"

	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/wiretap/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateHARStreamsFromPath(t *testing.T) {
	harPath := writeTestHAR(t, oneValidEntryHAR)
	docModel := buildHARValidationDoc(t)
	config := &shared.WiretapConfiguration{
		HARPathAllowList: []string{"/api"},
		Logger:           slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	errs := ValidateHAR(harPath, []shared.ApiDocumentModel{{
		DocumentName:  "test-spec.yaml",
		DocumentModel: docModel,
	}}, config)

	assert.Empty(t, errs)
}

func TestValidateHARMalformedFileReturnsError(t *testing.T) {
	cases := map[string]string{
		"invalid json": `not valid json`,
		"truncated": `{
  "log": {
    "version": "1.2",
    "entries": [
`,
	}

	for name, content := range cases {
		t.Run(name, func(t *testing.T) {
			harPath := writeTestHAR(t, content)
			docModel := buildHARValidationDoc(t)
			config := &shared.WiretapConfiguration{
				HARPathAllowList: []string{"/api"},
				Logger:           slog.New(slog.NewTextHandler(io.Discard, nil)),
			}

			result := ValidateHARWithResult(harPath, []shared.ApiDocumentModel{{
				DocumentName:  "test-spec.yaml",
				DocumentModel: docModel,
			}}, config)

			require.Error(t, result.Err)
			assert.Empty(t, result.Errors)
		})
	}
}

func buildHARValidationDoc(t *testing.T) *libopenapi.DocumentModel[v3.Document] {
	t.Helper()

	doc, err := libopenapi.NewDocument([]byte(`openapi: 3.1.0
info:
  title: HAR validation test
  version: "1.0"
paths:
  /pets/{petId}:
    get:
      responses:
        "200":
          description: ok
`))
	require.NoError(t, err)

	model, err := doc.BuildV3Model()
	require.NoError(t, err)
	return model
}

const oneValidEntryHAR = `{
  "log": {
    "version": "1.2",
    "creator": {
      "name": "wiretap-test",
      "version": "1.0"
    },
    "entries": [
      {
        "startedDateTime": "2026-04-27T00:00:00Z",
        "time": 1,
        "request": {
          "method": "GET",
          "url": "http://wiretap.local/api/pets/123",
          "httpVersion": "HTTP/1.1",
          "cookies": [],
          "headers": [],
          "queryString": [],
          "headersSize": -1,
          "bodySize": 0
        },
        "response": {
          "status": 200,
          "statusText": "OK",
          "httpVersion": "HTTP/1.1",
          "cookies": [],
          "headers": [],
          "content": {
            "size": 0,
            "mimeType": "application/json",
            "text": "{}"
          },
          "headersSize": -1,
          "bodySize": 2
        },
        "cache": {},
        "timings": {
          "send": 0,
          "wait": 1,
          "receive": 0
        }
      }
    ]
  }
}`
