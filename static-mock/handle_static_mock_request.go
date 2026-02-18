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

// getJsonBodyFromHttpRequest reads the body of the incoming request and returns it as a JSON interface{}
func (sms *StaticMockService) getJsonBodyFromHttpRequest(request *http.Request) interface{} {
	bodyBytes, err := io.ReadAll(request.Body)
	if err != nil {
		panic(err)
	}

	// Restore request.Body so it can be read again
	request.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	var bodyJsonObj interface{}

	if len(bodyBytes) == 0 {
		return bodyJsonObj
	}

	err = json.Unmarshal(bodyBytes, &bodyJsonObj)
	if err != nil {
		sms.logger.Error("Error decoding JSON of incoming request. JSON => \n%s", string(bodyBytes), err)
		panic(err)
	}

	return bodyJsonObj
}

// getFormBodyFromHttpRequest reads the body of the incoming request and returns it as a parsed form map
func (sms *StaticMockService) getFormBodyFromHttpRequest(request *http.Request) interface{} {
	bodyBytes, err := io.ReadAll(request.Body)
	if err != nil {
		panic(err)
	}

	// Restore request.Body so it can be read again
	request.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	if len(bodyBytes) == 0 {
		return nil
	}

	parsedForm, err := url.ParseQuery(string(bodyBytes))
	if err != nil {
		sms.logger.Error("Error parsing form-urlencoded body: %s", string(bodyBytes), err)
		panic(err)
	}

	// Convert url.Values to map[string]interface{} for consistent handling
	formMap := make(map[string]interface{})
	for key, values := range parsedForm {
		if len(values) == 1 {
			formMap[key] = values[0]
		} else {
			formMap[key] = values
		}
	}
	return formMap
}

// compareJsonBody compares the JSON body of the incoming request with the mock definition
func (sms *StaticMockService) compareJsonBody(mock StaticMockDefinitionRequest, request *http.Request) bool {
	incomingBody := sms.getJsonBodyFromHttpRequest(request)

	// Check if the JSON object or array is a subset of the incoming body
	return shared.IsSubset(mock.Body, incomingBody)
}

// compareFormBody compares the form-urlencoded body of the incoming request with the mock definition
func (sms *StaticMockService) compareFormBody(mock StaticMockDefinitionRequest, request *http.Request) bool {
	incomingBody := sms.getFormBodyFromHttpRequest(request)

	// Check if the mock body is a subset of the incoming form body
	return shared.IsSubset(mock.Body, incomingBody)
}

// transStrArrToInterfaceArr transforms a string array to an interface array (helper method)
func (sms *StaticMockService) transStrArrToInterfaceArr(strArr []string) []interface{} {
	strArrTransformedValues := make([]interface{}, 0)
	for _, value := range strArr {
		strArrTransformedValues = append(strArrTransformedValues, interface{}(value))
	}
	return strArrTransformedValues
}

// compareHeaders compares the headers of the incoming request with the mock definition
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

// compareQueryParams compares the query parameters of the incoming request with the mock definition
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

// compareBody compares the body of the incoming request with the mock definition
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
	case map[string]interface{}: // Case JSON Object or Form body
		contentType := incoming.Header.Get("Content-Type")
		switch contentType {
		case ContentTypeFormUrlEncoded:
			if !sms.compareFormBody(mock, incoming) {
				return false
			}
		case ContentTypeJson:
			if !sms.compareJsonBody(mock, incoming) {
				return false
			}
		default:
			sms.logger.Error("Unsupported Content-Type", "contentType", contentType)
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

// isRequestMatch checks if the incoming request matches a mock definition
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

// checkStaticMockExists checks if a static mock definition exists for the incoming request.
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

// handleStaticMockRequest handles incoming requests and checks against static mock definitions.
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
