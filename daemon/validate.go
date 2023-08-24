// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

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
	returnedResponse *http.Response) []*errors.ValidationError {

	var validationErrors []*errors.ValidationError

	if ws.document != nil && ws.docModel != nil {
		_, validationErrors = responseValidator.ValidateResponseBody(request.HttpRequest, returnedResponse)
	}

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
	return validationErrors
}

func (ws *WiretapService) validateRequest(
	modelRequest *model.Request,
	httpRequest *http.Request,
	requestValidator requests.RequestBodyValidator,
	paramValidator parameters.ParameterValidator,
	responseValidator responses.ResponseBodyValidator) []*errors.ValidationError {

	var validationErrors, cleanedErrors []*errors.ValidationError

	if ws.document != nil && ws.docModel != nil {

		// find path and populate validators.
		path, pathErrors, pv := paths.FindPath(httpRequest, ws.docModel)
		requestValidator.SetPathItem(path, pv)
		paramValidator.SetPathItem(path, pv)
		responseValidator.SetPathItem(path, pv)

		// record any path errors.
		validationErrors = append(validationErrors, pathErrors...)

		// validate params
		_, queryParams := paramValidator.ValidateQueryParams(httpRequest)
		_, headerParams := paramValidator.ValidateHeaderParams(httpRequest)
		_, cookieParams := paramValidator.ValidateCookieParams(httpRequest)
		_, pathParams := paramValidator.ValidatePathParams(httpRequest)

		validationErrors = append(validationErrors, queryParams...)
		validationErrors = append(validationErrors, headerParams...)
		validationErrors = append(validationErrors, cookieParams...)
		validationErrors = append(validationErrors, pathParams...)

		// validate modelRequest
		_, requestErrors := requestValidator.ValidateRequestBody(httpRequest)
		validationErrors = append(validationErrors, requestErrors...)

	}

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
	transaction := buildRequest(modelRequest, httpRequest)
	if len(cleanedErrors) > 0 {
		transaction.RequestValidation = cleanedErrors
	}
	ws.transactionStore.Put(modelRequest.Id.String(), modelRequest, nil)

	// broadcast what we found.
	if len(cleanedErrors) > 0 {
		ws.broadcastRequestValidationErrors(modelRequest, cleanedErrors, transaction)
	} else {
		ws.broadcastRequest(modelRequest, transaction)
	}
	return cleanedErrors
}
