// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package daemon

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pb33f/wiretap/shared"
)

type CloneRequest struct {
	Request       *http.Request
	Protocol      string
	Host          string
	BasePath      string
	Port          string
	PathTarget    string
	DropHeaders   []string
	InjectHeaders map[string]string
	Auth          string
	Variables     map[string]*shared.CompiledVariable
}

func CloneExistingRequest(request CloneRequest) *http.Request {
	// sniff and replace body.
	b, _ := io.ReadAll(request.Request.Body)
	_ = request.Request.Body.Close()
	request.Request.Body = io.NopCloser(bytes.NewBuffer(b))

	var newURL string
	var newReq *http.Request
	newURL = ReconstructURL(request.Request, request.Protocol, request.Host, request.BasePath, request.Port)

	// create cloned request
	newReq, _ = http.NewRequest(request.Request.Method, newURL, io.NopCloser(bytes.NewBuffer(b)))

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

	// inject headers
	for k, v := range request.InjectHeaders {
		newReq.Header.Set(k, ReplaceWithVariables(request.Variables, v))
	}

	// if the auth value is set, we need to base64 encode it and add it to the header.
	if request.Auth != "" {
		encoded := base64.StdEncoding.EncodeToString([]byte(ReplaceWithVariables(request.Variables, request.Auth)))
		// this will overwrite any existing auth header.
		newReq.Header.Set("Authorization", fmt.Sprintf("Basic %s", encoded))
	}

	return newReq
}
