// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package validation

import (
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/v3"
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

func TestNewValidator(t *testing.T) {

	validator := NewValidator(doc)
	assert.NotNil(t, validator)
	assert.NotNil(t, validator.doc)
	assert.NotNil(t, validator.requestValidator)
	assert.NotNil(t, validator.responseValidator)
	assert.NotNil(t, validator.paramValidator)
}
