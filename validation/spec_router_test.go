// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package validation

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/pb33f/libopenapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestSpecRouterResolveUsesMethodWhenSpecsShareRoute(t *testing.T) {
	router := NewSpecRouter([]DocumentValidator{
		buildRouterSpecWithMethod(t, "get-foo", "/foo", http.MethodGet),
		buildRouterSpecWithMethod(t, "post-foo", "/foo", http.MethodPost),
	})

	req, _ := http.NewRequest(http.MethodPost, "http://wiretap.local/foo", nil)
	match := router.ResolveMatch(req)

	assert.NotNil(t, match)
	assert.Equal(t, "post-foo", match.Document.DocumentName)
	assert.Equal(t, "/foo", match.MatchedPath)
	assert.True(t, match.MethodMatched)
}

func TestSpecRouterResolveFallsBackToPathMatchWhenMethodMissing(t *testing.T) {
	router := NewSpecRouter([]DocumentValidator{
		buildRouterSpec(t, "get-foo", "/foo"),
		buildRouterSpec(t, "get-bar", "/bar"),
	})

	req, _ := http.NewRequest(http.MethodDelete, "http://wiretap.local/foo", nil)
	match := router.ResolveMatch(req)

	assert.NotNil(t, match)
	assert.Equal(t, "get-foo", match.Document.DocumentName)
	assert.Equal(t, "/foo", match.MatchedPath)
	assert.False(t, match.MethodMatched)
}

func TestSpecRouterResolveUsesPathLevelServerRoutes(t *testing.T) {
	router := NewSpecRouter([]DocumentValidator{
		buildRouterSpecFromSource(t, "users", `openapi: 3.1.0
info:
  title: users
  version: "1.0"
paths:
  "/health":
    servers:
      - url: https://api.example.com/users
    get:
      responses:
        "200":
          description: ok
`),
		buildRouterSpecFromSource(t, "accounts", `openapi: 3.1.0
info:
  title: accounts
  version: "1.0"
paths:
  "/health":
    servers:
      - url: https://api.example.com/accounts
    get:
      responses:
        "200":
          description: ok
`),
	})

	req, _ := http.NewRequest(http.MethodGet, "http://wiretap.local/accounts/health", nil)
	match := router.ResolveMatch(req)

	assert.NotNil(t, match)
	assert.Equal(t, "accounts", match.Document.DocumentName)
	assert.Equal(t, "/health", match.MatchedPath)
	assert.Equal(t, "/accounts/health", match.EffectiveRoutePath)
	assert.Equal(t, "/accounts", match.BasePath)
	assert.True(t, match.MethodMatched)
}

func TestSpecRouterScoresEffectiveRoutePaths(t *testing.T) {
	router := NewSpecRouter([]DocumentValidator{
		buildRouterSpecFromSource(t, "users", `openapi: 3.1.0
info:
  title: users
  version: "1.0"
servers:
  - url: https://api.example.com/api
paths:
  "/users/{id}":
    get:
      responses:
        "200":
          description: ok
  "/me":
    servers:
      - url: https://api.example.com/api/users
    get:
      responses:
        "200":
          description: ok
`),
	})

	req, _ := http.NewRequest(http.MethodGet, "http://wiretap.local/api/users/me", nil)
	match := router.ResolveMatch(req)

	assert.NotNil(t, match)
	assert.Equal(t, "/me", match.MatchedPath)
	assert.Equal(t, "/api/users/me", match.EffectiveRoutePath)
	assert.Equal(t, "/api/users", match.BasePath)
	assert.True(t, match.MethodMatched)
}

func TestSpecRouterResolveUsesOperationLevelServerRoutes(t *testing.T) {
	router := NewSpecRouter([]DocumentValidator{
		buildRouterSpecFromSource(t, "users", `openapi: 3.1.0
info:
  title: users
  version: "1.0"
servers:
  - url: https://api.example.com/document
paths:
  "/health":
    servers:
      - url: https://api.example.com/path
    get:
      servers:
        - url: https://api.example.com/users
      responses:
        "200":
          description: ok
`),
		buildRouterSpecFromSource(t, "accounts", `openapi: 3.1.0
info:
  title: accounts
  version: "1.0"
servers:
  - url: https://api.example.com/document
paths:
  "/health":
    servers:
      - url: https://api.example.com/path
    get:
      servers:
        - url: https://api.example.com/accounts
      responses:
        "200":
          description: ok
`),
	})

	req, _ := http.NewRequest(http.MethodGet, "http://wiretap.local/accounts/health", nil)
	match := router.ResolveMatch(req)

	assert.NotNil(t, match)
	assert.Equal(t, "accounts", match.Document.DocumentName)
	assert.Equal(t, "/health", match.MatchedPath)
	assert.Equal(t, "/accounts/health", match.EffectiveRoutePath)
	assert.Equal(t, "/accounts", match.BasePath)
	assert.True(t, match.MethodMatched)
}

func TestSpecRouterDoesNotFallbackToAnotherOperationsServerBase(t *testing.T) {
	router := NewSpecRouter([]DocumentValidator{
		buildRouterSpecFromSource(t, "split", `openapi: 3.1.0
info:
  title: split
  version: "1.0"
paths:
  "/health":
    get:
      servers:
        - url: https://api.example.com/get
      responses:
        "200":
          description: ok
    post:
      servers:
        - url: https://api.example.com/post
      responses:
        "200":
          description: ok
`),
	})

	req, _ := http.NewRequest(http.MethodGet, "http://wiretap.local/post/health", nil)
	match := router.ResolveMatch(req)

	assert.NotNil(t, match)
	assert.Empty(t, match.MatchedPath)
	assert.Empty(t, match.BasePath)
	assert.False(t, match.MethodMatched)
	assert.Same(t, req, ValidationRequestForRouteMatch(req, match))
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

func TestSpecRouterResolveCachesByEscapedPath(t *testing.T) {
	router := NewSpecRouter([]DocumentValidator{
		buildRouterSpec(t, "encoded", "/files/{name}"),
		buildRouterSpec(t, "nested", "/files/{dir}/{name}"),
	})

	encodedReq, _ := http.NewRequest(http.MethodGet, "http://wiretap.local/files/a%2Fb", nil)
	encodedMatch := router.ResolveMatch(encodedReq)

	require.NotNil(t, encodedMatch)
	assert.Equal(t, "encoded", encodedMatch.Document.DocumentName)
	assert.Equal(t, "/files/{name}", encodedMatch.MatchedPath)
	assert.Equal(t, 1, router.cache.Len())

	nestedReq, _ := http.NewRequest(http.MethodGet, "http://wiretap.local/files/a/b", nil)
	nestedMatch := router.ResolveMatch(nestedReq)

	require.NotNil(t, nestedMatch)
	assert.Equal(t, "nested", nestedMatch.Document.DocumentName)
	assert.Equal(t, "/files/{dir}/{name}", nestedMatch.MatchedPath)
	assert.Equal(t, 2, router.cache.Len())
}

func TestValidationRequestForRouteMatchPreservesEscapedPath(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "http://wiretap.local/api/files/a%2Fb", nil)
	match := &RouteMatch{BasePath: "/api"}

	cloned := ValidationRequestForRouteMatch(req, match)

	assert.NotSame(t, req, cloned)
	assert.Equal(t, "/files/a/b", cloned.URL.Path)
	assert.Equal(t, "/files/a%2Fb", cloned.URL.RawPath)
	assert.Equal(t, "/files/a%2Fb", cloned.URL.EscapedPath())
}

func TestValidationRequestForRouteMatchClonesBodyFromGetBody(t *testing.T) {
	body := `{"name":"widget"}`
	req, _ := http.NewRequest(http.MethodPost, "http://wiretap.local/api/widgets", nil)
	req.Body = io.NopCloser(strings.NewReader(body))
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(body)), nil
	}
	req.ContentLength = int64(len(body))
	match := &RouteMatch{BasePath: "/api"}

	cloneA := ValidationRequestForRouteMatch(req, match)
	cloneB := ValidationRequestForRouteMatch(req, match)

	assert.NotSame(t, req, cloneA)
	assert.NotSame(t, req, cloneB)
	assert.Equal(t, "/widgets", cloneA.URL.Path)
	assert.Equal(t, "/widgets", cloneB.URL.Path)

	cloneABody, err := io.ReadAll(cloneA.Body)
	assert.NoError(t, err)
	assert.NoError(t, cloneA.Body.Close())
	assert.Equal(t, body, string(cloneABody))

	cloneBBody, err := io.ReadAll(cloneB.Body)
	assert.NoError(t, err)
	assert.NoError(t, cloneB.Body.Close())
	assert.Equal(t, body, string(cloneBBody))

	originalBody, err := io.ReadAll(req.Body)
	assert.NoError(t, err)
	assert.Equal(t, body, string(originalBody))
}

func TestValidationRequestForRouteMatchDoesNotConsumeBodyWithoutGetBody(t *testing.T) {
	body := `{"name":"widget"}`
	req, _ := http.NewRequest(http.MethodPost, "http://wiretap.local/api/widgets", nil)
	req.Body = io.NopCloser(strings.NewReader(body))
	req.ContentLength = int64(len(body))
	match := &RouteMatch{BasePath: "/api"}

	cloned := ValidationRequestForRouteMatch(req, match)

	assert.NotSame(t, req, cloned)
	assert.Equal(t, "/widgets", cloned.URL.Path)
	assert.Nil(t, cloned.GetBody)

	originalBody, err := io.ReadAll(req.Body)
	assert.NoError(t, err)
	assert.Equal(t, body, string(originalBody))
}

func buildRouterSpec(t *testing.T, name, path string) DocumentValidator {
	t.Helper()
	return buildRouterSpecWithMethod(t, name, path, http.MethodGet)
}

func buildRouterSpecWithMethod(t *testing.T, name, path, method string) DocumentValidator {
	t.Helper()

	spec := []byte(fmt.Sprintf(`openapi: 3.1.0
info:
  title: %s
  version: "1.0"
paths:
  "%s":
    %s:
      responses:
        "200":
          description: ok
`, name, path, strings.ToLower(method)))

	return buildRouterSpecFromSource(t, name, string(spec))
}

func buildRouterSpecFromSource(t *testing.T, name, spec string) DocumentValidator {
	t.Helper()

	doc, err := libopenapi.NewDocument([]byte(spec))
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
