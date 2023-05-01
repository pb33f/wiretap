// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package daemon

import (
    "bytes"
    "fmt"
    "io"
    "net/http"
)

func extractHeaders(resp *http.Response) map[string]string {
    headers := make(map[string]string)
    for k, v := range resp.Header {
        headers[k] = fmt.Sprint(v)
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
