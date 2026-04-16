// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package daemon

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/pb33f/libopenapi-validator/errors"
	"github.com/pb33f/wiretap/shared"
	"github.com/stretchr/testify/assert"
)

func buildSampleErrors(messages ...string) []*shared.WiretapValidationError {
	out := make([]*shared.WiretapValidationError, 0, len(messages))
	for _, m := range messages {
		out = append(out, &shared.WiretapValidationError{
			ValidationError: errors.ValidationError{Message: m},
			SpecName:        "spec.yaml",
		})
	}
	return out
}

func TestWriteValidationProblemResponse_StatusAndContentType(t *testing.T) {
	rec := httptest.NewRecorder()

	writeValidationProblemResponse(rec, 400, "/foo",
		buildSampleErrors("bad request"), nil)

	assert.Equal(t, 400, rec.Code)
	assert.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))
}

func TestWriteValidationProblemResponse_StripsUpstreamEncodingHeaders(t *testing.T) {
	rec := httptest.NewRecorder()

	// Simulate upstream headers already copied onto the response writer.
	rec.Header().Set("Accept-Ranges", "bytes")
	rec.Header().Set("Age", "60")
	rec.Header().Set("Content-Length", "9999")
	rec.Header().Set("Transfer-Encoding", "chunked")
	rec.Header().Set("Content-Encoding", "gzip")
	rec.Header().Set("ETag", "\"abc123\"")
	rec.Header().Set("Expires", "Thu, 01 Jan 1970 00:00:00 GMT")
	rec.Header().Set("Last-Modified", "Wed, 21 Oct 2015 07:28:00 GMT")
	rec.Header().Set("Cache-Control", "max-age=600")

	writeValidationProblemResponse(rec, 502, "/foo",
		nil, buildSampleErrors("bad response"))

	assert.Empty(t, rec.Header().Get("Accept-Ranges"),
		"Accept-Ranges must be stripped for the substituted body")
	assert.Empty(t, rec.Header().Get("Age"),
		"Age must be stripped; it refers to the upstream representation")
	assert.Empty(t, rec.Header().Get("Content-Length"),
		"Content-Length must be stripped; Go will set the correct one")
	assert.Empty(t, rec.Header().Get("Transfer-Encoding"),
		"Transfer-Encoding must be stripped for the substituted body")
	assert.Empty(t, rec.Header().Get("Content-Encoding"),
		"Content-Encoding must be stripped — substituted body is plain JSON")
	assert.Empty(t, rec.Header().Get("ETag"),
		"ETag must be stripped for the synthesized problem document")
	assert.Empty(t, rec.Header().Get("Expires"),
		"Expires must be stripped for the synthesized problem document")
	assert.Empty(t, rec.Header().Get("Last-Modified"),
		"Last-Modified must be stripped for the synthesized problem document")
	assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"),
		"problem responses must not inherit cacheability from upstream content")
}

func TestWriteValidationProblemResponse_BodyShape(t *testing.T) {
	rec := httptest.NewRecorder()

	writeValidationProblemResponse(rec, 400, "/foo",
		buildSampleErrors("missing property"), nil)

	var decoded map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &decoded))

	assert.Equal(t, shared.ValidationProblemType, decoded["type"])
	assert.Equal(t, "Request validation failed", decoded["title"])
	assert.Equal(t, float64(400), decoded["status"])
	assert.Equal(t, "/foo", decoded["instance"])
	assert.NotNil(t, decoded["requestErrors"])
	_, hasResp := decoded["responseErrors"]
	assert.False(t, hasResp, "empty responseErrors must be omitted")
}

func TestWriteValidationProblemResponse_BothSides(t *testing.T) {
	rec := httptest.NewRecorder()

	writeValidationProblemResponse(rec, 502, "/baz",
		buildSampleErrors("req bad"),
		buildSampleErrors("resp bad 1", "resp bad 2"))

	var decoded map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &decoded))

	assert.Equal(t, "Request and response validation failed", decoded["title"])
	assert.Len(t, decoded["requestErrors"], 1)
	assert.Len(t, decoded["responseErrors"], 2)
}
