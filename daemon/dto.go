// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package daemon

import (
    "bytes"
    "fmt"
    "github.com/pb33f/libopenapi-validator/errors"
    "github.com/pb33f/ranch/model"
    "io"
    "net/http"
)

type HttpRequest struct {
    URL     string            `json:"url,omitempty"`
    Method  string            `json:"method,omitempty"`
    Path    string            `json:"path,omitempty"`
    Query   string            `json:"query,omitempty"`
    Headers map[string]string `json:"headers,omitempty"`
    Body    string            `json:"requestBody,omitempty"`
}

type HttpResponse struct {
    *HttpRequest `json:"httpRequest,omitempty"`
    StatusCode   int    `json:"statusCode,omitempty"`
    Body         string `json:"responseBody,omitempty"`
}

type HttpTransaction struct {
    Request            *HttpRequest              `json:"httpRequest,omitempty"`
    RequestValidation  []*errors.ValidationError `json:"requestValidation,omitempty"`
    Response           *HttpResponse             `json:"httpResponse,omitempty"`
    ResponseValidation []*errors.ValidationError `json:"responseValidation,omitempty"`
    Id                 string                    `json:"id,omitempty"`
}

func buildResponse(r *model.Request, response *http.Response) *HttpTransaction {
    request := buildRequest(r)
    code := 500
    if response != nil {
        code = response.StatusCode
    }

    // sniff and replace response body.
    respBody, _ := io.ReadAll(response.Body)
    _ = response.Body.Close()
    response.Body = io.NopCloser(bytes.NewBuffer(respBody))

    return &HttpTransaction{
        Id:      r.Id.String(),
        Request: request.Request,
        Response: &HttpResponse{
            request.Request,
            code,
            string(respBody),
        },
    }
}

func buildRequest(r *model.Request) *HttpTransaction {
    headers := make(map[string]string)
    for k, v := range r.HttpRequest.Header {
        headers[k] = fmt.Sprint(v)
    }

    // sniff and replace request body.
    requestBody, _ := io.ReadAll(r.HttpRequest.Body)
    _ = r.HttpRequest.Body.Close()
    r.HttpRequest.Body = io.NopCloser(bytes.NewBuffer(requestBody))

    return &HttpTransaction{
        Id: r.Id.String(),
        Request: &HttpRequest{
            URL:     r.HttpRequest.URL.String(),
            Method:  r.HttpRequest.Method,
            Path:    r.HttpRequest.URL.Path,
            Query:   r.HttpRequest.URL.RawQuery,
            Headers: headers,
            Body:    string(requestBody),
        },
    }
}
