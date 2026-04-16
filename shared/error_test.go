// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package shared

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/pb33f/libopenapi-validator/errors"
	"github.com/stretchr/testify/assert"
)

func sampleValidationErrors(messages ...string) []*WiretapValidationError {
	out := make([]*WiretapValidationError, 0, len(messages))
	for _, m := range messages {
		out = append(out, &WiretapValidationError{
			ValidationError: errors.ValidationError{Message: m},
			SpecName:        "spec.yaml",
		})
	}
	return out
}

func TestBuildValidationProblem_RequestOnly(t *testing.T) {
	p := BuildValidationProblem(400, "/foo", sampleValidationErrors("a", "b"), nil)

	assert.Equal(t, "Request validation failed", p.Title)
	assert.Equal(t, 400, p.Status)
	assert.Equal(t, "/foo", p.Instance)
	assert.Equal(t, ValidationProblemType, p.Type)
	assert.Equal(t, "2 validation error(s) detected", p.Detail)
	assert.Len(t, p.RequestErrors, 2)
	assert.Empty(t, p.ResponseErrors)
	assert.Nil(t, p.Payload, "Payload must stay nil so its semantic meaning is preserved")
}

func TestBuildValidationProblem_ResponseOnly(t *testing.T) {
	p := BuildValidationProblem(502, "/bar", nil, sampleValidationErrors("c"))

	assert.Equal(t, "Response validation failed", p.Title)
	assert.Equal(t, 502, p.Status)
	assert.Equal(t, "1 validation error(s) detected", p.Detail)
	assert.Empty(t, p.RequestErrors)
	assert.Len(t, p.ResponseErrors, 1)
}

func TestBuildValidationProblem_BothSides(t *testing.T) {
	p := BuildValidationProblem(502, "/baz",
		sampleValidationErrors("a"),
		sampleValidationErrors("b", "c"))

	assert.Equal(t, "Request and response validation failed", p.Title)
	assert.Equal(t, "3 validation error(s) detected", p.Detail)
	assert.Len(t, p.RequestErrors, 1)
	assert.Len(t, p.ResponseErrors, 2)
}

func TestMarshalValidationProblem_NilGuard(t *testing.T) {
	assert.Equal(t, []byte("{}"), MarshalValidationProblem(nil))
	assert.Equal(t, []byte("{}"), MarshalValidationProblem(&ValidationProblem{}))
}

func TestMarshalValidationProblem_ShapeAndOmitempty(t *testing.T) {
	p := BuildValidationProblem(400, "/only-request",
		sampleValidationErrors("missing property"), nil)

	raw := MarshalValidationProblem(p)

	// Unmarshal into a loose map so we can assert on shape + omitempty.
	var decoded map[string]any
	assert.NoError(t, json.Unmarshal(raw, &decoded))

	assert.Equal(t, ValidationProblemType, decoded["type"])
	assert.Equal(t, "Request validation failed", decoded["title"])
	assert.Equal(t, float64(400), decoded["status"])
	assert.Equal(t, "/only-request", decoded["instance"])
	assert.NotNil(t, decoded["requestErrors"])
	_, hasResp := decoded["responseErrors"]
	assert.False(t, hasResp, "responseErrors must be omitted when empty")
	_, hasPayload := decoded["payload"]
	assert.False(t, hasPayload, "payload must be omitted when nil")

	// Content-type check isn't a thing here, but sanity-check that the wire
	// shape is what a client would see.
	assert.True(t, strings.HasPrefix(string(raw), "{"))
}
