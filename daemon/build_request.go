// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package daemon

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pb33f/wiretap/config"
	"github.com/pb33f/wiretap/shared"
	"github.com/pterm/pterm"
)

type HttpTransactionConfig struct {
	OriginalRequest   *http.Request
	NewRequest        *http.Request
	ID                *uuid.UUID
	TransactionConfig *shared.WiretapConfiguration
	DropHeaders       []string
	InjectHeaders     map[string]string
	Auth              string
	BasePath          string
	BodyBytes         []byte
}

func BuildHttpTransaction(build HttpTransactionConfig) *HttpTransaction {

	cf := build.TransactionConfig

	dropHeaders := build.DropHeaders
	injectHeaders := build.InjectHeaders
	auth := build.Auth
	basePath := build.BasePath

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

	newReq := CloneExistingRequest(CloneRequest{
		Request:       build.NewRequest,
		Protocol:      cf.RedirectProtocol,
		Host:          cf.RedirectHost,
		BasePath:      basePath,
		Port:          cf.RedirectPort,
		DropHeaders:   dropHeaders,
		Auth:          auth,
		InjectHeaders: injectHeaders,
		BodyBytes:     build.BodyBytes,
	})

	var requestBody []byte

	headers := make(map[string]any)
	for k, v := range newReq.Header {
		headers[k] = v[0]
	}

	cookies := make(map[string]*HttpCookie)
	for _, c := range newReq.Cookies() {
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

	// check if request is a multipart form
	if ct, ok := headers["Content-Type"].(string); ok {
		if strings.Contains(ct, "multipart/form-data") {
			err := newReq.ParseMultipartForm(32 << 2)
			if err != nil {
				pterm.Error.Println(err.Error())
			}
		}
	}

	if newReq.MultipartForm != nil {
		var parts []FormPart
		for i := range newReq.MultipartForm.Value {
			parts = append(parts, FormPart{
				Name:  i,
				Value: newReq.MultipartForm.Value[i],
			})
		}
		for k, fHeaders := range newReq.MultipartForm.File {

			var formFiles []*FormFile

			for z := range fHeaders {
				ff := &FormFile{
					Name:    fHeaders[z].Filename,
					Headers: fHeaders[z].Header,
				}
				formFiles = append(formFiles, ff)
			}

			parts = append(parts, FormPart{
				Name:  k,
				Files: formFiles,
			})
		}
		requestBody, _ = json.Marshal(parts)
	} else {
		requestBody, _ = io.ReadAll(newReq.Body)
	}

	replaced := config.RewritePath(build.NewRequest.URL.Path, newReq, cf)
	var newUrl = build.NewRequest.URL
	if replaced.RewrittenPath != "" {
		var e error
		newUrl, e = url.Parse(replaced.RewrittenPath)
		if e != nil {
			newUrl = build.NewRequest.URL
			pterm.Error.Printf("major configuration problem: cannot parse URL: `%s`: %s", replaced.RewrittenPath, e.Error())
		}
		if build.NewRequest.URL.RawQuery != "" {
			newUrl.RawQuery = build.NewRequest.URL.RawQuery
		}
	}

	// If newUrl has no host (e.g. no path rewriting configured), fill in the
	// destination scheme/host from the cloned request so the UI can display
	// the full destination URL.
	if newUrl.Host == "" && newReq.URL != nil && newReq.URL.Host != "" {
		destUrl := *newUrl
		destUrl.Scheme = newReq.URL.Scheme
		destUrl.Host = newReq.URL.Host
		newUrl = &destUrl
	}

	return &HttpTransaction{
		Id: build.ID.String(),
		Request: &HttpRequest{
			URL:             newUrl.String(),
			Method:          build.NewRequest.Method,
			Path:            newUrl.Path,
			Host:            newUrl.Host,
			Query:           newUrl.RawQuery,
			DroppedHeaders:  dropHeaders,
			InjectedHeaders: injectHeaders,
			OriginalPath:    build.NewRequest.URL.Path,
			Cookies:         cookies,
			Headers:         headers,
			Body:            string(requestBody),
			Timestamp:       time.Now().UnixMilli(),
		},
	}
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
