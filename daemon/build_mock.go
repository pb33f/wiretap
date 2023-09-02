// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package daemon

import (
	"encoding/json"
	"fmt"
	"github.com/pb33f/libopenapi-validator/helpers"
	"github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/ranch/model"
	"github.com/pb33f/wiretap/shared"
	"net/http"
	"strconv"
)

func (ws *WiretapService) findLowestSuccessCode(operation *v3.Operation) string {
	var lowestCode = 299
	for key := range operation.Responses.Codes {
		code, _ := strconv.Atoi(key)
		if code < lowestCode && code >= 200 {
			lowestCode = code
		}
	}
	if lowestCode == 299 {
		lowestCode = 200
	}
	return fmt.Sprintf("%d", lowestCode)
}

func (ws *WiretapService) lookForAppropriateErrorResponse(response *v3.Response) string {

	//return fmt.Sprintf("%d", lowestCode)
	return ""
}

func (ws *WiretapService) buildMockError(op *v3.Operation, request *http.Request, code int, err error) []byte {
	var wtError *shared.WiretapError
	errorResponse := op.Responses.FindResponseByCode(code)
	if errorResponse == nil {
		errorResponse = op.Responses.Default
		if errorResponse == nil {
			wtError = shared.GenerateError("[mock mode] unable to generate mock error for request", 404, err.Error(), "")
		}
	}
	if errorResponse != nil {

		mediaTypeSting := ws.extractMediaTypeHeader(request)
		responseBody := errorResponse.Content[mediaTypeSting]
		if responseBody == nil {
			wtError = shared.GenerateError("[mock mode] unable to generate mock  error for request", 404,
				"there is no response body for %d media type and error code", "")
		}

		//mock, error := ws.mockGenerator.GenerateMock(errorResponse, "")
	}

	wtError = shared.GenerateError("[mock error] unable to generate mock for request", 404, err.Error(), "")
	compiled, _ := json.Marshal(wtError)
	return compiled
}

func (ws *WiretapService) buildMockResponse(modelRequest *model.Request, httpRequest *http.Request) ([]byte, int, error) {

	//path, _, _ := paths.FindPath(httpRequest, ws.docModel)
	//var err error
	//if path == nil {
	//	err = fmt.Errorf("could not find path for request: %s", httpRequest.URL.Path)
	//	return ws.buildMockError(404, err), 404, err
	//}
	//
	//var operation = path.GetOperations()[strings.ToLower(httpRequest.Method)]
	//if operation == nil {
	//	err = fmt.Errorf("could not find operation for request: %s (%s)", httpRequest.URL.Path,
	//		httpRequest.Method)
	//	return ws.buildMockError(404, err), 404, err
	//}
	//
	//okResponse := operation.Responses.FindResponseByCode(200)
	//if okResponse == nil {
	//	okResponse = operation.Responses.Default
	//	if okResponse == nil {
	//
	//		err = fmt.Errorf("could not find 200 response for request: %s (%s)", httpRequest.URL.Path,
	//			httpRequest.Method)
	//		return ws.buildMockError(404, err), 404, err
	//	}
	//}
	//
	//mediaTypeSting := ws.extractMediaTypeHeader(modelRequest.HttpRequest)
	//
	//responseBody := okResponse.Content[mediaTypeSting]
	//if responseBody == nil {
	//	return nil, fmt.Errorf("could not find '%s' response body for request: %s (%s)",
	//		mediaTypeSting,
	//		httpRequest.URL.Path,
	//		httpRequest.Method)
	//}
	//
	//exampleName := ""
	//// check if there is a header named "preferred" in the request.
	//preferredHeaderValue := httpRequest.Header.Get("preferred")
	//if preferredHeaderValue != "" {
	//	exampleName = preferredHeaderValue
	//}
	//
	//// build a mock!
	//mock, err := ws.mockGenerator.GenerateMock(responseBody, exampleName)
	//if err != nil {
	//	return nil, err
	//}
	//
	//return mock, nil
	return nil, 0, nil
}

func (ws *WiretapService) extractMediaTypeHeader(request *http.Request) string {
	// extract the content type header from the request.
	contentType := request.Header.Get(helpers.ContentTypeHeader)

	// extract the media type from the content type header.
	mediaTypeSting, _, _ := helpers.ExtractContentType(contentType)

	if mediaTypeSting == "" {
		mediaTypeSting = "application/json"
	}
	return mediaTypeSting
}
