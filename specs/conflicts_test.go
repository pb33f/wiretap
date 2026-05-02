// Copyright 2026 Princess Beef Heavy Industries LLC
// SPDX-License-Identifier: AGPL

package specs

import (
	"fmt"
	"testing"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/wiretap/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyzeCrossSpecAmbiguous(t *testing.T) {
	report := Analyze([]shared.ApiDocument{
		buildSpecDoc(t, "users.yaml", "", "/users/{id}", "get", "getUser"),
		buildSpecDoc(t, "accounts.yaml", "", "/users/{name}", "get", "getAccount"),
	})

	require.Len(t, report.Conflicts, 1)
	assert.Equal(t, KindCrossSpecAmbiguous, report.Conflicts[0].Kind)
	assert.NotEmpty(t, report.RouteIndex.Lookup("GET", "/users/{id}"))
}

func TestAnalyzeCrossSpecLiteralTemplateOverlap(t *testing.T) {
	report := Analyze([]shared.ApiDocument{
		buildSpecDoc(t, "users.yaml", "", "/users/{id}", "get", "getUser"),
		buildSpecDoc(t, "profile.yaml", "", "/users/me", "get", "getCurrentUser"),
	})

	require.Len(t, report.Conflicts, 1)
	assert.Equal(t, KindCrossSpecAmbiguous, report.Conflicts[0].Kind)
	assert.Equal(t, []string{"/users/{id}", "/users/me"}, report.Conflicts[0].RoutePaths)
	assert.NotEmpty(t, report.RouteIndex.Lookup("GET", "/users/{id}"))
	assert.NotEmpty(t, report.RouteIndex.Lookup("GET", "/users/me"))
}

func TestAnalyzeCrossSpecTypedPathParamsOverlap(t *testing.T) {
	report := Analyze([]shared.ApiDocument{
		buildSpecDocFromSource(t, "users-by-id.yaml", `openapi: 3.1.0
info:
  title: users by id
  version: "1.0"
paths:
  "/users/{id}":
    get:
      operationId: getUserByID
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
      responses:
        "200":
          description: ok
`),
		buildSpecDocFromSource(t, "users-by-slug.yaml", `openapi: 3.1.0
info:
  title: users by slug
  version: "1.0"
paths:
  "/users/{slug}":
    get:
      operationId: getUserBySlug
      parameters:
        - name: slug
          in: path
          required: true
          schema:
            type: string
      responses:
        "200":
          description: ok
`),
	})

	require.Len(t, report.Conflicts, 1)
	assert.Equal(t, KindCrossSpecAmbiguous, report.Conflicts[0].Kind)
	assert.Equal(t, []string{"/users/{id}", "/users/{slug}"}, report.Conflicts[0].RoutePaths)
	assert.NotEmpty(t, report.RouteIndex.Lookup("GET", "/users/{id}"))
	assert.NotEmpty(t, report.RouteIndex.Lookup("GET", "/users/{slug}"))
}

func TestAnalyzeCrossSpecDisjointTypedPathParamsAllowed(t *testing.T) {
	report := Analyze([]shared.ApiDocument{
		buildSpecDocFromSource(t, "things-by-id.yaml", `openapi: 3.1.0
info:
  title: things by id
  version: "1.0"
paths:
  "/things/{id}":
    get:
      operationId: getThingByID
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
      responses:
        "200":
          description: ok
`),
		buildSpecDocFromSource(t, "things-by-flag.yaml", `openapi: 3.1.0
info:
  title: things by flag
  version: "1.0"
paths:
  "/things/{flag}":
    get:
      operationId: getThingByFlag
      parameters:
        - name: flag
          in: path
          required: true
          schema:
            type: boolean
      responses:
        "200":
          description: ok
`),
	})

	assert.Empty(t, report.Conflicts)
	assert.Empty(t, report.RouteIndex.Lookup("GET", "/things/{id}"))
	assert.Empty(t, report.RouteIndex.Lookup("GET", "/things/{flag}"))
}

func TestAnalyzeWithinSpecLiteralTemplateOverlapAllowed(t *testing.T) {
	report := Analyze([]shared.ApiDocument{
		buildSpecDocFromPaths(t, "users.yaml", "", map[string]string{
			"/users/{id}": "getUser",
			"/users/me":   "getCurrentUser",
		}),
	})

	assert.Empty(t, report.Conflicts)
}

func TestAnalyzeCrossSpecDuplicate(t *testing.T) {
	report := Analyze([]shared.ApiDocument{
		buildSpecDoc(t, "users.yaml", "", "/health", "get", "healthA"),
		buildSpecDoc(t, "accounts.yaml", "", "/health", "get", "healthB"),
	})

	require.Len(t, report.Conflicts, 1)
	assert.Equal(t, KindCrossSpecDuplicate, report.Conflicts[0].Kind)
}

func TestAnalyzeWithinSpecEffectiveRouteDuplicate(t *testing.T) {
	doc := buildSpecDocFromSource(t, "api.yaml", `openapi: 3.1.0
info:
  title: api
  version: "1.0"
servers:
  - url: https://api.example.com/api
paths:
  "/foo/bar":
    get:
      operationId: getFooBar
      responses:
        "200":
          description: ok
  "/bar":
    servers:
      - url: https://api.example.com/api/foo
    get:
      operationId: getBar
      responses:
        "200":
          description: ok
`)

	report := Analyze([]shared.ApiDocument{doc})

	require.Len(t, report.Conflicts, 1)
	assert.Equal(t, KindWithinSpecDuplicate, report.Conflicts[0].Kind)
	assert.Equal(t, []string{"/foo/bar", "/bar"}, report.Conflicts[0].Paths)
	assert.Equal(t, []string{"/api/foo/bar", "/api/foo/bar"}, report.Conflicts[0].RoutePaths)
	assert.Len(t, report.RouteIndex.Lookup("GET", "/api/foo/bar"), 2)
}

func TestPathConflictSkipsSameOperation(t *testing.T) {
	conflict, ok := pathConflict(operationEntry{
		specIndex: 0,
		path:      "/health",
		routePath: "/api/health",
		method:    "GET",
	}, operationEntry{
		specIndex: 0,
		path:      "/health",
		routePath: "/api/health",
		method:    "GET",
	})

	assert.False(t, ok)
	assert.Empty(t, conflict)
}

func TestAnalyzeWithinSpecAmbiguous(t *testing.T) {
	report := Analyze([]shared.ApiDocument{
		buildSpecDocFromPaths(t, "users.yaml", "", map[string]string{
			"/users/{id}":   "getUserByID",
			"/users/{name}": "getUserByName",
		}),
	})

	require.Len(t, report.Conflicts, 1)
	assert.Equal(t, KindWithinSpecAmbiguous, report.Conflicts[0].Kind)
}

func TestAnalyzeDuplicateOperationID(t *testing.T) {
	report := Analyze([]shared.ApiDocument{
		buildSpecDocFromPaths(t, "users.yaml", "", map[string]string{
			"/users":    "listThings",
			"/accounts": "listThings",
		}),
	})

	require.Len(t, report.Conflicts, 1)
	assert.Equal(t, KindDuplicateOperationID, report.Conflicts[0].Kind)
	assert.Equal(t, "listThings", report.Conflicts[0].OperationID)
}

func TestAnalyzeCrossSpecDuplicateOperationID(t *testing.T) {
	report := Analyze([]shared.ApiDocument{
		buildSpecDoc(t, "users.yaml", "/users-api", "/health", "get", "getHealth"),
		buildSpecDoc(t, "accounts.yaml", "/accounts-api", "/health", "get", "getHealth"),
	})

	require.Len(t, report.Conflicts, 1)
	assert.Equal(t, KindDuplicateOperationID, report.Conflicts[0].Kind)
	assert.Equal(t, "getHealth", report.Conflicts[0].OperationID)
}

func TestAnalyzeIgnoresDuplicateOperationIDWhenConfigured(t *testing.T) {
	report := Analyze([]shared.ApiDocument{
		buildSpecDoc(t, "users.yaml", "/users-api", "/health", "get", "getHealth"),
		buildSpecDoc(t, "accounts.yaml", "/accounts-api", "/health", "get", "getHealth"),
	}, AnalyzeOptions{IgnoreClashingOperationID: true})

	assert.Empty(t, report.Conflicts)
}

func TestAnalyzeServerBasePathsAvoidFalseDuplicate(t *testing.T) {
	report := Analyze([]shared.ApiDocument{
		buildSpecDoc(t, "users.yaml", "/users-api", "/health", "get", "healthUsers"),
		buildSpecDoc(t, "accounts.yaml", "/accounts-api", "/health", "get", "healthAccounts"),
	})

	assert.Empty(t, report.Conflicts)
}

func TestAnalyzePathLevelServersOverrideDocumentServers(t *testing.T) {
	usersDoc := buildSpecDocFromSource(t, "users.yaml", `openapi: 3.1.0
info:
  title: users
  version: "1.0"
servers:
  - url: https://api.example.com/shared
paths:
  "/health":
    servers:
      - url: https://api.example.com/users
    get:
      operationId: healthUsers
      responses:
        "200":
          description: ok
`)
	accountsDoc := buildSpecDocFromSource(t, "accounts.yaml", `openapi: 3.1.0
info:
  title: accounts
  version: "1.0"
servers:
  - url: https://api.example.com/shared
paths:
  "/health":
    servers:
      - url: https://api.example.com/accounts
    get:
      operationId: healthAccounts
      responses:
        "200":
          description: ok
`)

	report := Analyze([]shared.ApiDocument{usersDoc, accountsDoc})
	assert.Empty(t, report.Conflicts)
}

func TestAnalyzeOperationLevelServersOverridePathServers(t *testing.T) {
	usersDoc := buildSpecDocFromSource(t, "users.yaml", `openapi: 3.1.0
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
      operationId: healthUsers
      servers:
        - url: https://api.example.com/users
      responses:
        "200":
          description: ok
`)
	accountsDoc := buildSpecDocFromSource(t, "accounts.yaml", `openapi: 3.1.0
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
      operationId: healthAccounts
      servers:
        - url: https://api.example.com/accounts
      responses:
        "200":
          description: ok
`)

	report := Analyze([]shared.ApiDocument{usersDoc, accountsDoc})
	assert.Empty(t, report.Conflicts)
}

func TestAnalyzeServerVariableDefaultsAvoidFalseDuplicate(t *testing.T) {
	usersDoc := buildSpecDocWithServerBlock(t, "users.yaml", `servers:
  - url: https://api.example.com/api/{service}
    variables:
      service:
        default: users
`, "get", map[string]string{
		"/health": "getUsersHealth",
	})
	accountsDoc := buildSpecDocWithServerBlock(t, "accounts.yaml", `servers:
  - url: https://api.example.com/api/{service}
    variables:
      service:
        default: accounts
`, "get", map[string]string{
		"/health": "getAccountsHealth",
	})

	assert.Equal(t, []string{"/api/users"}, ServerBasePaths(&usersDoc.DocumentModel.Model))
	assert.Equal(t, []string{"/api/users/health"}, EffectiveRoutePaths(&usersDoc.DocumentModel.Model, "/health"))

	report := Analyze([]shared.ApiDocument{usersDoc, accountsDoc})
	assert.Empty(t, report.Conflicts)
}

func TestAnalyzeServerVariableDefaultsKeepRealDuplicate(t *testing.T) {
	report := Analyze([]shared.ApiDocument{
		buildSpecDocWithServerBlock(t, "users.yaml", `servers:
  - url: https://api.example.com/api/{service}
    variables:
      service:
        default: users
`, "get", map[string]string{
			"/health": "healthUsers",
		}),
		buildSpecDocWithServerBlock(t, "accounts.yaml", `servers:
  - url: https://api.example.com/api/{service}
    variables:
      service:
        default: users
`, "get", map[string]string{
			"/health": "healthAccounts",
		}),
	})

	require.Len(t, report.Conflicts, 1)
	assert.Equal(t, KindCrossSpecDuplicate, report.Conflicts[0].Kind)
	assert.Equal(t, []string{"/api/users/health", "/api/users/health"}, report.Conflicts[0].RoutePaths)
}

func TestAnalyzeTrailingSlashRoutesAreDistinct(t *testing.T) {
	trailingDoc := buildSpecDoc(t, "trailing.yaml", "", "/foo/", "get", "getFooTrailing")
	report := Analyze([]shared.ApiDocument{
		buildSpecDoc(t, "plain.yaml", "", "/foo", "get", "getFoo"),
		trailingDoc,
	})

	assert.Equal(t, []string{"/foo/"}, EffectiveRoutePaths(&trailingDoc.DocumentModel.Model, "/foo/"))
	assert.Empty(t, report.Conflicts)
}

func buildSpecDoc(t *testing.T, name, serverPath, pathName, method, operationID string) shared.ApiDocument {
	t.Helper()
	return buildSpecDocFromPathsWithMethod(t, name, serverPath, method, map[string]string{
		pathName: operationID,
	})
}

func buildSpecDocFromPaths(t *testing.T, name, serverPath string, paths map[string]string) shared.ApiDocument {
	t.Helper()
	return buildSpecDocFromPathsWithMethod(t, name, serverPath, "get", paths)
}

func buildSpecDocFromPathsWithMethod(t *testing.T, name, serverPath, method string, paths map[string]string) shared.ApiDocument {
	t.Helper()

	server := ""
	if serverPath != "" {
		server = fmt.Sprintf("servers:\n  - url: https://api.example.com%s\n", serverPath)
	}
	return buildSpecDocWithServerBlock(t, name, server, method, paths)
}

func buildSpecDocWithServerBlock(t *testing.T, name, server, method string, paths map[string]string) shared.ApiDocument {
	t.Helper()

	spec := fmt.Sprintf(`openapi: 3.1.0
info:
  title: %s
  version: "1.0"
%spaths:
`, name, server)
	for pathName, operationID := range paths {
		spec += fmt.Sprintf(`  "%s":
    %s:
      operationId: %s
      responses:
        "200":
          description: ok
`, pathName, method, operationID)
	}

	return buildSpecDocFromSource(t, name, spec)
}

func buildSpecDocFromSource(t *testing.T, name, spec string) shared.ApiDocument {
	t.Helper()

	doc, err := libopenapi.NewDocument([]byte(spec))
	require.NoError(t, err)
	model, err := doc.BuildV3Model()
	require.NoError(t, err)
	return shared.ApiDocument{
		DocumentName:  name,
		Document:      doc,
		DocumentModel: model,
	}
}
