// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package daemon

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

func extractHeaders(resp *http.Response) map[string]any {
	headers := make(map[string]any)
	for k, v := range resp.Header {
		headers[k] = v[0]
	}
	return headers
}

func cloneRequest(r *http.Request, protocol, host, port string) *http.Request {
	// todo: replace with config/server etc.
	// todo: check query params

	// sniff and replace body.
	b, _ := io.ReadAll(r.Body)
	_ = r.Body.Close()
	r.Body = io.NopCloser(bytes.NewBuffer(b))

	pattern := "%s://%s:%s?%s"
	urlString := fmt.Sprintf(pattern, protocol, host, port, r.URL.RawQuery)
	if port == "" {
		pattern = "%s://%s?%s"
		urlString = fmt.Sprintf(pattern, protocol, host, r.URL.RawQuery)
	}
	if r.URL.RawQuery == "" {
		pattern = "%s://%s:%s"
		urlString = fmt.Sprintf(pattern, protocol, host, port)
		if port == "" {
			pattern = "%s://%s"
			urlString = fmt.Sprintf(pattern, protocol, host)
		}
	}

	newBaseURL := urlString
	newReq, _ := http.NewRequest(r.Method, newBaseURL, io.NopCloser(bytes.NewBuffer(b)))
	newReq.Header = r.Header
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
