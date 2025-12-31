// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package validation

import (
	"net/http"

	validator "github.com/pb33f/libopenapi-validator"
	"github.com/pb33f/libopenapi-validator/config"
	"github.com/pb33f/libopenapi-validator/errors"
	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

type HttpValidator interface {
	ValidateHttpRequest(request *http.Request) (bool, []*errors.ValidationError)
	ValidateHttpResponse(request *http.Request, response *http.Response) (bool, []*errors.ValidationError)
}

func NewHttpValidator(doc *v3.Document) HttpValidator {
	return validator.NewValidatorFromV3Model(doc)
}

// NewStrictHttpValidator creates a validator with strict mode enabled.
// Strict mode detects undeclared properties, parameters, headers, and cookies.
func NewStrictHttpValidator(doc *v3.Document) HttpValidator {
	return validator.NewValidatorFromV3Model(doc, config.WithStrictMode())
}

// NewHttpValidatorWithConfig creates a validator based on configuration.
// When strictMode is true, undeclared properties, parameters, headers, and cookies are detected.
func NewHttpValidatorWithConfig(doc *v3.Document, strictMode bool) HttpValidator {
	if strictMode {
		return NewStrictHttpValidator(doc)
	}
	return NewHttpValidator(doc)
}
