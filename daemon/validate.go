// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package daemon

import (
	"github.com/pb33f/libopenapi-validator/paths"
	"github.com/pb33f/ranch/model"
	"github.com/pb33f/wiretap/shared"
	"net/http"
)

func (ws *WiretapService) getValidatorForRequest(request *model.Request) *documentValidator {

	if len(ws.documentValidators) == 1 {
		return &ws.documentValidators[0]
	} else {

		for _, docValidator := range ws.documentValidators {
			// Find the first path match between all provided specifications
			pathItem, _, _ := paths.FindPath(request.HttpRequest, docValidator.docModel, nil)

			if pathItem != nil {
				return &docValidator
			}
		}

		// If we haven't found a path, let's pick the first validator to validate against. This should just produce a path not found error.
		if len(ws.documentValidators) > 0 {
			return &ws.documentValidators[0]
		}
	}

	return nil
}

func (ws *WiretapService) ValidateResponse(
	request *model.Request,
	returnedResponse *http.Response) []*shared.WiretapValidationError {

	var validationErrors []*shared.WiretapValidationError

	docValidator := ws.getValidatorForRequest(request)
	if docValidator != nil {
		_, newValidationErrors := docValidator.validator.ValidateHttpResponse(request.HttpRequest, returnedResponse)
		validationErrors = shared.ConvertValidationErrors(docValidator.documentName, newValidationErrors)
	}

	// wipe out any path not found errors, they are not relevant to the response.
	var cleanedErrors []*shared.WiretapValidationError
	for x := range validationErrors {
		if !validationErrors[x].IsPathMissingError() {
			cleanedErrors = append(cleanedErrors, validationErrors[x])
		}
	}

	transaction := BuildResponse(request, returnedResponse)
	if len(cleanedErrors) > 0 {
		transaction.ResponseValidation = cleanedErrors
	}
	ws.transactionStore.Put(request.Id.String(), transaction, nil)

	if len(cleanedErrors) > 0 {
		ws.streamChan <- cleanedErrors
		ws.broadcastResponseValidationErrors(request, returnedResponse, cleanedErrors)
	} else {
		ws.broadcastResponse(request, returnedResponse)
	}
	return validationErrors
}

func (ws *WiretapService) ValidateRequest(
	modelRequest *model.Request,
	httpRequest *http.Request) []*shared.WiretapValidationError {

	var validationErrors, cleanedErrors []*shared.WiretapValidationError

	docValidator := ws.getValidatorForRequest(modelRequest)
	if docValidator != nil {
		_, newValidationErrors := docValidator.validator.ValidateHttpRequest(httpRequest)
		validationErrors = shared.ConvertValidationErrors(docValidator.documentName, newValidationErrors)
	}

	for _, validationError := range validationErrors {
		cleanedErrors = append(cleanedErrors, validationError)
	}
	// record results
	buildTransConfig := HttpTransactionConfig{
		OriginalRequest:   modelRequest.HttpRequest,
		NewRequest:        httpRequest,
		ID:                modelRequest.Id,
		TransactionConfig: ws.config,
	}

	transaction := BuildHttpTransaction(buildTransConfig)
	if len(cleanedErrors) > 0 {
		transaction.RequestValidation = cleanedErrors
	}
	ws.transactionStore.Put(modelRequest.Id.String(), modelRequest, nil)

	// broadcast what we found.
	if len(cleanedErrors) > 0 {
		ws.streamChan <- cleanedErrors
		ws.broadcastRequestValidationErrors(modelRequest, cleanedErrors, transaction)
	} else {
		ws.broadcastRequest(modelRequest, transaction)
	}
	return cleanedErrors
}
