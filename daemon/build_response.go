// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package daemon

import (
	"bytes"
	"github.com/pb33f/ranch/model"
	"io"
	"net/http"
	"time"
)

func CloneExistingResponse(r *http.Response) *http.Response {
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

func BuildResponse(r *model.Request, response *http.Response) *HttpTransaction {
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
			Timestamp:  time.Now().UnixMilli(),
			Headers:    headers,
			StatusCode: code,
			Body:       string(respBody),
			Cookies:    cookies,
		},
	}
}
