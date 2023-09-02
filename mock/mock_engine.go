// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package mock

import (
    "errors"
    "fmt"
    libopenapierrs "github.com/pb33f/libopenapi-validator/errors"
    "github.com/pb33f/libopenapi-validator/helpers"
    "github.com/pb33f/libopenapi-validator/paths"
    "github.com/pb33f/libopenapi/datamodel/high/v3"
    "github.com/pb33f/wiretap/validation"
    "net/http"
    "strconv"
    "strings"
)

type ResponseMockEngine struct {
    doc       *v3.Document
    validator validation.HttpValidator
}

func NewMockEngine(document *v3.Document) *ResponseMockEngine {
    return &ResponseMockEngine{
        doc:       document,
        validator: validation.NewHttpValidator(document),
    }
}

func (rme *ResponseMockEngine) extractMediaTypeHeader(request *http.Request) string {
    // extract the content type header from the request.
    contentType := request.Header.Get(helpers.ContentTypeHeader)

    // extract the media type from the content type header.
    mediaTypeSting, _, _ := helpers.ExtractContentType(contentType)

    if mediaTypeSting == "" {
        mediaTypeSting = "application/json"
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
    if len(ops) > 0 {
        op := pathItem.GetOperations()[strings.ToLower(request.Method)]
        return op
    }
    return nil
}

func (rme *ResponseMockEngine) ValidateSecurity(request *http.Request, operation *v3.Operation) error {
    // get out early if there is nothing to do.
    if len(rme.doc.Components.SecuritySchemes) <= 0 {
        return nil
    }

    mustApply := make(map[string][]string)

    // global security
    if len(rme.doc.Security) > 0 {
        for _, securityRequirement := range rme.doc.Security {
            for key, scopes := range securityRequirement.Requirements {
                mustApply[key] = scopes
            }
        }
    }

    // operation security
    if len(operation.Security) > 0 {
        for _, securityRequirement := range operation.Security {
            for key, scopes := range securityRequirement.Requirements {
                mustApply[key] = scopes
            }
        }
    }

    // check if we have any security requirements to apply.
    if len(mustApply) > 0 {
        // locate the security schemes components from the document.

        for scope, _ := range mustApply {

            securityComponent := rme.doc.Components.SecuritySchemes[scope]
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

func (rme *ResponseMockEngine) packErrors(errs []*libopenapierrs.ValidationError) error {
    var err error
    for _, e := range errs {
        err = errors.Join(err, e)
    }
    return err
}

func (rme *ResponseMockEngine) validateRequest(request *http.Request, pathItem *v3.PathItem) (*v3.Operation, error) {

    // get operation
    operation := rme.findOperation(request, pathItem)
    if operation == nil {
        return nil, fmt.Errorf("could not find operation for request: %s (%s)",
            request.URL.Path, request.Method)
    }

    return nil, nil

    // process security
    // check if
    //
    //if len(operation.Security) > 0 {
    //
    //	for _, securityRequirement := range operation.Security {
    //
    //		if securityRequirement.
    //
    //
    //	}
    //
    //
    //}
    //
    //
    //// validate the request against the document.
    //_, errs = rme.validator.ValidateHttpRequest(request)

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

}

func (rme *ResponseMockEngine) findLowestSuccessCode(operation *v3.Operation) string {
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

func (rme *ResponseMockEngine) lookForAppropriateResponse(
    response, defaultResponse *v3.Response, request *http.Request) *v3.MediaType {

    // extract the content type header from the request.
    contentType := request.Header.Get(helpers.ContentTypeHeader)

    // extract the media type from the content type header.
    mediaTypeSting, _, _ := helpers.ExtractContentType(contentType)

    // if there is no media type, let's default to json.
    if mediaTypeSting == "" {
        mediaTypeSting = "application/json"
    }

    // check

    // check if the media type exists in the response.
    responseBody := response.Content[mediaTypeSting]
    if responseBody == nil {

        // look through error codes until we find something we can use, in order of granularity
        resultCodes := []string{"415", "422", "400"}

        for _, code := range resultCodes {
            responseBody = response.Content[code]
            if responseBody != nil {
                return responseBody
            }
        }
    }

    // lets check if this request passes security validation.
    return nil
}
