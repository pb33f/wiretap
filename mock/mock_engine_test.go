// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

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
var specBytes []byte

func resetState() *v3.Document {
	if len(specBytes) <= 0 {
		resp, err := http.Get("https://api.pb33f.io/wiretap/giftshop-openapi.yaml")
		if err != nil {
			panic(err)
		}
		specBytes, _ = io.ReadAll(resp.Body)
	}
	d, _ := libopenapi.NewDocument(specBytes)
	compiled, _ := d.BuildV3Model()
	return &compiled.Model
}

func TestNewMockEngine_findPath(t *testing.T) {
	doc := resetState()
	me := NewMockEngine(doc, false)

	request, _ := http.NewRequest(http.MethodGet, "https://api.pb33f.io/wiretap/giftshop/products", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")

	path, _ := me.findPath(request)
	assert.NotNil(t, path)
}

func TestNewMockEngine_findPathNegative(t *testing.T) {
	doc := resetState()
	me := NewMockEngine(doc, false)

	request, _ := http.NewRequest(http.MethodGet, "https://api.pb33f.io/wiretap/giftshop/invalid", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")

	path, errors := me.findPath(request)
	assert.Nil(t, path)
	assert.Error(t, errors)
}

func TestNewMockEngine_findOperation(t *testing.T) {
	doc := resetState()
	me := NewMockEngine(doc, false)

	request, _ := http.NewRequest(http.MethodGet, "https://api.pb33f.io/wiretap/giftshop/products", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")

	path, _ := me.findPath(request)
	operation := me.findOperation(request, path)
	assert.NotNil(t, operation)

}

func TestNewMockEngine_findOperationNegative(t *testing.T) {
	doc := resetState()
	me := NewMockEngine(doc, false)

	request, _ := http.NewRequest(http.MethodPatch, "https://api.pb33f.io/wiretap/giftshop/products", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")

	path, _ := me.findPath(request)
	operation := me.findOperation(request, path)
	assert.Nil(t, operation)
}

func TestNewMockEngine_ValidateSecurity_FailAPIKey_Header(t *testing.T) {
	doc := resetState()
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
	doc := resetState()
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
	doc := resetState()
	me := NewMockEngine(doc, false)

	request, _ := http.NewRequest(http.MethodPost, "https://api.pb33f.io/wiretap/giftshop/products", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")
	path, _ := me.findPath(request)
	operation := me.findOperation(request, path)

	// mutate securityScheme to be a query param
	doc.Components.SecuritySchemes["ApiKeyAuth"].In = "query"
	doc.Components.SecuritySchemes["ApiKeyAuth"].Name = "pizza-cake-burger"

	err := me.ValidateSecurity(request, operation)
	assert.Error(t, err)
	assert.Equal(t, "apiKey not found, no `pizza-cake-burger` query parameter found in request", err.Error())
}

func TestNewMockEngine_ValidateSecurity_PassAPIKey_Query(t *testing.T) {
	doc := resetState()
	me := NewMockEngine(doc, false)

	request, _ := http.NewRequest(http.MethodPost, "https://api.pb33f.io/wiretap/giftshop/products?pizza-burger-cake=123", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")
	path, _ := me.findPath(request)
	operation := me.findOperation(request, path)

	// mutate securityScheme to be a query param
	doc.Components.SecuritySchemes["ApiKeyAuth"].In = "query"
	doc.Components.SecuritySchemes["ApiKeyAuth"].Name = "pizza-burger-cake"

	err := me.ValidateSecurity(request, operation)
	assert.NoError(t, err)
}

func TestNewMockEngine_ValidateSecurity_FailAPIKey_Cookie(t *testing.T) {
	doc := resetState()
	me := NewMockEngine(doc, false)

	request, _ := http.NewRequest(http.MethodPost, "https://api.pb33f.io/wiretap/giftshop/products", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")
	path, _ := me.findPath(request)
	operation := me.findOperation(request, path)

	// mutate securityScheme to be a query param
	doc.Components.SecuritySchemes["ApiKeyAuth"].In = "cookie"
	doc.Components.SecuritySchemes["ApiKeyAuth"].Name = "burger-chips-beer"

	err := me.ValidateSecurity(request, operation)
	assert.Error(t, err)
	assert.Equal(t, "apiKey not found, no `burger-chips-beer` cookie found in request", err.Error())
}

func TestNewMockEngine_ValidateSecurity_PassAPIKey_Cookie(t *testing.T) {
	doc := resetState()
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
	doc.Components.SecuritySchemes["ApiKeyAuth"].In = "cookie"
	doc.Components.SecuritySchemes["ApiKeyAuth"].Name = "burger-chips-beer"

	err := me.ValidateSecurity(request, operation)
	assert.NoError(t, err)
}

func TestNewMockEngine_ValidateSecurity_FailHTTP_Bearer(t *testing.T) {
	doc := resetState()
	me := NewMockEngine(doc, false)

	request, _ := http.NewRequest(http.MethodPost, "https://api.pb33f.io/wiretap/giftshop/products", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")

	path, _ := me.findPath(request)
	operation := me.findOperation(request, path)

	// mutate securityScheme to be a query param
	doc.Components.SecuritySchemes["ApiKeyAuth"].Type = "http"
	doc.Components.SecuritySchemes["ApiKeyAuth"].Scheme = "bearer"

	err := me.ValidateSecurity(request, operation)
	assert.Error(t, err)
	assert.Equal(t, "bearer token not found, no `Authorization` header found in request", err.Error())
}

func TestNewMockEngine_ValidateSecurity_PassHTTP_Bearer(t *testing.T) {
	doc := resetState()
	me := NewMockEngine(doc, false)

	request, _ := http.NewRequest(http.MethodPost, "https://api.pb33f.io/wiretap/giftshop/products", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")
	request.Header.Set("Authorization", "Bearer 1234")

	path, _ := me.findPath(request)
	operation := me.findOperation(request, path)

	// mutate securityScheme to be a query param
	doc.Components.SecuritySchemes["ApiKeyAuth"].Type = "http"
	doc.Components.SecuritySchemes["ApiKeyAuth"].Scheme = "bearer"

	err := me.ValidateSecurity(request, operation)
	assert.NoError(t, err)
}

func TestNewMockEngine_BuildResponse_SimpleValid(t *testing.T) {
	doc := resetState()
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

func TestNewMockEngine_BuildResponse_SimpleInvalid_NoContentType(t *testing.T) {
	doc := resetState()
	me := NewMockEngine(doc, false)

	request, _ := http.NewRequest(http.MethodGet, "https://api.pb33f.io/wiretap/giftshop/products", nil)
	b, status, err := me.GenerateResponse(request)

	assert.NoError(t, err)
	assert.Equal(t, 415, status)

	var decoded map[string]any
	_ = json.Unmarshal(b, &decoded)

	assert.Equal(t, "Media type not supported (415)", decoded["title"])
	assert.Equal(t, "The media type requested '' is not supported by this operation", decoded["detail"])
}

func TestNewMockEngine_BuildResponse_SimpleValid_Pretty(t *testing.T) {
	doc := resetState()
	me := NewMockEngine(doc, true)

	request, _ := http.NewRequest(http.MethodGet, "https://api.pb33f.io/wiretap/giftshop/products/pb0001", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")

	b, status, err := me.GenerateResponse(request)

	assert.NoError(t, err)
	assert.Equal(t, 200, status)
	assert.Equal(t, "{\n  \"category\": \"shirts\",\n  \"description\": \"A t-shirt with the pb33f logo on the"+
		" front\",\n  \"id\": \"d1404c5c-69bd-4cd2-a4cf-b47c79a30112\",\n  \"image\": \"https://pb33f.io/images/t-shirt.png\",\n "+
		" \"name\": \"pb33f t-shirt\",\n  \"price\": 19.99,\n  \"shortCode\": \"pb0001\"\n}", string(b))
}

func TestNewMockEngine_BuildResponse_MissingPath_404(t *testing.T) {
	doc := resetState()
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
	doc := resetState()
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

	doc := resetState()
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

	doc := resetState()
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
