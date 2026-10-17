// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package mock

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/testify/require"
)

func TestGenerateResponseConcurrentUUIDs(t *testing.T) {
	const spec = `openapi: 3.0.3
info:
  title: Concurrent UUID mocks
  version: 1.0.0
components:
  securitySchemes:
    token:
      type: http
      scheme: bearer
paths:
  /ids:
    get:
      security:
        - token: []
      responses:
        '200': &response
          description: Generated ID
          content:
            application/json:
              schema:
                type: string
                format: uuid
        '401': *response
`
	doc, err := libopenapi.NewDocument([]byte(spec))
	require.NoError(t, err)
	model, err := doc.BuildV3Model()
	require.NoError(t, err)
	engine := NewMockEngine(&model.Model, false, true)

	checkResponse := func(t *testing.T, authorized bool) {
		t.Helper()
		request := httptest.NewRequest(http.MethodGet, "http://example.com/ids", nil)
		request.Header.Set("Content-Type", "application/json")
		if authorized {
			request.Header.Set("Authorization", "Bearer test-token")
		}
		body, status, err := engine.GenerateResponse(request)
		if authorized {
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, status)
		} else {
			require.Error(t, err)
			require.Equal(t, http.StatusUnauthorized, status)
		}
		var id string
		require.NoError(t, json.Unmarshal(body, &id))
		_, err = uuid.Parse(id)
		require.NoError(t, err)
	}

	// Warm validation and schema caches before exercising the shared renderer.
	checkResponse(t, true)
	checkResponse(t, false)
	for worker := range 8 {
		t.Run(fmt.Sprintf("worker-%d", worker), func(t *testing.T) {
			t.Parallel()
			for request := range 32 {
				checkResponse(t, (worker+request)%2 == 0)
			}
		})
	}
}
