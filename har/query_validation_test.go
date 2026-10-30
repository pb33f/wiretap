// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package har

import (
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/testify/assert"
	"github.com/pb33f/testify/require"
	"github.com/pb33f/wiretap/shared"
)

func TestValidateHARPreservesQueryParameters(t *testing.T) {
	doc, err := libopenapi.NewDocument([]byte(`openapi: 3.1.0
info:
  title: HAR query validation
  version: "1.0"
paths:
  /pets/123:
    get:
      parameters:
        - name: tag
          in: query
          required: true
          style: form
          explode: true
          schema:
            type: array
            items:
              type: string
              enum: [cat, dog]
      responses:
        "200":
          description: ok
`))
	require.NoError(t, err)
	model, err := doc.BuildV3Model()
	require.NoError(t, err)

	for _, tc := range []struct {
		name       string
		query      string
		metadata   string
		wantErrors bool
	}{
		{"URL and metadata do not duplicate values", "?tag=cat&tag=dog", `[{"name":"tag","value":"cat"},{"name":"tag","value":"dog"}]`, false},
		{"queryString only remains supported", "", `[{"name":"tag","value":"cat"},{"name":"tag","value":"dog"}]`, false},
		{"URL only preserves repeated values", "?tag=cat&tag=dog", `[]`, false},
		{"invalid value still fails validation", "?tag=cat&tag=fish", `[]`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			content := strings.Replace(oneValidEntryHAR, "http://wiretap.local/api/pets/123", "http://wiretap.local/api/pets/123"+tc.query, 1)
			content = strings.Replace(content, `"queryString": []`, `"queryString": `+tc.metadata, 1)
			result := ValidateHARWithResult(writeTestHAR(t, content), []shared.ApiDocumentModel{{
				DocumentName:  "query-spec.yaml",
				DocumentModel: model,
			}}, &shared.WiretapConfiguration{
				HARPathAllowList: []string{"/api"},
				Logger:           slog.New(slog.NewTextHandler(io.Discard, nil)),
			})
			require.NoError(t, result.Err)
			assert.Equal(t, 2, result.MessageCount)
			if tc.wantErrors {
				assert.NotEmpty(t, result.Errors)
			} else {
				assert.Empty(t, result.Errors)
			}
		})
	}
}
