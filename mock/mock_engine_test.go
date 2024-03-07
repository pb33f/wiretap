// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package mock

import (
	"bytes"
	"encoding/json"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi-validator/helpers"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"
)

// var doc *v3.Document
var giftshopBytes []byte
var petstoreBytes []byte

func resetGiftshopState() *v3.Document {
	if len(giftshopBytes) <= 0 {
		resp, err := http.Get("https://api.pb33f.io/wiretap/giftshop-openapi.yaml")
		if err != nil {
			panic(err)
		}
		giftshopBytes, _ = io.ReadAll(resp.Body)
	}
	d, _ := libopenapi.NewDocument(giftshopBytes)
	compiled, _ := d.BuildV3Model()
	return &compiled.Model
}

func resetPetstoreState() *v3.Document {
	if len(petstoreBytes) <= 0 {
		resp, err := http.Get("https://raw.githubusercontent.com/swagger-api/swagger-petstore/master/src/main/resources/openapi.yaml")
		if err != nil {
			panic(err)
		}
		petstoreBytes, _ = io.ReadAll(resp.Body)
	}
	d, _ := libopenapi.NewDocument(petstoreBytes)
	compiled, _ := d.BuildV3Model()
	return &compiled.Model
}

func TestNewMockEngine_findPath(t *testing.T) {
	doc := resetGiftshopState()
	me := NewMockEngine(doc, false)

	request, _ := http.NewRequest(http.MethodGet, "https://api.pb33f.io/wiretap/giftshop/products", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")

	path, _ := me.findPath(request)
	assert.NotNil(t, path)
}

func TestNewMockEngine_findPathNegative(t *testing.T) {
	doc := resetGiftshopState()
	me := NewMockEngine(doc, false)

	request, _ := http.NewRequest(http.MethodGet, "https://api.pb33f.io/wiretap/giftshop/invalid", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")

	path, errors := me.findPath(request)
	assert.Nil(t, path)
	assert.Error(t, errors)
}

func TestNewMockEngine_findOperation(t *testing.T) {
	doc := resetGiftshopState()
	me := NewMockEngine(doc, false)

	request, _ := http.NewRequest(http.MethodGet, "https://api.pb33f.io/wiretap/giftshop/products", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")

	path, _ := me.findPath(request)
	operation := me.findOperation(request, path)
	assert.NotNil(t, operation)

}

func TestNewMockEngine_findOperationNegative(t *testing.T) {
	doc := resetGiftshopState()
	me := NewMockEngine(doc, false)

	request, _ := http.NewRequest(http.MethodPatch, "https://api.pb33f.io/wiretap/giftshop/products", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")

	path, _ := me.findPath(request)
	operation := me.findOperation(request, path)
	assert.Nil(t, operation)
}

func TestNewMockEngine_ValidateSecurity_FailAPIKey_Header(t *testing.T) {
	doc := resetGiftshopState()
	me := NewMockEngine(doc, false)

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
	me := NewMockEngine(doc, false)

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
	me := NewMockEngine(doc, false)

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
	me := NewMockEngine(doc, false)

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
	me := NewMockEngine(doc, false)

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
	me := NewMockEngine(doc, false)

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
	me := NewMockEngine(doc, false)

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
	me := NewMockEngine(doc, false)

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
	me := NewMockEngine(doc, false)

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
//	me := NewMockEngine(doc, false)
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
	me := NewMockEngine(doc, true)

	request, _ := http.NewRequest(http.MethodGet, "https://api.pb33f.io/wiretap/giftshop/products/pb0001", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")

	b, status, err := me.GenerateResponse(request)

	assert.NoError(t, err)
	assert.Equal(t, 200, status)
	assert.Equal(t, "{\n  \"category\": \"shirts\",\n  \"description\": \"A t-shirt with the pb33f logo "+
		"on it.\",\n  \"id\": \"d1404c5c-69bd-4cd2-a4cf-b47c79a30112\",\n  \"image\": \"https://pb33f.io/images/t-shirt.png\",\n "+
		" \"name\": \"pb33f t-shirt\",\n  \"price\": 19.99,\n  \"shortCode\": \"pb0001\"\n}", string(b))
}

func TestNewMockEngine_BuildResponse_MissingPath_404(t *testing.T) {
	doc := resetGiftshopState()
	me := NewMockEngine(doc, false)

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
	me := NewMockEngine(doc, false)

	request, _ := http.NewRequest(http.MethodPatch, "https://api.pb33f.io/wiretap/giftshop/products", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")

	b, status, err := me.GenerateResponse(request)

	assert.Error(t, err)
	assert.Equal(t, 404, status)

	var decoded map[string]any
	_ = json.Unmarshal(b, &decoded)

	assert.Equal(t, "Path / operation not found (404)", decoded["title"])
	assert.Equal(t, "Unable to locate the path '/wiretap/giftshop/products' with the method 'PATCH'. "+
		"Error: PATCH Path '/wiretap/giftshop/products' not found, Reason: The PATCH request contains a path of "+
		"'/wiretap/giftshop/products' however that path, or the PATCH method for that path does not exist in the "+
		"specification", decoded["detail"])
}

func TestNewMockEngine_BuildResponse_CreateProduct_NoSecurity_Invalid(t *testing.T) {

	doc := resetGiftshopState()
	me := NewMockEngine(doc, false)

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
	me := NewMockEngine(doc, false)

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
	me := NewMockEngine(doc, false)

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

	me := NewMockEngine(&doc.Model, false)

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

	me := NewMockEngine(&doc.Model, false)

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

	me := NewMockEngine(&doc.Model, false)

	payload := `{"basicAuth":{"password":"testPass","username":"testUser"}}`
	buf := bytes.NewBuffer([]byte(payload))
	request, _ := http.NewRequest(http.MethodPost, "https://api.pb33f.io/auth", buf)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")
	request.Header.Set(helpers.AuthorizationHeader, "the science man")

	b, status, err := me.GenerateResponse(request)
	assert.NoError(t, err)
	assert.Equal(t, 200, status)
	assert.Equal(t, "", string(b))
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

	me := NewMockEngine(&doc.Model, false)

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

	me := NewMockEngine(&doc.Model, false)

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

	me := NewMockEngine(&doc.Model, false)

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

	me := NewMockEngine(&doc.Model, false)

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

	me := NewMockEngine(&doc.Model, false)

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

	me := NewMockEngine(&doc.Model, false)

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

	me := NewMockEngine(&doc.Model, false)

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

	me := NewMockEngine(&doc.Model, false)

	request, _ := http.NewRequest(http.MethodGet, "https://api.pb33f.io/test", nil)

	b, status, err := me.GenerateResponse(request)

	assert.NoError(t, err)
	assert.Equal(t, 200, status)

	var decoded map[string]any
	_ = json.Unmarshal(b, &decoded)

	assert.NotEmpty(t, decoded["name"])
	assert.NotEmpty(t, decoded["description"])

}
