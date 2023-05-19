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
	"time"
)

func (ws *WiretapService) validateResponse(
	request *model.Request,
	responseValidator responses.ResponseBodyValidator,
	returnedResponse *http.Response) {

	time.Sleep(3 * time.Second) // simulate a slow response.

	_, validationErrors := responseValidator.ValidateResponseBody(request.HttpRequest, returnedResponse)
	if len(validationErrors) > 0 {
		ws.broadcastResponseValidationErrors(request, returnedResponse, validationErrors)
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

	// broadcast what we found.
	if len(cleanedErrors) > 0 {
		ws.broadcastRequestValidationErrors(request, cleanedErrors)
	} else {
		ws.broadcastRequest(request)
	}
}
