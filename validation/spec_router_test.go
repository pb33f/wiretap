// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package validation

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/pb33f/libopenapi"
	"github.com/stretchr/testify/assert"
)

func TestSpecRouterResolveSingleValidator(t *testing.T) {
	router := NewSpecRouter([]DocumentValidator{{
		DocumentName: "only",
	}})

	req, _ := http.NewRequest(http.MethodGet, "http://wiretap.local/anything", nil)
	doc := router.Resolve(req)

	assert.NotNil(t, doc)
	assert.Equal(t, "only", doc.DocumentName)
	assert.Nil(t, router.cache)
}

func TestSpecRouterResolveFindsMatchingValidator(t *testing.T) {
	router := NewSpecRouter([]DocumentValidator{
		buildRouterSpec(t, "pets", "/pets/{petId}"),
		buildRouterSpec(t, "orders", "/orders/{orderId}"),
		buildRouterSpec(t, "users", "/users/{userId}/settings"),
	})

	petReq, _ := http.NewRequest(http.MethodGet, "http://wiretap.local/pets/123", nil)
	orderReq, _ := http.NewRequest(http.MethodGet, "http://wiretap.local/orders/abc", nil)
	userReq, _ := http.NewRequest(http.MethodGet, "http://wiretap.local/users/abc/settings", nil)

	assert.Equal(t, "pets", router.Resolve(petReq).DocumentName)
	assert.Equal(t, "orders", router.Resolve(orderReq).DocumentName)
	assert.Equal(t, "users", router.Resolve(userReq).DocumentName)
}

func TestSpecRouterResolveFallsBackToFirstValidator(t *testing.T) {
	router := NewSpecRouter([]DocumentValidator{
		buildRouterSpec(t, "pets", "/pets/{petId}"),
		buildRouterSpec(t, "orders", "/orders/{orderId}"),
	})

	req, _ := http.NewRequest(http.MethodGet, "http://wiretap.local/not-found", nil)
	doc := router.Resolve(req)

	assert.NotNil(t, doc)
	assert.Equal(t, "pets", doc.DocumentName)
}

func TestSpecRouterResolveCachesByMethodAndPath(t *testing.T) {
	router := NewSpecRouter([]DocumentValidator{
		buildRouterSpec(t, "pets", "/pets/{petId}"),
		buildRouterSpec(t, "orders", "/orders/{orderId}"),
	})

	req, _ := http.NewRequest(http.MethodGet, "http://wiretap.local/orders/abc", nil)
	doc := router.Resolve(req)

	assert.Equal(t, "orders", doc.DocumentName)
	assert.NotNil(t, router.cache)
	assert.Equal(t, 1, router.cache.Len())

	doc = router.Resolve(req)

	assert.Equal(t, "orders", doc.DocumentName)
	assert.Equal(t, 1, router.cache.Len())

	postReq, _ := http.NewRequest(http.MethodPost, "http://wiretap.local/orders/abc", nil)
	doc = router.Resolve(postReq)

	assert.Equal(t, "orders", doc.DocumentName)
	assert.Equal(t, 2, router.cache.Len())
}

func buildRouterSpec(t *testing.T, name, path string) DocumentValidator {
	t.Helper()

	spec := []byte(fmt.Sprintf(`openapi: 3.1.0
info:
  title: %s
  version: "1.0"
paths:
  "%s":
    get:
      responses:
        "200":
          description: ok
`, name, path))

	doc, err := libopenapi.NewDocument(spec)
	if err != nil {
		t.Fatalf("failed to parse %s spec: %v", name, err)
	}
	model, err := doc.BuildV3Model()
	if err != nil {
		t.Fatalf("failed to build %s model: %v", name, err)
	}

	return DocumentValidator{
		DocumentName: name,
		DocModel:     &model.Model,
	}
}
