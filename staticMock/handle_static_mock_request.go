// Copyright 2023-2024 Princess Beef Heavy Industries, LLC / Dave Shanley
// https://pb33f.io
//
// SPDX-License-Identifier: AGPL

package staticMock

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/pb33f/ranch/model"
	"github.com/pb33f/wiretap/shared"
)

func (sms *StaticMockService) getBodyFromHttpRequest(request *http.Request) interface{} {
	bodyBytes, err := io.ReadAll(request.Body)
	if err != nil {
		panic(err)
	}

	var bodyJsonObj interface{}

	if len(bodyBytes) == 0 {
		return bodyJsonObj
	}

	err = json.Unmarshal(bodyBytes, &bodyJsonObj)
	if err != nil {
		sms.logger.Error("Error decoding JSON of incoming request")
		panic(err)
	}

	return bodyJsonObj
}

func (sms *StaticMockService) compareJsonBody(mock StaticMockDefinitionRequest, request *http.Request) bool {
	// Mock body is JSON but incoming body is not JSON
	if request.Header.Get("Content-Type") != "application/json" {
		return false
	}

	incomingBody := sms.getBodyFromHttpRequest(request)

	// Check if the JSON object or array is a subset of the incoming body
	return shared.IsSubset(mock.Body, incomingBody)
}

// Function to transform []string values to []interface{}(string)
func (sms *StaticMockService) transStrArrToInterfaceArr(strArr []string) []interface{} {
	strArrTransformedValues := make([]interface{}, 0)
	for _, value := range strArr {
		strArrTransformedValues = append(strArrTransformedValues, interface{}(value))
	}
	return strArrTransformedValues
}

// Function to compare headers
func (sms *StaticMockService) compareHeaders(mockHeaders map[string]any, incoming *http.Request) bool {
	found := true
	// Check if all headers in mockHeaders are subset of incoming headers
	for key, value := range mockHeaders {
		switch v := value.(type) {
		case string:
			found = found && shared.IsSubset([]interface{}{v}, sms.transStrArrToInterfaceArr(incoming.Header[key]))
		case []interface{}:
			found = found && shared.IsSubset(value, sms.transStrArrToInterfaceArr(incoming.Header[key]))
		}
	}

	return found
}

func (sms *StaticMockService) compareQueryParams(mockQueryParams map[string]any, incomingQueries url.Values) bool {
	found := true
	// Check if all headers in mockHeaders are subset of incoming headers
	for key, value := range mockQueryParams {
		switch v := value.(type) {
		case string:
			found = found && shared.IsSubset([]interface{}{v}, sms.transStrArrToInterfaceArr(incomingQueries[key]))
		case []interface{}:
			found = found && shared.IsSubset(value, sms.transStrArrToInterfaceArr(incomingQueries[key]))
		}
	}

	return found
}

func (sms *StaticMockService) compareBody(mock StaticMockDefinitionRequest, incoming *http.Request) bool {
	switch mb := mock.Body.(type) {
	case string: // Case string body
		incomingBodyBytes, err := io.ReadAll(incoming.Body)
		if err != nil {
			panic(err)
		}

		if string(incomingBodyBytes) != string(mb) {
			return false
		}
	case map[string]interface{}: // Case JSON Object
		if !sms.compareJsonBody(mock, incoming) {
			return false
		}
	case []interface{}: // Case JSON Array
		if !sms.compareJsonBody(mock, incoming) {
			return false
		}
	default:
		sms.logger.Error("Unsupported type of body in mock definition", mb)
		return false
	}

	return true
}

// Function to check if two requests are identical
func (sms *StaticMockService) isRequestMatch(mock StaticMockDefinitionRequest, incoming *http.Request) bool {
	// Compare Host if defined
	if mock.Host != "" && !shared.StringCompare(mock.Host, incoming.Host) {
		return false
	}

	// Compare HTTP method
	if incoming.Method != mock.Method {
		return false
	}

	// Compare url of the request
	if mock.UrlPath != "" && !shared.StringCompare(mock.UrlPath, incoming.URL.Path) {
		return false
	}

	// Compare headers
	if mock.Header != nil {
		if !sms.compareHeaders(*mock.Header, incoming) {
			return false
		}
	}

	// Compare query parameters
	if mock.QueryParams != nil {
		if !sms.compareQueryParams(*mock.QueryParams, incoming.URL.Query()) {
			return false
		}
	}

	// Compare body content
	if mock.Body != nil {
		if !sms.compareBody(mock, incoming) {
			return false
		}
	}

	// If all checks passed, the requests match
	return true
}

func (sms *StaticMockService) checkStaticMockExists(request *http.Request) *StaticMockDefinition {
	var matchedMockDefinition *StaticMockDefinition
	// check for a static mock definition.
	for _, mockDefinition := range sms.mockDefinitions {
		if sms.isRequestMatch(mockDefinition.Request, request) {
			// found a match
			matchedMockDefinition = &mockDefinition
			break
		}
	}

	return matchedMockDefinition
}

func (sms *StaticMockService) handleStaticMockRequest(request *model.Request) {
	defer func() {
		if r := recover(); r != nil {
			sms.logger.Error("Recovered from panic in handleStaticMockRequest:", r)
			errorMessage := "Error in static mock handler"
			if err, ok := r.(error); ok && err.Error() != "" {
				errorMessage = err.Error()
			}
			errorBody := shared.MarshalError(shared.GenerateError(errorMessage, 500, "Internal server error", "", r))
			errorResponse := http.Response{
				StatusCode: 500,
				Body:       io.NopCloser(bytes.NewBuffer([]byte(errorBody))),
			}
			sms.wiretapService.HandleStaticMockResponse(request, &errorResponse)
		}
	}()

	// check for a static mock definition.
	matchedMockDefinition := sms.checkStaticMockExists(request.HttpRequest)

	if matchedMockDefinition == nil {
		// no static mock found, pass the request to the wiretap service.
		sms.wiretapService.HandleHttpRequest(request)
		return
	}

	// found a static mock, handle it.
	response := sms.getStaticMockResponse(*matchedMockDefinition, request.HttpRequest)

	sms.wiretapService.HandleStaticMockResponse(request, response)
}
