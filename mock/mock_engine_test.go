// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package mock

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi-validator/helpers"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var giftshopBytes []byte

func resetGiftshopState() *v3.Document {
	if len(giftshopBytes) <= 0 {
		var err error
		giftshopBytes, err = os.ReadFile("../testdata/giftshop-openapi.yaml")
		if err != nil {
			panic(err)
		}
	}
	d, _ := libopenapi.NewDocument(giftshopBytes)
	compiled, _ := d.BuildV3Model()
	return &compiled.Model
}

func TestNewMockEngine_findPath(t *testing.T) {
	doc := resetGiftshopState()
	me := NewMockEngine(doc, false, true)

	request, _ := http.NewRequest(http.MethodGet, "https://api.pb33f.io/wiretap/giftshop/products", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")

	path, _ := me.findPath(request)
	assert.NotNil(t, path)
}

func TestNewMockEngine_findPathNegative(t *testing.T) {
	doc := resetGiftshopState()
	me := NewMockEngine(doc, false, true)

	request, _ := http.NewRequest(http.MethodGet, "https://api.pb33f.io/wiretap/giftshop/invalid", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")

	path, errors := me.findPath(request)
	assert.Nil(t, path)
	assert.Error(t, errors)
}

func TestNewMockEngine_findOperation(t *testing.T) {
	doc := resetGiftshopState()
	me := NewMockEngine(doc, false, true)

	request, _ := http.NewRequest(http.MethodGet, "https://api.pb33f.io/wiretap/giftshop/products", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")

	path, _ := me.findPath(request)
	operation := me.findOperation(request, path)
	assert.NotNil(t, operation)

}

func TestNewMockEngine_findOperationNegative(t *testing.T) {
	doc := resetGiftshopState()
	me := NewMockEngine(doc, false, true)

	request, _ := http.NewRequest(http.MethodPatch, "https://api.pb33f.io/wiretap/giftshop/products", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")

	path, _ := me.findPath(request)
	operation := me.findOperation(request, path)
	assert.Nil(t, operation)
}

func TestNewMockEngine_ValidateSecurity_FailAPIKey_Header(t *testing.T) {
	doc := resetGiftshopState()
	me := NewMockEngine(doc, false, true)

	request, _ := http.NewRequest(http.MethodPost, "https://api.pb33f.io/wiretap/giftshop/products", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")
	path, _ := me.findPath(request)
	operation := me.findOperation(request, path)
	err := me.ValidateSecurity(request, operation)
	assert.Error(t, err)
	assert.Equal(t, "apiKey not found, no `X-API-Key` header found in request", err.Error())
}

func TestNewMockEngine_ValidateSecurity_PassAPIKey_Header(t *testing.T) {
	doc := resetGiftshopState()
	me := NewMockEngine(doc, false, true)

	request, _ := http.NewRequest(http.MethodPost, "https://api.pb33f.io/wiretap/giftshop/products", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")
	request.Header.Set("X-API-Key", "doesnotmatter")
	path, _ := me.findPath(request)
	operation := me.findOperation(request, path)
	err := me.ValidateSecurity(request, operation)
	assert.NoError(t, err)
}

func TestNewMockEngine_ValidateSecurity_FailAPIKey_Query(t *testing.T) {
	doc := resetGiftshopState()
	me := NewMockEngine(doc, false, true)

	request, _ := http.NewRequest(http.MethodPost, "https://api.pb33f.io/wiretap/giftshop/products", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")
	path, _ := me.findPath(request)
	operation := me.findOperation(request, path)

	// mutate securityScheme to be a query param
	doc.Components.SecuritySchemes.GetOrZero("ApiKeyAuth").In = "query"
	doc.Components.SecuritySchemes.GetOrZero("ApiKeyAuth").Name = "pizza-cake-burger"

	err := me.ValidateSecurity(request, operation)
	assert.Error(t, err)
	assert.Equal(t, "apiKey not found, no `pizza-cake-burger` query parameter found in request", err.Error())
}

func TestNewMockEngine_ValidateSecurity_PassAPIKey_Query(t *testing.T) {
	doc := resetGiftshopState()
	me := NewMockEngine(doc, false, true)

	request, _ := http.NewRequest(http.MethodPost, "https://api.pb33f.io/wiretap/giftshop/products?pizza-burger-cake=123", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")
	path, _ := me.findPath(request)
	operation := me.findOperation(request, path)

	// mutate securityScheme to be a query param
	doc.Components.SecuritySchemes.GetOrZero("ApiKeyAuth").In = "query"
	doc.Components.SecuritySchemes.GetOrZero("ApiKeyAuth").Name = "pizza-burger-cake"

	err := me.ValidateSecurity(request, operation)
	assert.NoError(t, err)
}

func TestNewMockEngine_ValidateSecurity_FailAPIKey_Cookie(t *testing.T) {
	doc := resetGiftshopState()
	me := NewMockEngine(doc, false, true)

	request, _ := http.NewRequest(http.MethodPost, "https://api.pb33f.io/wiretap/giftshop/products", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")
	path, _ := me.findPath(request)
	operation := me.findOperation(request, path)

	// mutate securityScheme to be a query param
	doc.Components.SecuritySchemes.GetOrZero("ApiKeyAuth").In = "cookie"
	doc.Components.SecuritySchemes.GetOrZero("ApiKeyAuth").Name = "burger-chips-beer"

	err := me.ValidateSecurity(request, operation)
	assert.Error(t, err)
	assert.Equal(t, "apiKey not found, no `burger-chips-beer` cookie found in request", err.Error())
}

func TestNewMockEngine_ValidateSecurity_PassAPIKey_Cookie(t *testing.T) {
	doc := resetGiftshopState()
	me := NewMockEngine(doc, false, true)

	request, _ := http.NewRequest(http.MethodPost, "https://api.pb33f.io/wiretap/giftshop/products", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")
	request.AddCookie(&http.Cookie{
		Name:  "burger-chips-beer",
		Value: "123",
	})
	path, _ := me.findPath(request)
	operation := me.findOperation(request, path)

	// mutate securityScheme to be a query param
	doc.Components.SecuritySchemes.GetOrZero("ApiKeyAuth").In = "cookie"
	doc.Components.SecuritySchemes.GetOrZero("ApiKeyAuth").Name = "burger-chips-beer"

	err := me.ValidateSecurity(request, operation)
	assert.NoError(t, err)
}

func TestNewMockEngine_ValidateSecurity_FailHTTP_Bearer(t *testing.T) {
	doc := resetGiftshopState()
	me := NewMockEngine(doc, false, true)

	request, _ := http.NewRequest(http.MethodPost, "https://api.pb33f.io/wiretap/giftshop/products", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")

	path, _ := me.findPath(request)
	operation := me.findOperation(request, path)

	// mutate securityScheme to be a query param
	doc.Components.SecuritySchemes.GetOrZero("ApiKeyAuth").Type = "http"
	doc.Components.SecuritySchemes.GetOrZero("ApiKeyAuth").Scheme = "bearer"

	err := me.ValidateSecurity(request, operation)
	assert.Error(t, err)
	assert.Equal(t, "bearer authentication failed: bearer token not found, no `Authorization` header found in request", err.Error())
}

func TestNewMockEngine_ValidateSecurity_PassHTTP_Bearer(t *testing.T) {
	doc := resetGiftshopState()
	me := NewMockEngine(doc, false, true)

	request, _ := http.NewRequest(http.MethodPost, "https://api.pb33f.io/wiretap/giftshop/products", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")
	request.Header.Set("Authorization", "Bearer 1234")

	path, _ := me.findPath(request)
	operation := me.findOperation(request, path)

	// mutate securityScheme to be a query param
	doc.Components.SecuritySchemes.GetOrZero("ApiKeyAuth").Type = "http"
	doc.Components.SecuritySchemes.GetOrZero("ApiKeyAuth").Scheme = "bearer"

	err := me.ValidateSecurity(request, operation)
	assert.NoError(t, err)
}

func TestNewMockEngine_BuildResponse_SimpleValid(t *testing.T) {
	doc := resetGiftshopState()
	me := NewMockEngine(doc, false, true)

	request, _ := http.NewRequest(http.MethodGet, "https://api.pb33f.io/wiretap/giftshop/products", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")

	b, status, err := me.GenerateResponse(request)

	assert.NoError(t, err)
	assert.Equal(t, 200, status)

	var decoded []map[string]any
	_ = json.Unmarshal(b, &decoded)

	assert.Equal(t, "pb0001", decoded[0]["shortCode"])
	assert.Equal(t, 19.99, decoded[0]["price"])
}

// test disabled because I have updated the mock engine to drop down to 'application/json' if the content type is not
// found.

//func TestNewMockEngine_BuildResponse_SimpleInvalid_BadContentType(t *testing.T) {
//	doc := resetGiftshopState()
//	me := NewMockEngine(doc, false, true)
//
//	request, _ := http.NewRequest(http.MethodGet, "https://api.pb33f.io/wiretap/giftshop/products", nil)
//	request.Header.Set(helpers.ContentTypeHeader, "cup/tea")
//	b, status, err := me.GenerateResponse(request)
//
//	assert.NoError(t, err)
//	assert.Equal(t, 415, status)
//
//	var decoded map[string]any
//	_ = json.Unmarshal(b, &decoded)
//
//	assert.Equal(t, "Media type not supported (415)", decoded["title"])
//	assert.Equal(t, "The media type requested 'cup/tea' is not supported by this operation", decoded["detail"])
//}

func TestNewMockEngine_BuildResponse_SimpleValid_Pretty(t *testing.T) {
	doc := resetGiftshopState()
	me := NewMockEngine(doc, true, true)

	request, _ := http.NewRequest(http.MethodGet, "https://api.pb33f.io/wiretap/giftshop/products/bd1f3f70-d46f-4ea7-b178-de9a5abfe4d8", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")

	b, status, err := me.GenerateResponse(request)

	assert.NoError(t, err)
	assert.Equal(t, 200, status)
	assert.Equal(t, "{\n  \"category\": \"shirts\",\n  \"description\": \"A t-shirt with the pb33f logo "+
		"on the front\",\n  \"id\": \"d1404c5c-69bd-4cd2-a4cf-b47c79a30112\",\n  \"image\": \"https://pb33f.io/images/t-shirt.png\",\n "+
		" \"name\": \"pb33f t-shirt\",\n  \"price\": 19.99,\n  \"shortCode\": \"pb0001\"\n}", string(b))
}

func TestNewMockEngine_BuildResponse_MissingPath_404(t *testing.T) {
	doc := resetGiftshopState()
	me := NewMockEngine(doc, false, true)

	request, _ := http.NewRequest(http.MethodGet, "https://api.pb33f.io/minky/monkey/moo", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")

	b, status, err := me.GenerateResponse(request)

	assert.Error(t, err)
	assert.Equal(t, 404, status)

	var decoded map[string]any
	_ = json.Unmarshal(b, &decoded)

	assert.Equal(t, "Path / operation not found (404)", decoded["title"])
	assert.Equal(t, "Unable to locate the path '/minky/monkey/moo' with the method 'GET'. "+
		"Error: GET Path '/minky/monkey/moo' not found, Reason: The GET request contains a path of '/minky/monkey/moo' "+
		"however that path, or the GET method for that path does not exist in the specification", decoded["detail"])
}

func TestNewMockEngine_BuildResponse_MissingOperation_404(t *testing.T) {
	doc := resetGiftshopState()
	me := NewMockEngine(doc, false, true)

	request, _ := http.NewRequest(http.MethodPatch, "https://api.pb33f.io/wiretap/giftshop/products", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")

	b, status, err := me.GenerateResponse(request)

	assert.Error(t, err)
	assert.Equal(t, 404, status)

	var decoded map[string]any
	_ = json.Unmarshal(b, &decoded)

	assert.Equal(t, "Path / operation not found (404)", decoded["title"])
	assert.Equal(t, "Unable to locate the path '/wiretap/giftshop/products' with the method 'PATCH'. Error: PATCH Path '/wiretap/giftshop/products' not found, Reason: The PATCH method for that path does not exist in the specification", decoded["detail"])
}

func TestNewMockEngine_BuildResponse_CreateProduct_NoSecurity_Invalid(t *testing.T) {

	doc := resetGiftshopState()
	me := NewMockEngine(doc, false, true)

	product := make(map[string]any)
	product["price"] = 400.23
	product["shortCode"] = "pb0001"
	payload, _ := json.Marshal(product)

	request, _ := http.NewRequest(http.MethodPost, "https://api.pb33f.io/wiretap/giftshop/products", bytes.NewBuffer(payload))
	request.Header.Set(helpers.ContentTypeHeader, "application/json")

	b, status, err := me.GenerateResponse(request)

	assert.Error(t, err)
	assert.Equal(t, 401, status)

	var decoded map[string]any
	_ = json.Unmarshal(b, &decoded)

	assert.Equal(t, "authentication", decoded["code"])
	assert.Equal(t, "This request requires authentication. You are not authenticated.", decoded["message"])

}

func TestNewMockEngine_BuildResponse_CreateProduct_WithSecurity_Invalid(t *testing.T) {

	doc := resetGiftshopState()
	me := NewMockEngine(doc, false, true)

	product := make(map[string]any)
	product["price"] = 400.23
	product["shortCode"] = "pb0001"
	payload, _ := json.Marshal(product)

	request, _ := http.NewRequest(http.MethodPost, "https://api.pb33f.io/wiretap/giftshop/products", bytes.NewBuffer(payload))
	request.Header.Set(helpers.ContentTypeHeader, "application/json")
	request.Header.Set("X-API-Key", "doesnotmatter")

	b, status, err := me.GenerateResponse(request)

	assert.Error(t, err)
	assert.Equal(t, 422, status)

	var decoded map[string]any
	_ = json.Unmarshal(b, &decoded)

	assert.Equal(t, "Invalid request (422)", decoded["title"])
	assert.Len(t, decoded["payload"].([]any), 1)

}

func TestNewMockEngine_BuildResponse_Petstore_Sexurirt(t *testing.T) {
	doc := resetGiftshopState()
	me := NewMockEngine(doc, false, true)

	request, _ := http.NewRequest(http.MethodGet, "https://api.pb33f.io/minky/monkey/moo", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")

	b, status, err := me.GenerateResponse(request)

	assert.Error(t, err)
	assert.Equal(t, 404, status)

	var decoded map[string]any
	_ = json.Unmarshal(b, &decoded)

	assert.Equal(t, "Path / operation not found (404)", decoded["title"])
	assert.Equal(t, "Unable to locate the path '/minky/monkey/moo' with the method 'GET'. "+
		"Error: GET Path '/minky/monkey/moo' not found, Reason: The GET request contains a path of '/minky/monkey/moo' "+
		"however that path, or the GET method for that path does not exist in the specification", decoded["detail"])
}

func TestNewMockEngine_NoLocalSecurity(t *testing.T) {

	spec := `openapi: 3.1.0
info:
  title: Test
  version: 0.1.0
security:
  - apiKeyAuth: []
paths:
  /go:
    get:
      operationId: noAuth
      security:
        - {}
      responses:
        "200":
          description: O
components:
  securitySchemes:
    apiKeyAuth:
      type: apiKey
      in: header
      name: Authorization`

	d, _ := libopenapi.NewDocument([]byte(spec))
	doc, _ := d.BuildV3Model()

	me := NewMockEngine(&doc.Model, false, true)

	request, _ := http.NewRequest(http.MethodGet, "https://api.pb33f.io/go", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")
	request.AddCookie(&http.Cookie{
		Name:  "burger-chips-beer",
		Value: "123",
	})
	path, _ := me.findPath(request)
	operation := me.findOperation(request, path)

	err := me.ValidateSecurity(request, operation)
	assert.NoError(t, err)
}

// https://github.com/pb33f/wiretap/issues/78
func TestNewMockEngine_Fragment(t *testing.T) {

	spec := `openapi: 3.1.0
info:
  title: Test
  version: 0.1.0
security:
  - apiKeyAuth: []
paths:
  /auth#basicAuth:
    post:
      operationId: basicAuth
      security:
        - basicAuth: []
      servers:
        - url: http://localhost:35456
      requestBody:
        content:
          application/json:
            schemas:
              type: object
        required: true
      responses:
        "200":
          description: OK
components:
  securitySchemes:
    basicAuth:
      type: http
      scheme: basic`

	d, _ := libopenapi.NewDocument([]byte(spec))
	doc, _ := d.BuildV3Model()

	me := NewMockEngine(&doc.Model, false, true)

	request, _ := http.NewRequest(http.MethodPost, "https://api.pb33f.io/auth", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")
	path, _ := me.findPath(request)
	operation := me.findOperation(request, path)

	err := me.ValidateSecurity(request, operation)
	assert.Error(t, err)
	assert.Equal(t, "basic authentication failed: bearer token not found, "+
		"no `Authorization` header found in request", err.Error())
}

// https://github.com/pb33f/wiretap/issues/79
func TestNewMockEngine_ContentType(t *testing.T) {

	spec := `openapi: 3.1.0
info:
  title: Test
  version: 0.1.0
paths:
  /auth:
    post:
      operationId: basicAuth
      security:
        - basicAuth: []
      servers:
        - url: http://localhost:35456
      requestBody:
        content:
          application/json:
            schemas:
              type: object
        required: true
      responses:
        "200":
          description: OK
components:
  securitySchemes:
    basicAuth:
      type: http
      scheme: basic`

	d, _ := libopenapi.NewDocument([]byte(spec))
	doc, _ := d.BuildV3Model()

	me := NewMockEngine(&doc.Model, false, true)

	payload := `{"basicAuth":{"password":"testPass","username":"testUser"}}`
	buf := bytes.NewBuffer([]byte(payload))
	request, _ := http.NewRequest(http.MethodPost, "https://api.pb33f.io/auth", buf)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")
	request.Header.Set(helpers.AuthorizationHeader, "Basic dGVzdFVzZXI6dGVzdFBhc3M=")

	b, status, err := me.GenerateResponse(request)
	assert.NoError(t, err)
	assert.Equal(t, 200, status)
	assert.Equal(t, `{"type":"https://pb33f.io/wiretap/errors#empty","title":"Response is empty (200)","status":200,"detail":"Nothing was generated for the request '/auth' with the method 'POST'. Response is empty"}`, string(b))
}

// https://github.com/pb33f/wiretap/issues/80
func TestNewMockEngine_MultiAuth(t *testing.T) {

	spec := `openapi: 3.1.0
info:
  title: Test
  version: 0.1.0
security:
  - xApiKey: []
  - apiKey: []
paths:
  /test:
    get:
      responses:
        '200':
          description: OK
components:
  securitySchemes:
    xApiKey:
      type: apiKey
      in: header
      name: x-api-key
    apiKey:
      type: apiKey
      in: header
      name: Authorization`

	d, _ := libopenapi.NewDocument([]byte(spec))
	doc, _ := d.BuildV3Model()

	me := NewMockEngine(&doc.Model, false, true)

	request, _ := http.NewRequest(http.MethodGet, "https://api.pb33f.io/test", nil)
	request.Header.Set(helpers.AuthorizationHeader, "ding-a-ling")

	path, _ := me.findPath(request)
	operation := me.findOperation(request, path)

	err := me.ValidateSecurity(request, operation)
	assert.NoError(t, err)

	request, _ = http.NewRequest(http.MethodGet, "https://api.pb33f.io/test", nil)
	request.Header.Set("x-api-key", "ding-a-ling")

	path, _ = me.findPath(request)
	operation = me.findOperation(request, path)

	err = me.ValidateSecurity(request, operation)
	assert.NoError(t, err)

}

// https://github.com/pb33f/wiretap/issues/80
func TestNewMockEngine_OptionalAuth(t *testing.T) {

	spec := `openapi: 3.1.0
info:
  title: Test
  version: 0.1.0
security:
  - xApiKey: []
  - {}
paths:
  /test:
    get:
      responses:
        '200':
          description: OK
components:
  securitySchemes:
    xApiKey:
      type: apiKey
      in: header
      name: x-api-key
    apiKey:
      type: apiKey
      in: header
      name: Authorization`

	d, _ := libopenapi.NewDocument([]byte(spec))
	doc, _ := d.BuildV3Model()

	me := NewMockEngine(&doc.Model, false, true)

	request, _ := http.NewRequest(http.MethodGet, "https://api.pb33f.io/test", nil)

	path, _ := me.findPath(request)
	operation := me.findOperation(request, path)

	err := me.ValidateSecurity(request, operation)
	assert.NoError(t, err)

}

// https://github.com/pb33f/wiretap/issues/84
func TestNewMockEngine_UseExamples(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /test:
    get:
      responses:
        '200':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Thing'
              examples:
                happyDays:
                  value:
                    name: happy days
                    description: a terrible show from a time that never existed. 
components:
  schemas:
    Thing:
      type: object
      properties:
        name:
          type: string
          example: nameExample
        description:
          type: string
          example: descriptionExample
`

	d, _ := libopenapi.NewDocument([]byte(spec))
	doc, _ := d.BuildV3Model()

	me := NewMockEngine(&doc.Model, false, true)

	request, _ := http.NewRequest(http.MethodGet, "https://api.pb33f.io/test", nil)

	b, status, err := me.GenerateResponse(request)

	assert.NoError(t, err)
	assert.Equal(t, 200, status)

	var decoded map[string]any
	_ = json.Unmarshal(b, &decoded)

	assert.Equal(t, "happy days", decoded["name"])
	assert.Equal(t, "a terrible show from a time that never existed.", decoded["description"])

}

// https://github.com/pb33f/wiretap/issues/84
func TestNewMockEngine_UseExamples_Preferred(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /test:
    get:
      responses:
        '200':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Thing'
              examples:
                happyDays:
                  value:
                    name: happy days
                    description: a terrible show from a time that never existed.
                robocop:
                  value:
                    name: robocop
                    description: perhaps the best cyberpunk movie ever made.
components:
  schemas:
    Thing:
      type: object
      properties:
        name:
          type: string
          example: nameExample
        description:
          type: string
          example: descriptionExample
`

	d, _ := libopenapi.NewDocument([]byte(spec))
	doc, _ := d.BuildV3Model()

	me := NewMockEngine(&doc.Model, false, true)

	request, _ := http.NewRequest(http.MethodGet, "https://api.pb33f.io/test", nil)
	request.Header.Set(helpers.Preferred, "robocop")

	b, status, err := me.GenerateResponse(request)

	assert.NoError(t, err)
	assert.Equal(t, 200, status)

	var decoded map[string]any
	_ = json.Unmarshal(b, &decoded)

	assert.Equal(t, "robocop", decoded["name"])
	assert.Equal(t, "perhaps the best cyberpunk movie ever made.", decoded["description"])

}

// https://github.com/pb33f/wiretap/issues/84
func TestNewMockEngine_UseExamples_FromSchemaExamples(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /test:
    get:
      responses:
        '200':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Thing'
components:
  schemas:
    Thing:
      type: object
      examples:
        - name: happy days
          description: a terrible show from a time that never existed.
      properties:
        name:
          type: string
          example: nameExample
        description:
          type: string
          example: descriptionExample
`

	d, _ := libopenapi.NewDocument([]byte(spec))
	doc, _ := d.BuildV3Model()

	me := NewMockEngine(&doc.Model, false, true)

	request, _ := http.NewRequest(http.MethodGet, "https://api.pb33f.io/test", nil)

	b, status, err := me.GenerateResponse(request)

	assert.NoError(t, err)
	assert.Equal(t, 200, status)

	var decoded map[string]any
	_ = json.Unmarshal(b, &decoded)

	assert.Equal(t, "happy days", decoded["name"])
	assert.Equal(t, "a terrible show from a time that never existed.", decoded["description"])

}

// https://github.com/pb33f/wiretap/issues/84
func TestNewMockEngine_UseExamples_FromSchemaExamples_Preferred(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /test:
    get:
      responses:
        '200':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Thing'
components:
  schemas:
    Thing:
      type: object
      examples:
        - name: happy days
          description: a terrible show from a time that never existed.
        - name: robocop
          description: perhaps the best cyberpunk movie ever made.
      properties:
        name:
          type: string
          example: nameExample
        description:
          type: string
          example: descriptionExample
`

	d, _ := libopenapi.NewDocument([]byte(spec))
	doc, _ := d.BuildV3Model()

	me := NewMockEngine(&doc.Model, false, true)

	request, _ := http.NewRequest(http.MethodGet, "https://api.pb33f.io/test", nil)
	request.Header.Set(helpers.Preferred, "1")

	b, status, err := me.GenerateResponse(request)

	assert.NoError(t, err)
	assert.Equal(t, 200, status)

	var decoded map[string]any
	_ = json.Unmarshal(b, &decoded)

	assert.Equal(t, "robocop", decoded["name"])
	assert.Equal(t, "perhaps the best cyberpunk movie ever made.", decoded["description"])

}

// https://github.com/pb33f/wiretap/issues/84
func TestNewMockEngine_UseExamples_FromSchema(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /test:
    get:
      responses:
        '200':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Thing'
components:
  schemas:
    Thing:
      type: object
      properties:
        name:
          type: string
          example: nameExample
        description:
          type: string
          example: descriptionExample
`

	d, _ := libopenapi.NewDocument([]byte(spec))
	doc, _ := d.BuildV3Model()

	me := NewMockEngine(&doc.Model, false, true)

	request, _ := http.NewRequest(http.MethodGet, "https://api.pb33f.io/test", nil)

	b, status, err := me.GenerateResponse(request)

	assert.NoError(t, err)
	assert.Equal(t, 200, status)

	var decoded map[string]any
	_ = json.Unmarshal(b, &decoded)

	assert.Equal(t, "nameExample", decoded["name"])
	assert.Equal(t, "descriptionExample", decoded["description"])

}

// https://github.com/pb33f/wiretap/issues/84
func TestNewMockEngine_UseExamples_FromSchema_Generated(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /test:
    get:
      responses:
        '200':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Thing'
components:
  schemas:
    Thing:
      type: object
      properties:
        name:
          type: string
        description:
          type: string
`

	d, _ := libopenapi.NewDocument([]byte(spec))
	doc, _ := d.BuildV3Model()

	me := NewMockEngine(&doc.Model, false, true)

	request, _ := http.NewRequest(http.MethodGet, "https://api.pb33f.io/test", nil)

	b, status, err := me.GenerateResponse(request)

	assert.NoError(t, err)
	assert.Equal(t, 200, status)

	var decoded map[string]any
	_ = json.Unmarshal(b, &decoded)

	assert.NotEmpty(t, decoded["name"])
	assert.NotEmpty(t, decoded["description"])

}

func TestNewMockEngine_UseExamples_Preferred_From_400(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /test:
    get:
      responses:
        '200':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Thing'
              examples:
                happyDays:
                  value:
                    name: happy days
                    description: a terrible show from a time that never existed.
                robocop:
                  value:
                    name: robocop
                    description: perhaps the best cyberpunk movie ever made.
        '400':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorThing'
              examples:
                sadErrorDays:
                  value:
                    name: sad error days
                    description: a sad error prone show
                sadcop:
                  value:
                    name: sad cop
                    description: perhaps the saddest cyberpunk movie ever made.
components:
  schemas:
    Thing:
      type: object
      properties:
        name:
          type: string
          example: nameExample
        description:
          type: string
          example: descriptionExample
    ErrorThing:
      type: object
      properties:
        name:
          type: string
          example: errorNameExample
        description:
          type: string
          example: errorDescriptionExample
`

	d, _ := libopenapi.NewDocument([]byte(spec))
	doc, _ := d.BuildV3Model()

	me := NewMockEngine(&doc.Model, false, true)

	request, _ := http.NewRequest(http.MethodGet, "https://api.pb33f.io/test", nil)
	request.Header.Set(helpers.Preferred, "sadcop")

	b, status, err := me.GenerateResponse(request)

	assert.NoError(t, err)
	assert.Equal(t, 400, status)

	var decoded map[string]any
	_ = json.Unmarshal(b, &decoded)

	assert.Equal(t, "sad cop", decoded["name"])
	assert.Equal(t, "perhaps the saddest cyberpunk movie ever made.", decoded["description"])
}

func TestNewMockEngine_UseExamples_Preferred_200_Not_Json(t *testing.T) {
	// A little far-fetched for an API to behave this way,
	// where lowest 2xx response is html and second is json,
	// including the test case just in case
	spec := `openapi: 3.1.0
paths:
  /test:
    get:
      responses:
        '200':
          content:
            text/html:
              schema:
                $ref: '#/components/schemas/HtmlThing'
              examples:
                happyHtmlDays:
                  value: <!DOCTYPE html><html lang="en"><body><h1>Happy Days</h1</body></html>
                robocopInHtml:
                  value: <!DOCTYPE html><html lang="en"><body><h1>Robo cop</h1</body></html>
        '202':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Thing'
              examples:
                happyDays:
                  value:
                    name: happy days
                    description: a terrible show from a time that never existed.
                robocop:
                  value:
                    name: robocop
                    description: perhaps the best cyberpunk movie ever made.
        '400':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorThing'
              examples:
                sadErrorDays:
                  value:
                    name: sad error days
                    description: a sad error prone show
                sadcop:
                  value:
                    name: sad cop
                    description: perhaps the saddest cyberpunk movie ever made.
components:
  schemas:
    Thing:
      type: object
      properties:
        name:
          type: string
          example: nameExample
        description:
          type: string
          example: descriptionExample
    HtmlThing:
      type: string
    ErrorThing:
      type: object
      properties:
        name:
          type: string
          example: errorNameExample
        description:
          type: string
          example: errorDescriptionExample
`

	d, _ := libopenapi.NewDocument([]byte(spec))
	doc, _ := d.BuildV3Model()

	me := NewMockEngine(&doc.Model, false, true)

	// Check that we don't panic if first 2xx does not match media type
	request, _ := http.NewRequest(http.MethodGet, "https://api.pb33f.io/test", nil)
	request.Header.Set(helpers.Preferred, "robocop")

	b, status, err := me.GenerateResponse(request)

	assert.NoError(t, err)
	assert.Equal(t, 202, status)

	var decoded map[string]any
	_ = json.Unmarshal(b, &decoded)

	assert.Equal(t, "robocop", decoded["name"])
	assert.Equal(t, "perhaps the best cyberpunk movie ever made.", decoded["description"])

	// Now see if html will work with preferred header for second html example
	request, _ = http.NewRequest(http.MethodGet, "https://api.pb33f.io/test", nil)
	request.Header.Set(helpers.Preferred, "robocopInHtml")
	request.Header.Set("Content-Type", "text/html")

	b, status, err = me.GenerateResponse(request)

	assert.NoError(t, err)
	assert.Equal(t, 200, status)
	assert.Equal(t, "<!DOCTYPE html><html lang=\"en\"><body><h1>Robo cop</h1</body></html>", string(b[:]))

	// Now see if html will work w/ Accept header and no preferred
	request, _ = http.NewRequest(http.MethodGet, "https://api.pb33f.io/test", nil)
	request.Header.Set("Accept", "text/html")

	b, status, err = me.GenerateResponse(request)

	assert.NoError(t, err)
	assert.Equal(t, 200, status)
	assert.Equal(t, "<!DOCTYPE html><html lang=\"en\"><body><h1>Happy Days</h1</body></html>", string(b[:]))
}

// https://github.com/pb33f/wiretap/issues/83
func TestNewMockEngine_UseExamples_Items_Issue83(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /chip-shop:
    get:
      operationId: itemsExamples
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  args:
                    type: object
                    properties:
                      arrParam:
                        type: string
                        example: "test,test2"
                      arrParamExploded:
                        type: array
                        items:
                          type: string
                          examples:
                            - "1"
                            - "2"
                    required:
                      - arrParam
                      - arrParamExploded
                required:
                  - args
`

	d, _ := libopenapi.NewDocument([]byte(spec))
	doc, _ := d.BuildV3Model()

	me := NewMockEngine(&doc.Model, false, true)

	request, _ := http.NewRequest(http.MethodGet, "https://api.pb33f.io/chip-shop", nil)

	b, status, err := me.GenerateResponse(request)

	assert.NoError(t, err)
	assert.Equal(t, 200, status)

	var decoded map[string]any
	_ = json.Unmarshal(b, &decoded)

	args := decoded["args"].(map[string]any)

	assert.Equal(t, "test,test2", args["arrParam"])

	items := args["arrParamExploded"].([]any)
	assert.Equal(t, "1", items[0])
	assert.Equal(t, "2", items[1])

}

// https://github.com/pb33f/wiretap/issues/83
func TestNewMockEngine_UseExample_Items_Issue83(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /chip-shop:
    get:
      operationId: itemsExamples
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  args:
                    type: object
                    properties:
                      arrParam:
                        type: string
                        example: "test,test2"
                      arrParamExploded:
                        type: array
                        items:
                          type: string
                          example: "1"
                    required:
                      - arrParam
                      - arrParamExploded
                required:
                  - args
`

	d, _ := libopenapi.NewDocument([]byte(spec))
	doc, _ := d.BuildV3Model()

	me := NewMockEngine(&doc.Model, false, true)

	request, _ := http.NewRequest(http.MethodGet, "https://api.pb33f.io/chip-shop", nil)

	b, status, err := me.GenerateResponse(request)

	assert.NoError(t, err)
	assert.Equal(t, 200, status)

	var decoded map[string]any
	_ = json.Unmarshal(b, &decoded)

	args := decoded["args"].(map[string]any)

	assert.Equal(t, "test,test2", args["arrParam"])

	items := args["arrParamExploded"].([]any)
	assert.Equal(t, "1", items[0])

}

// https://github.com/pb33f/wiretap/issues/89
func TestNewMockEngine_OverrideStatusCode_Issue89(t *testing.T) {

	spec := `openapi: 3.1.0
paths:
  /test:
    get:
      responses:
        '200':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Thing'
components:
  schemas:
    Thing:
      type: object
      properties:
        name:
          type: string
        description:
          type: string
`

	d, _ := libopenapi.NewDocument([]byte(spec))
	doc, _ := d.BuildV3Model()

	me := NewMockEngine(&doc.Model, false, true)

	request, _ := http.NewRequest(http.MethodGet, "https://api.pb33f.io/test", nil)
	request.Header.Set("wiretap-status-code", "418")

	b, status, err := me.GenerateResponse(request)

	assert.NoError(t, err)
	assert.Equal(t, 418, status)

	var decoded map[string]any
	_ = json.Unmarshal(b, &decoded)

	assert.NotEmpty(t, decoded["name"])
	assert.NotEmpty(t, decoded["description"])

}

func TestNewMockEngine_UseExamples_204_Response(t *testing.T) {
	spec := `openapi: 3.1.0
info:
  title: Simple API
  version: 1.0.0
paths:
  /test:
    post:
      summary: Submit a request
      description: Submits a request to the server which processes it but returns no content.
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                message:
                  type: string
                  description: Message to be processed by the server.
                  example: "Hello, this is a test message."
      responses:
        '204':
          description: Request processed successfully, no content to return.
`

	d, _ := libopenapi.NewDocument([]byte(spec))
	doc, _ := d.BuildV3Model()

	me := NewMockEngine(&doc.Model, false, true)

	request, err := http.NewRequest(http.MethodPost, "https://api.pb33f.io/test",
		bytes.NewBufferString("{\"message\": \"Hello, this is a test message.\"}"))
	require.NoError(t, err)

	request.Header.Set("Content-Type", "application/json")

	b, status, err := me.GenerateResponse(request)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNoContent, status)
	assert.Empty(t, b)
}

func TestNewMockEngine_GenerateResponse_CombinedExampleObject(t *testing.T) {
	spec := `openapi: 3.0.3
info:
  title: Example API
  description: An example API for testing purposes
  version: 1.0.0
paths:
  /examples:
    get:
      summary: Get example data
      description: Retrieve an example response with various fields
      responses:
        '200':
          description: Ok
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Response'
components:
  schemas:
    ID:
      type: integer

    Response:
      type: object
      required:
        - id
        - name
      properties:
        id:
          $ref: '#/components/schemas/ID'
        name:
          type: string
        username:
          type: string
        active:
          type: boolean
        balance:
          type: number
          format: float
        tags:
          type: array
          items:
            type: string
      example:
        id: 123
        name: "John Doe"
        username: "jack"
        active: true
        balance: 99.99
        tags: ["tag1", "tag2", "tag3"]`

	d, _ := libopenapi.NewDocument([]byte(spec))
	doc, _ := d.BuildV3Model()

	me := NewMockEngine(&doc.Model, false, true)

	request, err := http.NewRequest(http.MethodGet, "https://api.pb33f.io/examples", http.NoBody)
	require.NoError(t, err)

	b, status, err := me.GenerateResponse(request)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, `{"active":true,"balance":99.99,"id":123,"name":"John Doe","tags":["tag1","tag2","tag3"],"username":"jack"}`, string(b))
}

func TestNewMockEngine_GenerateResponse_DefaultWithAllPropertyExamplesInResponse(t *testing.T) {
	spec := `openapi: 3.0.3
info:
  title: Example API
  description: An example API for testing purposes
  version: 1.0.0
paths:
  /examples:
    get:
      summary: Get example data
      description: Retrieve an example response with various fields
      responses:
        '200':
          description: Ok
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Response'
components:
  schemas:
    ID:
      type: integer
      example: 123

    Response:
      type: object
      required:
        - id
        - name
      properties:
        id:
          $ref: '#/components/schemas/ID'
        name:
          type: string
          example: "John Doe"
        username:
          type: string
          example: "jack"
        active:
          type: boolean
          example: true
        balance:
          type: number
          format: float
          example: 99.99
        tags:
          type: array
          items:
            type: string
          example: ["tag1", "tag2", "tag3"]`

	d, _ := libopenapi.NewDocument([]byte(spec))
	doc, _ := d.BuildV3Model()

	me := NewMockEngine(&doc.Model, false, true)

	request, err := http.NewRequest(http.MethodGet, "https://api.pb33f.io/examples", http.NoBody)
	require.NoError(t, err)

	b, status, err := me.GenerateResponse(request)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, `{"active":true,"balance":99.99,"id":123,"name":"John Doe","tags":["tag1","tag2","tag3"],"username":"jack"}`, string(b))
}

func TestNewMockEngine_GenerateResponse_OnlyRequiredPropertyExamplesInResponse(t *testing.T) {
	spec := `openapi: 3.0.3
info:
  title: Example API
  description: An example API for testing purposes
  version: 1.0.0
paths:
  /examples:
    get:
      summary: Get example data
      description: Retrieve an example response with various fields
      responses:
        '200':
          description: Ok
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Response'
components:
  schemas:
    ID:
      type: integer
      example: 123

    Response:
      type: object
      required:
        - id
        - name
      properties:
        id:
          $ref: '#/components/schemas/ID'
        name:
          type: string
          example: "John Doe"
        username:
          type: string
          example: "jack"
        active:
          type: boolean
          example: true
        balance:
          type: number
          format: float
          example: 99.99
        tags:
          type: array
          items:
            type: string
          example: ["tag1", "tag2", "tag3"]`

	d, _ := libopenapi.NewDocument([]byte(spec))
	doc, _ := d.BuildV3Model()

	me := NewMockEngine(&doc.Model, false, false)

	request, err := http.NewRequest(http.MethodGet, "https://api.pb33f.io/examples", http.NoBody)
	require.NoError(t, err)

	b, status, err := me.GenerateResponse(request)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, `{"id":123,"name":"John Doe"}`, string(b))
}

func loadStrictTestSpec(t *testing.T) *v3.Document {
	spec, err := io.ReadAll(mustOpen(t, "../testdata/strict_mode_test.yaml"))
	if err != nil {
		t.Fatalf("failed to read strict mode test spec: %v", err)
	}
	d, err := libopenapi.NewDocument(spec)
	if err != nil {
		t.Fatalf("failed to parse strict mode test spec: %v", err)
	}
	compiled, err := d.BuildV3Model()
	if err != nil {
		t.Fatalf("failed to build v3 model: %v", err)
	}
	return &compiled.Model
}

func mustOpen(t *testing.T, path string) io.Reader {
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open file %s: %v", path, err)
	}
	return f
}

func TestNewMockEngineStrict_RejectsUndeclaredProperty(t *testing.T) {
	doc := loadStrictTestSpec(t)
	engine := NewStrictMockEngine(doc, false, true)

	// POST /users with undeclared property "extra"
	body := bytes.NewBufferString(`{"name": "test", "extra": "undeclared"}`)
	request, _ := http.NewRequest(http.MethodPost, "http://localhost/users", body)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")

	b, statusCode, err := engine.GenerateResponse(request)

	// Assert: GenerateResponse returns 422 for validation failures
	assert.Equal(t, 422, statusCode)
	assert.NotNil(t, err)

	// Assert body contains validation error payload
	var decoded map[string]any
	_ = json.Unmarshal(b, &decoded)
	assert.Equal(t, "Invalid request (422)", decoded["title"])
}

func TestNewMockEngine_AllowsUndeclaredProperty(t *testing.T) {
	doc := loadStrictTestSpec(t)
	engine := NewMockEngine(doc, false, true)

	// POST /users with undeclared property "extra" - should be allowed in non-strict mode
	body := bytes.NewBufferString(`{"name": "test", "extra": "undeclared"}`)
	request, _ := http.NewRequest(http.MethodPost, "http://localhost/users", body)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")

	_, statusCode, err := engine.GenerateResponse(request)

	// Assert: Returns 2xx success, no validation error
	assert.Equal(t, 201, statusCode)
	assert.NoError(t, err)
}

// Issue #109 — bypass mock-mode request validation so clients driving wiretap
// from fixtures that deliberately send malformed payloads still receive the
// example they asked for via Preferred / wiretap-status-code.

const bypassValidationSpec = `openapi: 3.1.0
info:
  title: Bypass test
  version: 1.0.0
paths:
  /things:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [name, count]
              properties:
                name: {type: string}
                count: {type: integer}
      responses:
        '201':
          content:
            application/json:
              schema:
                type: object
                properties:
                  id: {type: string}
              examples:
                created:
                  value: {id: "ok-123"}
        '422':
          content:
            application/json:
              schema:
                type: object
                properties:
                  error: {type: string}
              examples:
                badInput:
                  value: {error: "bad input demo"}
`

func bypassTestEngine(t *testing.T, hardValidation, bypassValidation bool) *ResponseMockEngine {
	t.Helper()
	d, err := libopenapi.NewDocument([]byte(bypassValidationSpec))
	require.NoError(t, err)
	doc, errs := d.BuildV3Model()
	require.Empty(t, errs)
	return NewMockEngineWithConfig(&doc.Model, false, true, false, hardValidation, bypassValidation)
}

func invalidPost(t *testing.T) *http.Request {
	t.Helper()
	// missing both required fields + wrong type on `count`
	req, _ := http.NewRequest(http.MethodPost, "https://api.pb33f.io/things",
		bytes.NewBufferString(`{"count": "nope"}`))
	req.Header.Set(helpers.ContentTypeHeader, "application/json")
	return req
}

// Regression guard — the default path must keep returning the 422 validation
// error body when hard-validation is on and nothing signals a bypass.
func TestMockBypass_HardValidation_NoBypass_ReturnsValidationError(t *testing.T) {
	me := bypassTestEngine(t, true, false)
	req := invalidPost(t)

	b, status, err := me.GenerateResponse(req)
	assert.Error(t, err)
	assert.Equal(t, 422, status)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(b, &decoded))
	assert.Equal(t, "Invalid request (422)", decoded["title"])
}

// Either the per-request header or the engine-level flag engages the bypass,
// and the Preferred example at the wiretap-status-code is returned instead of
// the validation-error body.
func TestMockBypass_ReturnsPreferredExample(t *testing.T) {
	tests := []struct {
		name         string
		engineBypass bool
		headerBypass bool
	}{
		{"per-request header", false, true},
		{"global engine flag", true, false},
		{"both signals set", true, true},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			me := bypassTestEngine(t, true, tc.engineBypass)
			req := invalidPost(t)
			req.Header.Set(helpers.Preferred, "badInput")
			req.Header.Set("wiretap-status-code", "422")
			if tc.headerBypass {
				req.Header.Set("wiretap-bypass-validation", "true")
			}

			b, status, err := me.GenerateResponse(req)
			assert.NoError(t, err)
			assert.Equal(t, 422, status)

			var decoded map[string]any
			require.NoError(t, json.Unmarshal(b, &decoded))
			assert.Equal(t, "bad input demo", decoded["error"])
		})
	}
}

// Header parsing is case-insensitive so CI tooling isn't required to normalise.
func TestMockBypass_HeaderCaseInsensitive(t *testing.T) {
	for _, v := range []string{"true", "True", "TRUE", "tRuE"} {
		v := v
		t.Run(v, func(t *testing.T) {
			me := bypassTestEngine(t, true, false)
			req := invalidPost(t)
			req.Header.Set(helpers.Preferred, "badInput")
			req.Header.Set("wiretap-status-code", "422")
			req.Header.Set("wiretap-bypass-validation", v)

			_, status, err := me.GenerateResponse(req)
			assert.NoError(t, err)
			assert.Equal(t, 422, status, "header value %q should engage bypass", v)
		})
	}
}

// Any value other than "true" (in any case) leaves the gate intact.
func TestMockBypass_HeaderOtherValuesDoNotBypass(t *testing.T) {
	for _, v := range []string{"false", "no", "0", "", "yes"} {
		v := v
		t.Run("value="+v, func(t *testing.T) {
			me := bypassTestEngine(t, true, false)
			req := invalidPost(t)
			if v != "" {
				req.Header.Set("wiretap-bypass-validation", v)
			}

			_, status, err := me.GenerateResponse(req)
			assert.Error(t, err)
			assert.Equal(t, 422, status, "header value %q must not engage bypass", v)
		})
	}
}

// Bypass without Preferred falls through to the lowest-success-code example
// rather than the validation-error body.
func TestMockBypass_WithoutPreferredFallsThroughToSuccess(t *testing.T) {
	me := bypassTestEngine(t, true, true)
	req := invalidPost(t)

	b, status, err := me.GenerateResponse(req)
	assert.NoError(t, err)
	assert.Equal(t, 201, status)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(b, &decoded))
	assert.Equal(t, "ok-123", decoded["id"])
}

// No regression on valid requests — the bypass header is a no-op.
func TestMockBypass_HeaderOnValidRequest_IsNoOp(t *testing.T) {
	me := bypassTestEngine(t, true, false)
	validBody := bytes.NewBufferString(`{"name":"widget","count":3}`)
	req, _ := http.NewRequest(http.MethodPost, "https://api.pb33f.io/things", validBody)
	req.Header.Set(helpers.ContentTypeHeader, "application/json")
	req.Header.Set("wiretap-bypass-validation", "true")

	_, status, err := me.GenerateResponse(req)
	assert.NoError(t, err)
	assert.Equal(t, 201, status)
}

// The non-hard-validation path already serves examples regardless of request
// validity. Bypass must not change that — it's a no-op on this branch.
func TestMockBypass_NonHardValidation_IsNoOp(t *testing.T) {
	for _, bypass := range []bool{false, true} {
		bypass := bypass
		t.Run("bypass="+strconv.FormatBool(bypass), func(t *testing.T) {
			me := bypassTestEngine(t, false, bypass)
			req := invalidPost(t)

			_, status, err := me.GenerateResponse(req)
			assert.NoError(t, err)
			assert.Equal(t, 201, status, "non-hard-validation path must serve the example; bypass=%v", bypass)
		})
	}
}

