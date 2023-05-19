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

func cloneRequest(r *http.Request) *http.Request {
	// todo: replace with config/server etc.
	// todo: check query params

	// sniff and replace body.
	b, _ := io.ReadAll(r.Body)
	_ = r.Body.Close()
	r.Body = io.NopCloser(bytes.NewBuffer(b))

	newBaseURL := fmt.Sprintf("http://localhost%s?%s", r.URL.Path, r.URL.RawQuery)
	newReq, _ := http.NewRequest(r.Method, newBaseURL, io.NopCloser(bytes.NewBuffer(b)))
	newReq.Header = r.Header
	return newReq
}

func cloneResponse(r *http.Response) *http.Response {
	// sniff and replace body.
	b, _ := io.ReadAll(r.Body)
	_ = r.Body.Close()
	r.Body = io.NopCloser(bytes.NewBuffer(b))
	resp := &http.Response{
		StatusCode: r.StatusCode,
		Body:       io.NopCloser(bytes.NewBuffer(b)),
		Header:     r.Header,
	}
	return resp
}
