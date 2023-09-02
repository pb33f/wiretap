// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package validation

import (
	validator "github.com/pb33f/libopenapi-validator"
	"github.com/pb33f/libopenapi-validator/errors"
	"github.com/pb33f/libopenapi/datamodel/high/v3"
	"net/http"
)

type HttpValidator interface {
	ValidateHttpRequest(request *http.Request) (bool, []*errors.ValidationError)
	ValidateHttpResponse(request *http.Request, response *http.Response) (bool, []*errors.ValidationError)
}

func NewHttpValidator(doc *v3.Document) HttpValidator {
	return validator.NewValidatorFromV3Model(doc)
}
