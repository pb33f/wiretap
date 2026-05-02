// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package validator

import (
	"net/http"
	"testing"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/wiretap/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateRequestUsesMethodSpecificSpecWhenRoutesMatch(t *testing.T) {
	validator := New([]DocumentValidator{
		buildDocumentValidator(t, "get-foo", `openapi: 3.1.0
info:
  title: get foo
  version: "1.0"
paths:
  "/foo":
    get:
      responses:
        "200":
          description: ok
`),
		buildDocumentValidator(t, "post-foo", `openapi: 3.1.0
info:
  title: post foo
  version: "1.0"
paths:
  "/foo":
    post:
      responses:
        "200":
          description: ok
`),
	})

	req, err := http.NewRequest(http.MethodPost, "http://wiretap.local/foo", nil)
	require.NoError(t, err)

	assert.Equal(t, "post-foo", validator.GetValidatorForHTTPRequest(req).DocumentName)
	assert.Empty(t, validator.ValidateRequest(nil, req))
}

func TestValidateRequestStripsPathLevelServerBase(t *testing.T) {
	validator := New([]DocumentValidator{
		buildDocumentValidator(t, "users", `openapi: 3.1.0
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
		buildDocumentValidator(t, "accounts", `openapi: 3.1.0
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

	req, err := http.NewRequest(http.MethodGet, "http://wiretap.local/accounts/health", nil)
	require.NoError(t, err)

	assert.Equal(t, "accounts", validator.GetValidatorForHTTPRequest(req).DocumentName)
	assert.Empty(t, validator.ValidateRequest(nil, req))
}

func TestGetValidatorAndRequestForHTTPRequestStripsPathLevelServerBase(t *testing.T) {
	validator := New([]DocumentValidator{
		buildDocumentValidator(t, "accounts", `openapi: 3.1.0
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

	req, err := http.NewRequest(http.MethodGet, "http://wiretap.local/accounts/health", nil)
	require.NoError(t, err)

	docValidator, validationRequest := validator.GetValidatorAndRequestForHTTPRequest(req)

	require.NotNil(t, docValidator)
	require.NotNil(t, validationRequest)
	assert.Equal(t, "accounts", docValidator.DocumentName)
	assert.Equal(t, "/accounts/health", req.URL.Path)
	assert.Equal(t, "/health", validationRequest.URL.Path)
}

func TestValidateRequestPreservesEscapedSlashWhenStrippingServerBase(t *testing.T) {
	validator := New([]DocumentValidator{
		buildDocumentValidator(t, "files", `openapi: 3.1.0
info:
  title: files
  version: "1.0"
paths:
  "/files/{name}":
    servers:
      - url: https://api.example.com/api
    get:
      parameters:
        - name: name
          in: path
          required: true
          schema:
            type: string
      responses:
        "200":
          description: ok
`),
	})

	req, err := http.NewRequest(http.MethodGet, "http://wiretap.local/api/files/a%2Fb", nil)
	require.NoError(t, err)

	docValidator, validationRequest := validator.GetValidatorAndRequestForHTTPRequest(req)

	require.NotNil(t, docValidator)
	require.NotNil(t, validationRequest)
	assert.Equal(t, "files", docValidator.DocumentName)
	assert.Equal(t, "/files/a/b", validationRequest.URL.Path)
	assert.Equal(t, "/files/a%2Fb", validationRequest.URL.RawPath)
	assert.Equal(t, "/files/a%2Fb", validationRequest.URL.EscapedPath())
	assert.Empty(t, validator.ValidateRequest(nil, req))
}

func TestValidateRequestStripsOperationLevelServerBase(t *testing.T) {
	validator := New([]DocumentValidator{
		buildDocumentValidator(t, "users", `openapi: 3.1.0
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
		buildDocumentValidator(t, "accounts", `openapi: 3.1.0
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

	req, err := http.NewRequest(http.MethodGet, "http://wiretap.local/accounts/health", nil)
	require.NoError(t, err)

	assert.Equal(t, "accounts", validator.GetValidatorForHTTPRequest(req).DocumentName)
	assert.Empty(t, validator.ValidateRequest(nil, req))
}

func TestValidateRequestDoesNotStripAnotherOperationServerBase(t *testing.T) {
	validator := New([]DocumentValidator{
		buildDocumentValidator(t, "split", `openapi: 3.1.0
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

	req, err := http.NewRequest(http.MethodGet, "http://wiretap.local/post/health", nil)
	require.NoError(t, err)

	_, validationRequest := validator.GetValidatorAndRequestForHTTPRequest(req)
	require.NotNil(t, validationRequest)
	assert.Equal(t, "/post/health", validationRequest.URL.Path)

	errs := validator.ValidateRequest(nil, req)
	require.NotEmpty(t, errs)
	assert.True(t, errs[0].IsPathMissingError())
}

func buildDocumentValidator(t *testing.T, name, spec string) DocumentValidator {
	t.Helper()

	doc, err := libopenapi.NewDocument([]byte(spec))
	require.NoError(t, err)
	model, err := doc.BuildV3Model()
	require.NoError(t, err)

	return DocumentValidator{
		DocumentName: name,
		Document:     doc,
		DocModel:     &model.Model,
		Validator:    validation.NewHttpValidator(&model.Model),
	}
}
