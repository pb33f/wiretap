// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package mock

import (
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi-validator/helpers"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"
)

var doc *v3.Document

func init() {
	resp, err := http.Get("https://api.pb33f.io/wiretap/giftshop-openapi.yaml")
	if err != nil {
		panic(err)
	}
	spec, _ := io.ReadAll(resp.Body)
	d, _ := libopenapi.NewDocument(spec)
	compiled, _ := d.BuildV3Model()
	doc = &compiled.Model
}

/*
   // build a request
   	request, _ := http.NewRequest(http.MethodPost, "https://things.com/burgers/createBurger", bytes.NewReader(badJson))
   	request.Header.Set(helpers.ContentTypeHeader, "application/json")

   	// simulate a request/response
   	res := httptest.NewRecorder()
   	handler := func(w http.ResponseWriter, r *http.Request) {
   		w.Header().Set(helpers.ContentTypeHeader, r.Header.Get(helpers.ContentTypeHeader))
   		w.WriteHeader(http.StatusOK)
   		_, _ = w.Write(badJson)
   	}

   	// fire the request
   	handler(res, request)
*/

func TestNewMockEngine_findPath(t *testing.T) {
	me := NewMockEngine(doc)

	request, _ := http.NewRequest(http.MethodGet, "https://api.pb33f.io/wiretap/giftshop/products", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")

	path, errors := me.findPath(request)
	assert.NotNil(t, path)
	assert.Len(t, errors, 0)
}

func TestNewMockEngine_findPathNegative(t *testing.T) {
	me := NewMockEngine(doc)

	request, _ := http.NewRequest(http.MethodGet, "https://api.pb33f.io/wiretap/giftshop/invalid", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")

	path, errors := me.findPath(request)
	assert.Nil(t, path)
	assert.Len(t, errors, 1)
}

func TestNewMockEngine_findOperation(t *testing.T) {
	me := NewMockEngine(doc)

	request, _ := http.NewRequest(http.MethodGet, "https://api.pb33f.io/wiretap/giftshop/products", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")

	path, _ := me.findPath(request)
	operation := me.findOperation(request, path)
	assert.NotNil(t, operation)

}

func TestNewMockEngine_findOperationNegative(t *testing.T) {
	me := NewMockEngine(doc)

	request, _ := http.NewRequest(http.MethodPatch, "https://api.pb33f.io/wiretap/giftshop/products", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")

	path, _ := me.findPath(request)
	operation := me.findOperation(request, path)
	assert.Nil(t, operation)
}

func TestNewMockEngine_ValidateSecurity_FailAPIKey_Header(t *testing.T) {
	me := NewMockEngine(doc)

	request, _ := http.NewRequest(http.MethodPost, "https://api.pb33f.io/wiretap/giftshop/products", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")
	path, _ := me.findPath(request)
	operation := me.findOperation(request, path)
	err := me.ValidateSecurity(request, operation)
	assert.Error(t, err)
	assert.Equal(t, "apiKey not found, no `X-API-Key` header found in request", err.Error())
}

func TestNewMockEngine_ValidateSecurity_PassAPIKey_Header(t *testing.T) {
	me := NewMockEngine(doc)

	request, _ := http.NewRequest(http.MethodPost, "https://api.pb33f.io/wiretap/giftshop/products", nil)
	request.Header.Set(helpers.ContentTypeHeader, "application/json")
	request.Header.Set("X-API-Key", "doesnotmatter")
	path, _ := me.findPath(request)
	operation := me.findOperation(request, path)
	err := me.ValidateSecurity(request, operation)
	assert.NoError(t, err)
}

func TestNewMockEngine_ValidateSecurity_FailAPIKey_Query(t *testing.T) {
	me := NewMockEngine(doc)

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
	me := NewMockEngine(doc)

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
	me := NewMockEngine(doc)

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
	me := NewMockEngine(doc)

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
	me := NewMockEngine(doc)

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
	me := NewMockEngine(doc)

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
