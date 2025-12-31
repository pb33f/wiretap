// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package validation

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/stretchr/testify/assert"
)

func TestNewValidator(t *testing.T) {
	doc := loadStrictTestSpec(t)
	validator := NewHttpValidator(doc)
	assert.NotNil(t, validator)
}

// loadStrictTestSpec loads the local strict mode test spec
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

func TestNewHttpValidatorStrict_UndeclaredBodyProperty(t *testing.T) {
	strictDoc := loadStrictTestSpec(t)
	validator := NewStrictHttpValidator(strictDoc)

	// Create request with undeclared property "extra"
	body := strings.NewReader(`{"name": "test", "extra": "undeclared"}`)
	req, _ := http.NewRequest("POST", "http://localhost/users", body)
	req.Header.Set("Content-Type", "application/json")

	valid, errs := validator.ValidateHttpRequest(req)
	assert.False(t, valid)
	assert.NotEmpty(t, errs)

	// Check that the error is about the undeclared property
	found := false
	for _, err := range errs {
		if strings.Contains(err.Message, "extra") || strings.Contains(err.Reason, "extra") {
			found = true
			break
		}
	}
	assert.True(t, found, "expected error about undeclared property 'extra'")
}

func TestNewHttpValidator_UndeclaredBodyProperty_Allowed(t *testing.T) {
	strictDoc := loadStrictTestSpec(t)
	validator := NewHttpValidator(strictDoc)

	// Create request with undeclared property "extra"
	body := strings.NewReader(`{"name": "test", "extra": "undeclared"}`)
	req, _ := http.NewRequest("POST", "http://localhost/users", body)
	req.Header.Set("Content-Type", "application/json")

	valid, errs := validator.ValidateHttpRequest(req)
	// Non-strict mode should allow undeclared properties (no additionalProperties: false)
	assert.True(t, valid, "non-strict mode should allow undeclared properties")
	assert.Empty(t, errs)
}

func TestNewHttpValidatorStrict_UndeclaredQueryParam(t *testing.T) {
	strictDoc := loadStrictTestSpec(t)
	validator := NewStrictHttpValidator(strictDoc)

	// Create request with undeclared query param "debug"
	req, _ := http.NewRequest("GET", "http://localhost/users?limit=10&debug=true", nil)

	valid, errs := validator.ValidateHttpRequest(req)
	assert.False(t, valid)
	assert.NotEmpty(t, errs)

	// Check that the error is about the undeclared query param
	found := false
	for _, err := range errs {
		if strings.Contains(err.Message, "debug") || strings.Contains(err.Reason, "debug") {
			found = true
			break
		}
	}
	assert.True(t, found, "expected error about undeclared query param 'debug'")
}

func TestNewHttpValidatorStrict_UndeclaredHeader(t *testing.T) {
	strictDoc := loadStrictTestSpec(t)
	validator := NewStrictHttpValidator(strictDoc)

	// Create request with undeclared header "X-Debug-Mode"
	req, _ := http.NewRequest("GET", "http://localhost/users", nil)
	req.Header.Set("X-Debug-Mode", "true")

	valid, errs := validator.ValidateHttpRequest(req)
	assert.False(t, valid)
	assert.NotEmpty(t, errs)

	// Check that the error is about the undeclared header
	found := false
	for _, err := range errs {
		if strings.Contains(err.Message, "X-Debug-Mode") || strings.Contains(err.Reason, "X-Debug-Mode") {
			found = true
			break
		}
	}
	assert.True(t, found, "expected error about undeclared header 'X-Debug-Mode'")
}

func TestNewHttpValidatorStrict_UndeclaredCookie(t *testing.T) {
	strictDoc := loadStrictTestSpec(t)
	validator := NewStrictHttpValidator(strictDoc)

	// Create request with undeclared cookie "tracking"
	req, _ := http.NewRequest("GET", "http://localhost/users", nil)
	req.AddCookie(&http.Cookie{Name: "tracking", Value: "abc123"})

	valid, errs := validator.ValidateHttpRequest(req)
	assert.False(t, valid)
	assert.NotEmpty(t, errs)

	// Check that the error is about the undeclared cookie
	found := false
	for _, err := range errs {
		if strings.Contains(err.Message, "tracking") || strings.Contains(err.Reason, "tracking") {
			found = true
			break
		}
	}
	assert.True(t, found, "expected error about undeclared cookie 'tracking'")
}

func TestNewHttpValidatorStrict_UndeclaredResponseProperty(t *testing.T) {
	strictDoc := loadStrictTestSpec(t)
	validator := NewStrictHttpValidator(strictDoc)

	// Create a request for GET /users
	req, _ := http.NewRequest("GET", "http://localhost/users", nil)

	// Create a mock response with undeclared property "internal_id"
	responseBody := `{"id": 1, "name": "test", "internal_id": 999}`
	resp := &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
	}

	valid, errs := validator.ValidateHttpResponse(req, resp)
	assert.False(t, valid)
	assert.NotEmpty(t, errs)

	// Check that the error is about the undeclared response property
	found := false
	for _, err := range errs {
		if strings.Contains(err.Message, "internal_id") || strings.Contains(err.Reason, "internal_id") {
			found = true
			break
		}
	}
	assert.True(t, found, "expected error about undeclared response property 'internal_id'")
}

func TestNewHttpValidator_UndeclaredResponseProperty_Allowed(t *testing.T) {
	strictDoc := loadStrictTestSpec(t)
	validator := NewHttpValidator(strictDoc)

	// Create a request for GET /users
	req, _ := http.NewRequest("GET", "http://localhost/users", nil)

	// Create a mock response with undeclared property "internal_id"
	responseBody := `{"id": 1, "name": "test", "internal_id": 999}`
	resp := &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
	}

	valid, errs := validator.ValidateHttpResponse(req, resp)
	// Non-strict mode should allow undeclared properties
	assert.True(t, valid, "non-strict mode should allow undeclared response properties")
	assert.Empty(t, errs)
}

func TestStrictUsers_AdditionalPropertiesFalse_NonStrict(t *testing.T) {
	// Tests the /strict-users endpoint which has additionalProperties: false
	// Even in non-strict mode, additionalProperties: false should reject extra properties
	strictDoc := loadStrictTestSpec(t)
	validator := NewHttpValidator(strictDoc)

	// Create request with undeclared property "extra"
	body := strings.NewReader(`{"name": "test", "extra": "undeclared"}`)
	req, _ := http.NewRequest("POST", "http://localhost/strict-users", body)
	req.Header.Set("Content-Type", "application/json")

	valid, errs := validator.ValidateHttpRequest(req)
	// Even non-strict mode should reject this because additionalProperties: false is explicit
	assert.False(t, valid, "additionalProperties: false should reject extra properties even in non-strict mode")
	assert.NotEmpty(t, errs)
}

func TestStrictUsers_AdditionalPropertiesFalse_Strict(t *testing.T) {
	// Tests the /strict-users endpoint which has additionalProperties: false
	// Strict mode should also reject extra properties
	strictDoc := loadStrictTestSpec(t)
	validator := NewStrictHttpValidator(strictDoc)

	// Create request with undeclared property "extra"
	body := strings.NewReader(`{"name": "test", "extra": "undeclared"}`)
	req, _ := http.NewRequest("POST", "http://localhost/strict-users", body)
	req.Header.Set("Content-Type", "application/json")

	valid, errs := validator.ValidateHttpRequest(req)
	assert.False(t, valid, "strict mode should reject extra properties")
	assert.NotEmpty(t, errs)
}

func TestStrictUsers_ValidRequest(t *testing.T) {
	// Tests the /strict-users endpoint with a valid request (no extra properties)
	strictDoc := loadStrictTestSpec(t)
	validator := NewHttpValidator(strictDoc)

	// Create valid request with only declared property
	body := strings.NewReader(`{"name": "test"}`)
	req, _ := http.NewRequest("POST", "http://localhost/strict-users", body)
	req.Header.Set("Content-Type", "application/json")

	valid, errs := validator.ValidateHttpRequest(req)
	assert.True(t, valid, "valid request should pass")
	assert.Empty(t, errs)
}

// loadGiftshopSpec loads the giftshop OpenAPI spec for testing
func loadGiftshopSpec(t *testing.T) *v3.Document {
	spec, err := io.ReadAll(mustOpen(t, "../testdata/giftshop-openapi.yaml"))
	if err != nil {
		t.Fatalf("failed to read giftshop spec: %v", err)
	}
	d, err := libopenapi.NewDocument(spec)
	if err != nil {
		t.Fatalf("failed to parse giftshop spec: %v", err)
	}
	compiled, err := d.BuildV3Model()
	if err != nil {
		t.Fatalf("failed to build v3 model: %v", err)
	}
	return &compiled.Model
}

// TestGiftshop_CreateProduct_WithAdditionalProperties tests the create product endpoint
// with a valid product that includes undeclared additional properties.
// Without strict mode: should PASS (additionalProperties defaults to true)
// With strict mode: should FAIL (detects undeclared properties)
func TestGiftshop_CreateProduct_WithAdditionalProperties_NonStrict(t *testing.T) {
	doc := loadGiftshopSpec(t)
	validator := NewHttpValidator(doc)

	// Build a product that fulfills the contract completely
	// but also includes additional undeclared properties
	product := `{
		"id": "d1404c5c-69bd-4cd2-a4cf-b47c79a30112",
		"shortCode": "pb0001",
		"name": "pb33f t-shirt",
		"description": "A t-shirt with the pb33f logo on the front",
		"price": 19.99,
		"category": "shirts",
		"image": "https://pb33f.io/images/t-shirt.png",

		"internalSku": "SKU-12345-INTERNAL",
		"warehouseLocation": "SHELF-A3-BIN-42",
		"costPrice": 5.99,
		"supplierCode": "SUPP-001",
		"isDiscontinued": false,
		"lastRestockDate": "2024-01-15T10:30:00Z",
		"inventoryCount": 150,
		"metadata": {
			"createdBy": "admin",
			"tags": ["featured", "sale", "new-arrival"]
		}
	}`

	req, _ := http.NewRequest("POST", "https://api.pb33f.io/wiretap/giftshop/products", strings.NewReader(product))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-api-key") // Required for security

	valid, errs := validator.ValidateHttpRequest(req)

	// Non-strict mode: additional properties are allowed (no additionalProperties: false in schema)
	assert.True(t, valid, "non-strict mode should allow additional properties")
	assert.Empty(t, errs, "non-strict mode should have no validation errors")

	t.Logf("Non-strict mode: valid=%v, errors=%d", valid, len(errs))
}

func TestGiftshop_CreateProduct_WithAdditionalProperties_Strict(t *testing.T) {
	doc := loadGiftshopSpec(t)
	validator := NewStrictHttpValidator(doc)

	// Same product with additional undeclared properties
	product := `{
		"id": "d1404c5c-69bd-4cd2-a4cf-b47c79a30112",
		"shortCode": "pb0001",
		"name": "pb33f t-shirt",
		"description": "A t-shirt with the pb33f logo on the front",
		"price": 19.99,
		"category": "shirts",
		"image": "https://pb33f.io/images/t-shirt.png",

		"internalSku": "SKU-12345-INTERNAL",
		"warehouseLocation": "SHELF-A3-BIN-42",
		"costPrice": 5.99,
		"supplierCode": "SUPP-001",
		"isDiscontinued": false,
		"lastRestockDate": "2024-01-15T10:30:00Z",
		"inventoryCount": 150,
		"metadata": {
			"createdBy": "admin",
			"tags": ["featured", "sale", "new-arrival"]
		}
	}`

	req, _ := http.NewRequest("POST", "https://api.pb33f.io/wiretap/giftshop/products", strings.NewReader(product))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-api-key") // Required for security

	valid, errs := validator.ValidateHttpRequest(req)

	// Strict mode: should detect undeclared properties
	assert.False(t, valid, "strict mode should reject additional properties")
	assert.NotEmpty(t, errs, "strict mode should have validation errors for undeclared properties")

	// Log the errors for visibility
	t.Logf("Strict mode: valid=%v, errors=%d", valid, len(errs))
	for _, err := range errs {
		t.Logf("  - %s: %s", err.Message, err.Reason)
	}

	// Check that at least one of our undeclared properties was caught
	foundUndeclared := false
	undeclaredProps := []string{"internalSku", "warehouseLocation", "costPrice", "supplierCode", "isDiscontinued", "lastRestockDate", "inventoryCount", "metadata"}
	for _, err := range errs {
		for _, prop := range undeclaredProps {
			if strings.Contains(err.Message, prop) || strings.Contains(err.Reason, prop) {
				foundUndeclared = true
				break
			}
		}
	}
	assert.True(t, foundUndeclared, "should detect at least one undeclared property")
}

func TestGiftshop_GetProducts_ValidRequest(t *testing.T) {
	// Use GET /products which doesn't require auth and has no body
	// This tests that a completely valid request passes both modes
	doc := loadGiftshopSpec(t)

	// Non-strict: GET /products with valid category param
	req, _ := http.NewRequest("GET", "https://api.pb33f.io/wiretap/giftshop/products?category=shirts", nil)

	nonStrictValidator := NewHttpValidator(doc)
	valid, errs := nonStrictValidator.ValidateHttpRequest(req)
	assert.True(t, valid, "valid GET should pass non-strict mode")
	assert.Empty(t, errs)

	// Strict: same request should also pass
	req, _ = http.NewRequest("GET", "https://api.pb33f.io/wiretap/giftshop/products?category=shirts", nil)

	strictValidator := NewStrictHttpValidator(doc)
	valid, errs = strictValidator.ValidateHttpRequest(req)
	assert.True(t, valid, "valid GET should also pass strict mode")
	assert.Empty(t, errs)
}

func TestGiftshop_GetProducts_UndeclaredQueryParam_Strict(t *testing.T) {
	// GET /products with an undeclared query param should fail strict mode
	doc := loadGiftshopSpec(t)

	req, _ := http.NewRequest("GET", "https://api.pb33f.io/wiretap/giftshop/products?category=shirts&debug=true&verbose=1", nil)

	// Non-strict: should pass (undeclared params allowed)
	nonStrictValidator := NewHttpValidator(doc)
	valid, errs := nonStrictValidator.ValidateHttpRequest(req)
	assert.True(t, valid, "non-strict should allow undeclared query params")
	assert.Empty(t, errs)

	// Strict: should fail (undeclared params detected)
	req, _ = http.NewRequest("GET", "https://api.pb33f.io/wiretap/giftshop/products?category=shirts&debug=true&verbose=1", nil)

	strictValidator := NewStrictHttpValidator(doc)
	valid, errs = strictValidator.ValidateHttpRequest(req)
	assert.False(t, valid, "strict should reject undeclared query params")
	assert.NotEmpty(t, errs)

	t.Logf("Strict mode caught %d undeclared query params", len(errs))
	for _, err := range errs {
		t.Logf("  - %s", err.Message)
	}
}
