// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package mock

import (
	"encoding/json"
	"errors"
	"fmt"
	libopenapierrs "github.com/pb33f/libopenapi-validator/errors"
	"github.com/pb33f/libopenapi-validator/helpers"
	"github.com/pb33f/libopenapi-validator/paths"
	"github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/renderer"
	"github.com/pb33f/wiretap/shared"
	"github.com/pb33f/wiretap/validation"
	"net/http"
	"strconv"
	"strings"
)

type ResponseMockEngine struct {
	doc        *v3.Document
	validator  validation.HttpValidator
	mockEngine *renderer.MockGenerator
	pretty     bool
}

func NewMockEngine(document *v3.Document, pretty bool) *ResponseMockEngine {
	me := renderer.NewMockGenerator(renderer.JSON)
	if pretty {
		me.SetPretty()
	}
	return &ResponseMockEngine{
		doc:        document,
		validator:  validation.NewHttpValidator(document),
		mockEngine: me,
		pretty:     pretty,
	}
}

func (rme *ResponseMockEngine) GenerateResponse(request *http.Request) ([]byte, int, error) {
	return rme.runWorkflow(request)
}

func (rme *ResponseMockEngine) ValidateSecurity(request *http.Request, operation *v3.Operation) error {
	// get out early if there is nothing to do.
	if rme.doc.Components.SecuritySchemes.Len() <= 0 {
		return nil
	}

	mustApply := make(map[string][]string)

	// global security
	if len(rme.doc.Security) > 0 {
		for _, securityRequirement := range rme.doc.Security {
			for securityPairs := securityRequirement.Requirements.First(); securityPairs != nil; securityPairs = securityPairs.Next() {
				key := securityPairs.Key()
				scopes := securityPairs.Value()
				mustApply[key] = scopes
			}
		}
	}

	// operation security
	if len(operation.Security) > 0 {
		for _, securityRequirement := range operation.Security {
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

		for scope, _ := range mustApply {

			securityComponent := rme.doc.Components.SecuritySchemes.GetOrZero(scope)
			if securityComponent != nil {

				// check if we have a security scheme that matches the type.
				switch securityComponent.Type {
				case "http":
					// check if we have a bearer scheme.
					if securityComponent.Scheme == "bearer" || securityComponent.Scheme == "basic" {
						// check if we have a bearer token.
						if request.Header.Get("Authorization") == "" {
							return fmt.Errorf("bearer token not found, no `Authorization` header found in request")
						}
					}

				case "apiKey":
					// check if the api key is being used in the header
					if securityComponent.In == "header" {
						// check if we have a bearer token.
						if request.Header.Get(securityComponent.Name) == "" {
							return fmt.Errorf("apiKey not found, no `%s` header found in request",
								securityComponent.Name)
						}
					}
					if securityComponent.In == "query" {
						if request.URL.Query().Get(securityComponent.Name) == "" {
							return fmt.Errorf("apiKey not found, no `%s` query parameter found in request",
								securityComponent.Name)
						}
					}
					if securityComponent.In == "cookie" {
						cookie, _ := request.Cookie(securityComponent.Name)
						if cookie == nil {
							return fmt.Errorf("apiKey not found, no `%s` cookie found in request",
								securityComponent.Name)
						}
					}
				}
			}
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
		mediaTypeSting = "application/json" // default
	}

	return mediaTypeSting
}

func (rme *ResponseMockEngine) findPath(request *http.Request) (*v3.PathItem, error) {
	path, errs, _ := paths.FindPath(request, rme.doc)
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
		mt, _ := rme.lookForResponseCodes(operation, request, []string{"401"})
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
	_, validationErrors := rme.validator.ValidateHttpRequest(request)
	if len(validationErrors) > 0 {
		mt, _ := rme.lookForResponseCodes(operation, request, []string{"422", "400"})
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

	// get the lowest success code
	lo := rme.findLowestSuccessCode(operation)

	// find the lowest success code.
	mt, noMT := rme.lookForResponseCodes(operation, request, []string{lo})
	if mt == nil && noMT {
		mtString := rme.extractMediaTypeHeader(request)
		return rme.buildError(
			415,
			"Media type not supported",
			fmt.Sprintf("The media type requested '%s' is not supported by this operation", mtString),
			"build_mock_error",
		), 415, nil
	}

	mock, mockErr := rme.mockEngine.GenerateMock(mt, rme.extractPreferred(request))
	if mockErr != nil {
		return rme.buildError(
			422,
			"Unable to build mock (422)",
			fmt.Sprintf("Errors occurred while generating an error 422 mock response: %s",
				errors.Join(err, mockErr)),
			"build_mock_error",
		), 422, mockErr
	}
	c, _ := strconv.Atoi(lo)
	return mock, c, nil
}

func (rme *ResponseMockEngine) findLowestSuccessCode(operation *v3.Operation) string {
	var lowestCode = 299

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

func (rme *ResponseMockEngine) lookForResponseCodes(
	op *v3.Operation,
	request *http.Request,
	resultCodes []string) (*v3.MediaType, bool) {

	mediaTypeString := rme.extractMediaTypeHeader(request)

	// check if the media type exists in the response.
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
		}
	}

	if op.Responses.Default != nil && op.Responses.Default.Content != nil {
		if op.Responses.Default.Content.GetOrZero(mediaTypeString) != nil {
			return op.Responses.Default.Content.GetOrZero(mediaTypeString), false
		}
	}

	return nil, true
}
