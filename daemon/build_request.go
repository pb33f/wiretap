// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: BUSL-1.1

package daemon

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pb33f/wiretap/config"
	"github.com/pb33f/wiretap/shared"
	"github.com/pb33f/wiretap/transaction"
)

type HttpTransactionConfig struct {
	OriginalRequest   *http.Request
	NewRequest        *http.Request
	APIRequest        *http.Request
	DisplayURL        *url.URL
	ID                *uuid.UUID
	TransactionConfig *shared.WiretapConfiguration
	DropHeaders       []string
	InjectHeaders     map[string]string
	Auth              string
	BasePath          string
	BodyBytes         []byte
	SpecConflict      *transaction.SpecConflict
}

func BuildHttpTransaction(build HttpTransactionConfig) *transaction.HttpTransaction {

	cf := build.TransactionConfig

	dropHeaders := build.DropHeaders
	injectHeaders := build.InjectHeaders
	auth := build.Auth

	// If pre-resolved headers were not provided, resolve them now (backward compat).
	// Uses copy-on-write to avoid mutating shared config state.
	if dropHeaders == nil && injectHeaders == nil {
		if cf.Headers != nil {
			dropHeaders = append([]string(nil), cf.Headers.DropHeaders...)
			injectHeaders = mergeInjectHeaders(nil, cf.Headers.InjectHeaders)
		}

		matchedPaths := config.FindPaths(build.OriginalRequest.URL.Path, cf)
		if len(matchedPaths) > 0 {
			var matchedPath *shared.WiretapPathConfig
			matchedPath = config.FindPathWithRewriteId(matchedPaths, build.NewRequest)
			if matchedPath == nil {
				matchedPath = matchedPaths[0]
			}
			auth = matchedPath.Auth
			if matchedPath.Headers != nil {
				dropHeaders = append(dropHeaders, matchedPath.Headers.DropHeaders...)
				injectHeaders = mergeInjectHeaders(matchedPath.Headers.InjectHeaders, injectHeaders)
			}
		}
	}

	newReq := build.NewRequest
	bodyBytes := build.BodyBytes
	if bodyBytes == nil && newReq != nil && newReq.Body != nil {
		bodyBytes, _ = io.ReadAll(newReq.Body)
	}
	if bodyBytes != nil && newReq != nil {
		newReq.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	var requestBody []byte

	headers := buildTransactionHeaders(prepareHeaders(newReq.Header, dropHeaders, injectHeaders, auth, cf.CompiledVariables))

	cookies := make(map[string]*transaction.HttpCookie)
	for _, c := range newReq.Cookies() {
		cookies[c.Name] = &transaction.HttpCookie{
			Value:    c.Value,
			Path:     c.Path,
			Domain:   c.Domain,
			Expires:  c.RawExpires,
			MaxAge:   c.MaxAge,
			Secure:   c.Secure,
			HttpOnly: c.HttpOnly,
		}
	}

	// check if request is a multipart form
	if ct, ok := headers["Content-Type"].(string); ok {
		if newReq != nil && strings.Contains(ct, "multipart/form-data") {
			err := newReq.ParseMultipartForm(32 << 20)
			if err != nil {
				wiretapLogger(cf).Error(err.Error())
			}
		}
	}

	if newReq.MultipartForm != nil {
		var parts []transaction.FormPart
		for i := range newReq.MultipartForm.Value {
			parts = append(parts, transaction.FormPart{
				Name:  i,
				Value: newReq.MultipartForm.Value[i],
			})
		}
		for k, fHeaders := range newReq.MultipartForm.File {

			var formFiles []*transaction.FormFile

			for z := range fHeaders {
				ff := &transaction.FormFile{
					Name:    fHeaders[z].Filename,
					Headers: fHeaders[z].Header,
				}
				formFiles = append(formFiles, ff)
			}

			parts = append(parts, transaction.FormPart{
				Name:  k,
				Files: formFiles,
			})
		}
		requestBody, _ = json.Marshal(parts)
	} else if bodyBytes != nil {
		requestBody = bodyBytes
	} else if newReq.Body != nil {
		requestBody, _ = io.ReadAll(newReq.Body)
	}
	if bodyBytes != nil && newReq != nil {
		newReq.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	newUrl := cloneURL(build.DisplayURL)
	if newUrl == nil && build.APIRequest != nil {
		newUrl = cloneURL(build.APIRequest.URL)
	}
	if newUrl == nil {
		rawURL := ReconstructURL(build.NewRequest, cf.RedirectProtocol, cf.RedirectHost, build.BasePath, cf.RedirectPort)
		var err error
		newUrl, err = url.Parse(rawURL)
		if err != nil {
			newUrl = cloneURL(build.NewRequest.URL)
			wiretapLogger(cf).Error("major configuration problem: cannot parse URL", "url", rawURL, "error", err)
		}
	}

	originalPath := ""
	if build.OriginalRequest != nil && build.OriginalRequest.URL != nil {
		originalPath = build.OriginalRequest.URL.Path
	} else if build.NewRequest != nil && build.NewRequest.URL != nil {
		originalPath = build.NewRequest.URL.Path
	}

	return &transaction.HttpTransaction{
		Id:           build.ID.String(),
		SpecConflict: build.SpecConflict,
		Request: &transaction.HttpRequest{
			URL:             newUrl.String(),
			Method:          build.NewRequest.Method,
			Path:            newUrl.Path,
			Host:            newUrl.Host,
			Query:           newUrl.RawQuery,
			DroppedHeaders:  dropHeaders,
			InjectedHeaders: injectHeaders,
			OriginalPath:    originalPath,
			Cookies:         cookies,
			Headers:         headers,
			Body:            string(requestBody),
			Timestamp:       time.Now().UnixMilli(),
		},
	}
}

func buildTransactionHeaders(source http.Header) map[string]any {
	headers := make(map[string]any, len(source))
	for k, v := range source {
		if len(v) == 0 {
			continue
		}
		headers[k] = v[0]
	}
	return headers
}

func ReconstructURL(r *http.Request, protocol, host, basepath string, port string) string {
	if host == "" {
		host = r.Host
	}
	if protocol == "" {
		protocol = "http"
	}
	var b strings.Builder
	b.Grow(len(protocol) + 3 + len(host) + 1 + len(port) + len(basepath) + len(r.URL.Path) + 1 + len(r.URL.RawQuery))
	b.WriteString(protocol)
	b.WriteString("://")
	b.WriteString(host)
	if port != "" {
		b.WriteByte(':')
		b.WriteString(port)
	}
	if basepath != "" {
		b.WriteString(basepath)
	}
	if r.URL.Path != "" {
		b.WriteString(r.URL.Path)
	}
	if r.URL.RawQuery != "" {
		b.WriteByte('?')
		b.WriteString(r.URL.RawQuery)
	}
	return b.String()
}

func ExtractHeaders(resp *http.Response) map[string][]string {
	headers := make(map[string][]string)
	for k, v := range resp.Header {
		headers[k] = v
	}
	return headers
}

func ReplaceWithVariables(variables map[string]*shared.CompiledVariable, input string) string {
	for x := range variables {
		if variables[x] != nil {
			input = variables[x].CompiledVariable.ReplaceAllString(input, variables[x].VariableValue)
		}
	}
	return input
}
