// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package daemon

import (
	"bytes"
	"github.com/pb33f/libopenapi-validator/errors"
	"github.com/pb33f/ranch/model"
	"io"
	"net/http"
)

type HttpCookie struct {
	Value   string `json:"value,omitempty"`
	Path    string `json:"path,omitempty"`
	Domain  string `json:"domain,omitempty"`
	Expires string `json:"expires,omitempty"`
	// MaxAge=0 means no 'Max-Age' attribute specified.
	// MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'
	// MaxAge>0 means Max-Age attribute present and given in seconds
	MaxAge   int  `json:"maxAge,omitempty"`
	Secure   bool `json:"secure,omitempty"`
	HttpOnly bool `json:"httpOnly,omitempty"`
}

type HttpRequest struct {
	URL     string                 `json:"url,omitempty"`
	Method  string                 `json:"method,omitempty"`
	Path    string                 `json:"path,omitempty"`
	Query   string                 `json:"query,omitempty"`
	Headers map[string]any         `json:"headers,omitempty"`
	Body    string                 `json:"requestBody,omitempty"`
	Cookies map[string]*HttpCookie `json:"cookies,omitempty"`
}

type HttpResponse struct {
	Headers    map[string]any         `json:"headers,omitempty"`
	StatusCode int                    `json:"statusCode,omitempty"`
	Body       string                 `json:"responseBody,omitempty"`
	Cookies    map[string]*HttpCookie `json:"cookies,omitempty"`
}

type HttpTransaction struct {
	Request            *HttpRequest              `json:"httpRequest,omitempty"`
	RequestValidation  []*errors.ValidationError `json:"requestValidation,omitempty"`
	Response           *HttpResponse             `json:"httpResponse,omitempty"`
	ResponseValidation []*errors.ValidationError `json:"responseValidation,omitempty"`
	Id                 string                    `json:"id,omitempty"`
}

func buildResponse(r *model.Request, response *http.Response) *HttpTransaction {
	code := 500
	headers := make(map[string]any)
	cookies := make(map[string]*HttpCookie)
	var respBody []byte

	if response != nil {
		code = response.StatusCode
		for k, v := range response.Header {
			headers[k] = v[0]
		}

		for _, c := range response.Cookies() {
			cookies[c.Name] = &HttpCookie{
				Value:    c.Value,
				Path:     c.Path,
				Domain:   c.Domain,
				Expires:  c.RawExpires,
				MaxAge:   c.MaxAge,
				Secure:   c.Secure,
				HttpOnly: c.HttpOnly,
			}
		}

		// sniff and replace response body.
		respBody, _ = io.ReadAll(response.Body)
		_ = response.Body.Close()
		response.Body = io.NopCloser(bytes.NewBuffer(respBody))
	}
	return &HttpTransaction{
		Id: r.Id.String(),
		Response: &HttpResponse{
			headers,
			code,
			string(respBody),
			cookies,
		},
	}
}

func buildRequest(r *model.Request) *HttpTransaction {
	headers := make(map[string]any)
	for k, v := range r.HttpRequest.Header {
		headers[k] = v[0]
	}

	cookies := make(map[string]*HttpCookie)
	for _, c := range r.HttpRequest.Cookies() {
		cookies[c.Name] = &HttpCookie{
			Value:    c.Value,
			Path:     c.Path,
			Domain:   c.Domain,
			Expires:  c.RawExpires,
			MaxAge:   c.MaxAge,
			Secure:   c.Secure,
			HttpOnly: c.HttpOnly,
		}
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
			Cookies: cookies,
			Headers: headers,
			Body:    string(requestBody),
		},
	}
}
