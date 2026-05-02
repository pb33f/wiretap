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
	BodyBytes     []byte
}

func CloneExistingRequest(request CloneRequest) *http.Request {
	// use pre-read body bytes if available, otherwise read from the request body
	var b []byte
	if request.BodyBytes != nil {
		b = request.BodyBytes
	} else {
		b, _ = io.ReadAll(request.Request.Body)
		_ = request.Request.Body.Close()
	}
	request.Request.Body = io.NopCloser(bytes.NewBuffer(b))

	var newURL string
	var newReq *http.Request

	newURL = ReconstructURL(request.Request, request.Protocol, request.Host, request.BasePath, request.Port)

	bodyReader := bytes.NewReader(b)

	// create cloned request
	var err error
	newReq, err = http.NewRequest(request.Request.Method, newURL, bodyReader)

	if err != nil {
		return nil
	}
	getBody := func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(b)), nil
	}
	request.Request.GetBody = getBody
	request.Request.ContentLength = int64(len(b))
	newReq.GetBody = getBody
	newReq.ContentLength = int64(len(b))

	newReq.Header = prepareHeaders(
		request.Request.Header,
		request.DropHeaders,
		request.InjectHeaders,
		request.Auth,
		request.Variables,
	)

	return newReq
}

func prepareHeaders(
	source http.Header,
	dropHeaders []string,
	injectHeaders map[string]string,
	auth string,
	variables map[string]*shared.CompiledVariable,
) http.Header {
	dropSet := make(map[string]struct{}, len(dropHeaders))
	for _, h := range dropHeaders {
		dropSet[strings.ToLower(h)] = struct{}{}
	}

	headers := make(http.Header, len(source)+len(injectHeaders))
	for k, v := range source {
		if _, drop := dropSet[strings.ToLower(k)]; drop || len(v) == 0 {
			continue
		}
		headers.Set(k, v[0])
	}

	for k, v := range injectHeaders {
		headers.Set(k, ReplaceWithVariables(variables, v))
	}

	if auth != "" {
		encoded := base64.StdEncoding.EncodeToString([]byte(ReplaceWithVariables(variables, auth)))
		headers.Set("Authorization", fmt.Sprintf("Basic %s", encoded))
	}

	return headers
}
