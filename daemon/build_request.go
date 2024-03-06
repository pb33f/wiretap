// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package daemon

import (
	"encoding/json"
	"fmt"
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
}

func BuildHttpTransaction(build HttpTransactionConfig) *HttpTransaction {

	var dropHeaders []string
	var injectHeaders map[string]string

	cf := build.TransactionConfig

	// add global headers with injection.
	if cf.Headers != nil {
		dropHeaders = cf.Headers.DropHeaders
		injectHeaders = cf.Headers.InjectHeaders
	}

	// now add path specific headers.
	matchedPaths := config.FindPaths(build.OriginalRequest.URL.Path, cf)
	auth := ""
	if len(matchedPaths) > 0 {
		for _, path := range matchedPaths {
			auth = path.Auth
			if path.Headers != nil {
				dropHeaders = append(dropHeaders, path.Headers.DropHeaders...)
				newInjectHeaders := path.Headers.InjectHeaders
				for key := range injectHeaders {
					newInjectHeaders[key] = injectHeaders[key]
				}
				injectHeaders = newInjectHeaders
			}
			break
		}
	}

	newReq := CloneExistingRequest(CloneRequest{
		Request:       build.NewRequest,
		Protocol:      cf.RedirectProtocol,
		Host:          cf.RedirectHost,
		Port:          cf.RedirectPort,
		DropHeaders:   dropHeaders,
		Auth:          auth,
		InjectHeaders: injectHeaders,
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

	replaced := config.RewritePath(build.NewRequest.URL.Path, cf)
	var newUrl = build.NewRequest.URL
	if replaced != "" {
		var e error
		newUrl, e = url.Parse(replaced)
		if e != nil {
			newUrl = build.NewRequest.URL
			pterm.Error.Printf("major configuration problem: cannot parse URL: `%s`: %s", replaced, e.Error())
		}
		if build.NewRequest.URL.RawQuery != "" {
			newUrl.RawQuery = build.NewRequest.URL.RawQuery
		}
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
	url := fmt.Sprintf("%s://%s", protocol, host)
	if port != "" {
		url += fmt.Sprintf(":%s", port)
	}
	if basepath != "" {
		url += basepath
	}
	if r.URL.Path != "" {
		url += r.URL.Path
	}
	if r.URL.RawQuery != "" {
		url += fmt.Sprintf("?%s", r.URL.RawQuery)
	}
	return url
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
