// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package mock

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	libopenapierrs "github.com/pb33f/libopenapi-validator/errors"
	"github.com/pb33f/libopenapi-validator/helpers"
	"github.com/pb33f/libopenapi-validator/paths"
	"github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/renderer"
	"github.com/pb33f/wiretap/shared"
	"github.com/pb33f/wiretap/validation"
)

type ResponseMockEngine struct {
	doc            *v3.Document
	validator      validation.HttpValidator
	mockEngine     *renderer.MockGenerator
	pretty         bool
	regexCache     *sync.Map
	hardValidation bool // when true, reject requests with validation errors
}

func NewMockEngine(document *v3.Document, pretty, useAllPropertyExamples bool) *ResponseMockEngine {
	me := renderer.NewMockGenerator(renderer.JSON)
	if pretty {
		me.SetPretty()
	}

	if useAllPropertyExamples {
		me.DisableRequiredCheck()
	}

	return &ResponseMockEngine{
		doc:            document,
		validator:      validation.NewHttpValidator(document),
		mockEngine:     me,
		pretty:         pretty,
		regexCache:     &sync.Map{},
		hardValidation: true, // default to rejecting on validation errors for backward compatibility
	}
}

// NewStrictMockEngine creates a mock engine with strict validation enabled.
// Strict mode detects undeclared properties, parameters, headers, and cookies in requests.
func NewStrictMockEngine(document *v3.Document, pretty, useAllPropertyExamples bool) *ResponseMockEngine {
	me := renderer.NewMockGenerator(renderer.JSON)
	if pretty {
		me.SetPretty()
	}

	if useAllPropertyExamples {
		me.DisableRequiredCheck()
	}

	return &ResponseMockEngine{
		doc:            document,
		validator:      validation.NewStrictHttpValidator(document),
		mockEngine:     me,
		pretty:         pretty,
		regexCache:     &sync.Map{},
		hardValidation: true, // default to rejecting on validation errors for backward compatibility
	}
}

// NewMockEngineWithConfig creates a mock engine based on configuration.
// When strictMode is true, undeclared request elements are validated strictly.
// When hardValidation is true, requests with validation errors are rejected with 422.
// When hardValidation is false, validation errors are reported but mocks are still served.
func NewMockEngineWithConfig(document *v3.Document, pretty, useAllPropertyExamples, strictMode, hardValidation bool) *ResponseMockEngine {
	var engine *ResponseMockEngine
	if strictMode {
		engine = NewStrictMockEngine(document, pretty, useAllPropertyExamples)
	} else {
		engine = NewMockEngine(document, pretty, useAllPropertyExamples)
	}
	engine.hardValidation = hardValidation
	return engine
}

func (rme *ResponseMockEngine) GenerateResponse(request *http.Request) ([]byte, int, error) {
	return rme.runWorkflow(request)
}

func (rme *ResponseMockEngine) ValidateSecurity(request *http.Request, operation *v3.Operation) error {
	// get out early if there is nothing to do.

	if rme.doc.Components != nil && rme.doc.Components.SecuritySchemes.Len() <= 0 {
		return nil
	}

	mustApply := make(map[string][]string)

	// operation security
	if len(operation.Security) > 0 {
		for _, securityRequirement := range operation.Security {

			// check if this operation has an empty security requirement, if so, we can skip it.
			if securityRequirement.Requirements.Len() <= 0 {
				return nil
			}

			for securityPairs := securityRequirement.Requirements.First(); securityPairs != nil; securityPairs = securityPairs.Next() {
				key := securityPairs.Key()
				scopes := securityPairs.Value()
				mustApply[key] = scopes
			}
		}
	}

	// global security if no local security found.
	if len(mustApply) <= 0 && len(rme.doc.Security) > 0 {
		for _, securityRequirement := range rme.doc.Security {
			// if an empty requirement is found, we can skip it, it's optional.
			if securityRequirement.Requirements.Len() <= 0 && securityRequirement.ContainsEmptyRequirement {
				return nil
			}
			for securityPairs := securityRequirement.Requirements.First(); securityPairs != nil; securityPairs = securityPairs.Next() {
				key := securityPairs.Key()
				scopes := securityPairs.Value()
				mustApply[key] = scopes
			}
		}
	}

	// check if we have any security requirements to apply.
	if len(mustApply) > 0 {
		// locate the security schemes components from the document.

		var failures []error
		compared := 0

		for scope, _ := range mustApply {

			securityComponent := rme.doc.Components.SecuritySchemes.GetOrZero(scope)
			if securityComponent != nil {

				// check if we have a security scheme that matches the type.
				switch securityComponent.Type {
				case "http":
					// check if we have a bearer scheme.
					if securityComponent.Scheme == "bearer" || securityComponent.Scheme == "basic" {
						compared++
						// check if we have a bearer token.
						if request.Header.Get("Authorization") == "" {
							failures = append(failures, fmt.Errorf("%s authentication failed: bearer token not found, "+
								"no `Authorization` header found in request", securityComponent.Scheme))
						}
					}

				case "apiKey":
					// check if the api key is being used in the header
					if securityComponent.In == "header" {
						compared++
						// check if we have a bearer token.
						if request.Header.Get(securityComponent.Name) == "" {
							failures = append(failures, fmt.Errorf("apiKey not found, no `%s` header found in request",
								securityComponent.Name))
						}
					}
					if securityComponent.In == "query" {
						compared++
						if request.URL.Query().Get(securityComponent.Name) == "" {
							failures = append(failures, fmt.Errorf("apiKey not found, no `%s` query parameter found in request",
								securityComponent.Name))
						}
					}
					if securityComponent.In == "cookie" {
						compared++
						cookie, _ := request.Cookie(securityComponent.Name)
						if cookie == nil {
							failures = append(failures, fmt.Errorf("apiKey not found, no `%s` cookie found in request",
								securityComponent.Name))
						}
					}
				}
			}
		}
		// only one needs to pass
		if len(failures) == compared {
			return errors.Join(failures...)
		}
	}
	return nil
}

func (rme *ResponseMockEngine) extractMediaTypeHeader(request *http.Request) string {
	// extract the content type header from the request.
	contentType := request.Header.Get(helpers.ContentTypeHeader)

	// extract the media type from the content type header.
	mediaTypeSting, _, _ := helpers.ExtractContentType(contentType)

	if mediaTypeSting == "" {
		mediaTypeSting = contentType // anything?
	}

	if mediaTypeSting == "" {
		// Check the Accept header for a content type
		contentType = request.Header.Get("Accept")
		mediaTypeSting, _, _ = helpers.ExtractContentType(contentType)
	}

	if mediaTypeSting == "" {
		mediaTypeSting = "application/json" // default
	}

	return mediaTypeSting
}

func (rme *ResponseMockEngine) findPath(request *http.Request) (*v3.PathItem, error) {
	path, errs, _ := paths.FindPath(request, rme.doc, rme.regexCache)
	return path, rme.packErrors(errs)
}

func (rme *ResponseMockEngine) findOperation(request *http.Request, pathItem *v3.PathItem) *v3.Operation {
	if pathItem == nil {
		return nil
	}
	ops := pathItem.GetOperations()
	if ops.Len() > 0 {
		op := pathItem.GetOperations().GetOrZero(strings.ToLower(request.Method))
		return op
	}
	return nil
}

func (rme *ResponseMockEngine) packErrors(errs []*libopenapierrs.ValidationError) error {
	var err error
	for _, e := range errs {
		err = errors.Join(err, e)
	}
	return err
}

func (rme *ResponseMockEngine) render(obj any) []byte {
	if obj == nil {
		return []byte{}
	}
	var b []byte
	var err error
	if !rme.pretty {
		b, err = json.Marshal(obj)
	} else {
		b, err = json.MarshalIndent(obj, "", "  ")
	}
	if err != nil {
		return []byte{}
	}
	return b
}

func (rme *ResponseMockEngine) buildErrorObject(status int, title, msg, hash string) *shared.WiretapError {
	return &shared.WiretapError{
		Type:   fmt.Sprintf("https://pb33f.io/wiretap/errors#%s", hash),
		Title:  fmt.Sprintf("%s (%d)", title, status),
		Status: status,
		Detail: msg,
	}
}

func (rme *ResponseMockEngine) buildError(status int, title, msg, hash string) []byte {
	return rme.render(rme.buildErrorObject(status, title, msg, hash))
}

func (rme *ResponseMockEngine) buildErrorWithPayload(status int, title, msg, hash string, payload any) []byte {
	wte := rme.buildErrorObject(status, title, msg, hash)
	wte.Payload = payload
	return rme.render(wte)
}

func (rme *ResponseMockEngine) extractPreferred(request *http.Request) string {
	return request.Header.Get(helpers.Preferred)
}

func (rme *ResponseMockEngine) runWorkflow(request *http.Request) ([]byte, int, error) {

	// get path, not valid? return 404
	path, err := rme.findPath(request)
	if err != nil {
		return rme.buildError(
			404,
			"Path / operation not found",
			fmt.Sprintf("Unable to locate the path '%s' with the method '%s'. %s",
				request.URL.Path, request.Method, err.Error()),
			"not_found",
		), 404, err

	}

	// find operation, not valid? return 404
	operation := rme.findOperation(request, path) // missing operation is cauight by

	// check the request is valid against security requirements.
	err = rme.ValidateSecurity(request, operation)
	if err != nil {
		mt, _ := rme.findBestMediaTypeMatch(operation, request, []string{"401"})
		if mt != nil {
			mock, mockErr := rme.mockEngine.GenerateMock(mt, rme.extractPreferred(request))
			if mockErr != nil {
				return rme.buildError(
					500,
					"Unable to build mock (401)",
					fmt.Sprintf("Errors occurred while generating an error 401 mock response: %s",
						errors.Join(err, mockErr)),
					"build_mock_error",
				), 500, mockErr
			}
			return mock, 401, err
		} else {
			return rme.buildError(
				401,
				"Unauthorized (401)",
				fmt.Sprintf("Unable to call '%s' on '%s', you are not authorized to access this resource",
					request.Method, request.URL.Path),
				"build_mock_error",
			), 401, err
		}
	}

	// validate the request against the document.
	// Only reject with error response if hardValidation is enabled.
	// When hardValidation is false, validation errors are reported separately via the daemon,
	// but the mock is still served to allow development/testing workflows.
	_, validationErrors := rme.validator.ValidateHttpRequest(request)
	if rme.hardValidation && len(validationErrors) > 0 {
		mt, _ := rme.findBestMediaTypeMatch(operation, request, []string{"422", "400"})
		if mt == nil {
			// no default, no valid response, inform use with a 500
			return rme.buildErrorWithPayload(
				500,
				"Invalid request, specification is insufficient",
				"The request failed validation, and the specification does not contain a "+
					"'422' or '400' response for this operation. Check payload for validation errors.",
				"validation_failed_and_spec_insufficient_error",
				validationErrors,
			), 500, rme.packErrors(validationErrors)
		}
		return rme.buildErrorWithPayload(
			422,
			"Invalid request",
			"The request failed validation, Check payload for validation errors.",
			"validation_failed_error",
			validationErrors,
		), 422, rme.packErrors(validationErrors)

	}

	preferred := rme.extractPreferred(request)

	var lo string
	var mt *v3.MediaType
	var noMT bool = true

	if preferred != "" {
		// If an explicit preferred header is present, let it have a chance to take precedence
		// This allows a developer to cause a 3xx, 4xx, or 5xx mocked response by passing
		// the appropriate example header value.
		mt, lo, noMT = rme.findMediaTypeContainingNamedExample(operation, request, preferred)
	}

	if noMT {
		// When no preferred header is passed, or preferred header did not match a named example
		lo = rme.findLowestSuccessCode(operation)
		mt, noMT = rme.findBestMediaTypeMatch(operation, request, []string{lo})
	}

	c, _ := strconv.Atoi(lo)
	if c == http.StatusNoContent {
		return nil, c, nil
	}

	if mt == nil && noMT {
		mtString := rme.extractMediaTypeHeader(request)
		return rme.buildError(
			415,
			"Media type not supported",
			fmt.Sprintf("The media type requested '%s' is not supported by this operation", mtString),
			"build_mock_error",
		), 415, nil
	}

	mock, mockErr := rme.mockEngine.GenerateMock(mt, preferred)
	if mockErr != nil {
		return rme.buildError(
			422,
			"Unable to build mock (422)",
			fmt.Sprintf("Errors occurred while generating an error 422 mock response: %s",
				errors.Join(err, mockErr)),
			"build_mock_error",
		), 422, mockErr
	}

	if len(mock) == 0 {
		return rme.buildError(
			200,
			"Response is empty",
			fmt.Sprintf("Nothing was generated for the request '%s' with the method '%s'. Response is empty",
				request.URL.Path, request.Method),
			"empty",
		), 200, err
	}

	// check for wiretap-status-code in header and override the code, regardless of what was found in the spec.
	if statusCode := request.Header.Get("wiretap-status-code"); statusCode != "" {
		c, _ = strconv.Atoi(statusCode)
	}

	return mock, c, nil
}

func (rme *ResponseMockEngine) findMediaTypeContainingNamedExample(
	operation *v3.Operation,
	request *http.Request,
	preferredExample string) (*v3.MediaType, string, bool) {

	mediaTypeString := rme.extractMediaTypeHeader(request)

	for codePairs := operation.Responses.Codes.First(); codePairs != nil; codePairs = codePairs.Next() {
		resp := codePairs.Value()

		if resp.Content != nil {
			responseBody := resp.Content.GetOrZero(mediaTypeString)
			if responseBody == nil {
				responseBody = resp.Content.GetOrZero("application/json")
			}

			if responseBody == nil {
				continue
			}

			_, present := responseBody.Examples.Get(preferredExample)

			if present {
				return responseBody, codePairs.Key(), false
			}
		}
	}

	return nil, "", true
}

func (rme *ResponseMockEngine) findLowestSuccessCode(operation *v3.Operation) string {
	var lowestCode = 299
	if operation.Responses == nil {
		return "404" // no responses defined!
	}
	for codePairs := operation.Responses.Codes.First(); codePairs != nil; codePairs = codePairs.Next() {
		code, _ := strconv.Atoi(codePairs.Key())
		if code < lowestCode && code >= 200 {
			lowestCode = code
		}
	}
	if lowestCode == 299 {
		lowestCode = 200
	}
	return fmt.Sprintf("%d", lowestCode)
}

func (rme *ResponseMockEngine) findBestMediaTypeMatch(
	op *v3.Operation,
	request *http.Request,
	resultCodes []string) (*v3.MediaType, bool) {

	if op.Responses == nil {
		return nil, false
	}

	mediaTypeString := rme.extractMediaTypeHeader(request)

	// Try to find a matching media type in responses matching
	// parameterized result codes
	for _, code := range resultCodes {

		resp := op.Responses.Codes.GetOrZero(code)
		if resp == nil {
			continue
		}
		if resp.Content != nil {
			responseBody := resp.Content.GetOrZero(mediaTypeString)
			if responseBody != nil {
				// try and extract a default JSON response
				return responseBody, false
			} else {
				responseBody = resp.Content.GetOrZero("application/json")
				return responseBody, false
			}
		} else {
			// no content, so try and extract a default JSON response
			return nil, false
		}
	}

	// As a last resort, check if a default response is specified and attempt
	// to use that
	if op.Responses.Default != nil && op.Responses.Default.Content != nil {
		if op.Responses.Default.Content.GetOrZero(mediaTypeString) != nil {
			return op.Responses.Default.Content.GetOrZero(mediaTypeString), false
		}
	}

	return nil, true
}
