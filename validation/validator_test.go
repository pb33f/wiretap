// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

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

	validator := NewHttpValidator(doc)
	assert.NotNil(t, validator)
}
