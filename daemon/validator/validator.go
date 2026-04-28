// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package validator

import (
	"net/http"

	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/ranch/model"
	"github.com/pb33f/wiretap/mock"
	"github.com/pb33f/wiretap/shared"
	"github.com/pb33f/wiretap/validation"
)

type DocumentValidator struct {
	DocumentName string
	Document     libopenapi.Document
	DocModel     *v3.Document
	Validator    validation.HttpValidator
	MockEngine   *mock.ResponseMockEngine
}

type Validator struct {
	documentValidators []DocumentValidator
	router             *validation.SpecRouter
}

func New(documentValidators []DocumentValidator) *Validator {
	docs := make([]DocumentValidator, len(documentValidators))
	copy(docs, documentValidators)

	routeDocs := make([]validation.DocumentValidator, len(docs))
	for i := range docs {
		routeDocs[i] = validation.DocumentValidator{
			DocumentName: docs[i].DocumentName,
			DocModel:     docs[i].DocModel,
			Validator:    docs[i].Validator,
		}
	}

	return &Validator{
		documentValidators: docs,
		router:             validation.NewSpecRouter(routeDocs),
	}
}

func (v *Validator) GetValidatorForRequest(request *model.Request) *DocumentValidator {
	if v == nil || len(v.documentValidators) == 0 {
		return nil
	}
	if request == nil || request.HttpRequest == nil {
		return nil
	}

	index, routeDoc := v.router.ResolveIndex(request.HttpRequest)
	if routeDoc == nil {
		return nil
	}
	if index >= 0 && index < len(v.documentValidators) {
		return &v.documentValidators[index]
	}

	// Preserve existing behavior: fall back to the first validator so callers
	// still receive the usual path-not-found validation error.
	return &v.documentValidators[0]
}

func (v *Validator) ValidateResponse(
	request *model.Request,
	returnedResponse *http.Response,
) ([]*shared.WiretapValidationError, []*shared.WiretapValidationError) {
	var validationErrors []*shared.WiretapValidationError

	docValidator := v.GetValidatorForRequest(request)
	if docValidator != nil {
		_, newValidationErrors := docValidator.Validator.ValidateHttpResponse(request.HttpRequest, returnedResponse)
		validationErrors = shared.ConvertValidationErrors(docValidator.DocumentName, newValidationErrors)
	}

	// Wipe out any path-not-found errors; they are not relevant to responses.
	var cleanedErrors []*shared.WiretapValidationError
	for x := range validationErrors {
		if !validationErrors[x].IsPathMissingError() {
			cleanedErrors = append(cleanedErrors, validationErrors[x])
		}
	}
	return validationErrors, cleanedErrors
}

func (v *Validator) ValidateRequest(
	modelRequest *model.Request,
	httpRequest *http.Request,
) []*shared.WiretapValidationError {
	var validationErrors []*shared.WiretapValidationError

	docValidator := v.GetValidatorForRequest(modelRequest)
	if docValidator != nil {
		_, newValidationErrors := docValidator.Validator.ValidateHttpRequest(httpRequest)
		validationErrors = shared.ConvertValidationErrors(docValidator.DocumentName, newValidationErrors)
	}
	return validationErrors
}
