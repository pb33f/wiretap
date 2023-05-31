// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package daemon

import (
	"github.com/pb33f/libopenapi-validator/errors"
	"github.com/pb33f/libopenapi-validator/parameters"
	"github.com/pb33f/libopenapi-validator/paths"
	"github.com/pb33f/libopenapi-validator/requests"
	"github.com/pb33f/libopenapi-validator/responses"
	"github.com/pb33f/ranch/model"
	"net/http"
)

func (ws *WiretapService) validateResponse(
	request *model.Request,
	responseValidator responses.ResponseBodyValidator,
	returnedResponse *http.Response) {

	_, validationErrors := responseValidator.ValidateResponseBody(request.HttpRequest, returnedResponse)

	// wipe out any path not found errors, they are not relevant to the response.
	var cleanedErrors []*errors.ValidationError
	for x := range validationErrors {
		if !validationErrors[x].IsPathMissingError() {
			cleanedErrors = append(cleanedErrors, validationErrors[x])
		}
	}

	transaction := buildResponse(request, returnedResponse)
	if len(cleanedErrors) > 0 {
		transaction.ResponseValidation = cleanedErrors
	}
	ws.transactionStore.Put(request.Id.String(), transaction, nil)

	if len(cleanedErrors) > 0 {
		ws.broadcastResponseValidationErrors(request, returnedResponse, cleanedErrors)
	} else {
		ws.broadcastResponse(request, returnedResponse)
	}
}

func (ws *WiretapService) validateRequest(
	request *model.Request,
	requestValidator requests.RequestBodyValidator,
	paramValidator parameters.ParameterValidator,
	responseValidator responses.ResponseBodyValidator) {

	var validationErrors, cleanedErrors []*errors.ValidationError

	// find path and populate validators.
	path, pathErrors, pv := paths.FindPath(request.HttpRequest, ws.docModel)
	requestValidator.SetPathItem(path, pv)
	paramValidator.SetPathItem(path, pv)
	responseValidator.SetPathItem(path, pv)

	// record any path errors.
	validationErrors = append(validationErrors, pathErrors...)

	// validate params
	_, queryParams := paramValidator.ValidateQueryParams(request.HttpRequest)
	_, headerParams := paramValidator.ValidateHeaderParams(request.HttpRequest)
	_, cookieParams := paramValidator.ValidateCookieParams(request.HttpRequest)
	_, pathParams := paramValidator.ValidatePathParams(request.HttpRequest)

	validationErrors = append(validationErrors, queryParams...)
	validationErrors = append(validationErrors, headerParams...)
	validationErrors = append(validationErrors, cookieParams...)
	validationErrors = append(validationErrors, pathParams...)

	// validate request
	_, requestErrors := requestValidator.ValidateRequestBody(request.HttpRequest)
	validationErrors = append(validationErrors, requestErrors...)

	pm := false
	for i := range validationErrors {
		if validationErrors[i].IsPathMissingError() {
			if !pm {
				cleanedErrors = append(cleanedErrors, validationErrors[i])
				pm = true
			}
		} else {
			cleanedErrors = append(cleanedErrors, validationErrors[i])
		}
	}

	// record results
	transaction := buildRequest(request)
	if len(cleanedErrors) > 0 {
		transaction.RequestValidation = cleanedErrors
	}
	ws.transactionStore.Put(request.Id.String(), transaction, nil)

	// broadcast what we found.
	if len(cleanedErrors) > 0 {
		ws.broadcastRequestValidationErrors(request, cleanedErrors)
	} else {
		ws.broadcastRequest(request)
	}
}
