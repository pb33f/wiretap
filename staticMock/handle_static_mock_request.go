// Copyright 2023-2024 Princess Beef Heavy Industries, LLC / Dave Shanley
// https://pb33f.io
//
// SPDX-License-Identifier: AGPL

package staticMock

import (
	"bytes"
	"io"
	"net/http"

	"github.com/pb33f/ranch/model"
)

// Function to check if a slice contains a given element
func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// Function to check if two requests are identical
func isRequestMatch(incoming *http.Request, mock StaticMockDefinitionRequest) bool {
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
		mockHeaders := *mock.Header
		for key, value := range mockHeaders {
			if incomingValue, exists := incoming.Header[key]; !exists || len(incomingValue) != len(value) {
				return false
			} else {
				for _, v := range value {
					if contains(incomingValue, v) {
						return false
					}
				}
			}
		}
	}

	// // Compare body content
	// if incoming.Body != mock.Body {
	// 	return false
	// }

	// If all checks passed, the requests match
	return true
}

func (sms *StaticMockService) checkStaticMockExists(request *http.Request) *StaticMockDefinition {
	var matchedMockDefinition *StaticMockDefinition
	// check for a static mock definition.
	for _, mockDefinition := range sms.mockDefinitions {
		if isRequestMatch(request, mockDefinition.Request) {
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
	body := []byte(matchedMockDefinition.Response.Body)
	buff := bytes.NewBuffer(body)

	response := &http.Response{
		StatusCode: matchedMockDefinition.Response.StatusCode,
		Body:       io.NopCloser(buff),
	}
	header := http.Header{}
	response.Header = header

	sms.wiretapService.HandleStaticMockResponse(request, response)
}
