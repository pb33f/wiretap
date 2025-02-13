// Copyright 2023-2024 Princess Beef Heavy Industries, LLC / Dave Shanley
// https://pb33f.io
//
// SPDX-License-Identifier: AGPL

package staticMock

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/pb33f/ranch/model"
	"github.com/pb33f/wiretap/shared"
)

func (sms *StaticMockService) getBodyBytesFromHttpRequest(request *http.Request) []byte {
	bodyBytes, err := io.ReadAll(request.Body)
	if err != nil {
		panic(err)
	}
	return bodyBytes
}

func (sms *StaticMockService) compareJsonBody(mock StaticMockDefinitionRequest, request *http.Request) bool {
	// Mock body is JSON but incoming body is not JSON
	if request.Header.Get("Content-Type") != "application/json" {
		return false
	}

	incomingBodyBytes := sms.getBodyBytesFromHttpRequest(request)
	var incomingBodyJson interface{}
	err := json.Unmarshal(incomingBodyBytes, &incomingBodyJson)
	if err != nil {
		sms.logger.Error("Error decoding JSON of incoming request")
		panic(err)
	}
	// Check if the JSON object or array is a subset of the incoming body
	return shared.IsSubset(mock.Body, incomingBodyJson)
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
		incomingBodyBytes := sms.getBodyBytesFromHttpRequest(incoming)
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
	// Compare HTTP method
	if incoming.Method != mock.Method {
		return false
	}

	// Compare URL
	// Condition where urlPath is defined in mock
	if mock.UrlPath != "" && incoming.URL.Path != mock.UrlPath {
		return false
	}

	// Condition where url object is defined in mock
	if mock.URL != nil && incoming.URL != mock.URL {
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
	// check for a static mock definition.
	matchedMockDefinition := sms.checkStaticMockExists(request.HttpRequest)

	if matchedMockDefinition == nil {
		// no static mock found, pass the request to the wiretap service.
		sms.wiretapService.HandleHttpRequest(request)
		return
	}

	// found a static mock, handle it.
	response := sms.getStaticMockResponse(*matchedMockDefinition)

	sms.wiretapService.HandleStaticMockResponse(request, response)
}
