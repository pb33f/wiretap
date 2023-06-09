// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package daemon

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func extractHeaders(resp *http.Response) map[string]any {
	headers := make(map[string]any)
	for k, v := range resp.Header {
		headers[k] = v[0]
	}
	return headers
}

func reconstructURL(r *http.Request, protocol, host, port string) string {
	url := fmt.Sprintf("%s://%s", protocol, host)
	if port != "" {
		url += fmt.Sprintf(":%s", port)
	}
	if r.URL.Path != "" {
		url += r.URL.Path
	}
	if r.URL.RawQuery != "" {
		url += fmt.Sprintf("?%s", r.URL.RawQuery)
	}
	return url
}

type CloneRequest struct {
	Request     *http.Request
	Protocol    string
	Host        string
	Port        string
	DropHeaders []string
}

func cloneRequest(request CloneRequest) *http.Request {
	// sniff and replace body.
	b, _ := io.ReadAll(request.Request.Body)
	_ = request.Request.Body.Close()
	request.Request.Body = io.NopCloser(bytes.NewBuffer(b))

	// create cloned request
	newURL := reconstructURL(request.Request, request.Protocol, request.Host, request.Port)
	newReq, _ := http.NewRequest(request.Request.Method, newURL, io.NopCloser(bytes.NewBuffer(b)))

	// copy headers, drop those that are specified.
	for k, v := range request.Request.Header {
		skip := false
		for h := range request.DropHeaders {
			if strings.EqualFold(request.DropHeaders[h], k) {
				skip = true
			}
		}
		if !skip {
			newReq.Header.Set(k, v[0])
		}
	}
	return newReq
}

func cloneResponse(r *http.Response) *http.Response {
	// sniff and replace body.
	var b []byte
	if r == nil {
		return nil // something else went wrong, nothing to do.
	}
	if r.Body != nil {
		b, _ = io.ReadAll(r.Body)
		_ = r.Body.Close()
		r.Body = io.NopCloser(bytes.NewBuffer(b))
	}
	resp := &http.Response{
		StatusCode: r.StatusCode,
		Header:     r.Header,
	}
	if r.Body != nil {
		resp.Body = io.NopCloser(bytes.NewBuffer(b))
	}
	return resp
}
