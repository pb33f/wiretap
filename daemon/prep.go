// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package daemon

import (
	"bytes"
	"io"
	"net/http"

	"github.com/pb33f/ranch/model"
	configModel "github.com/pb33f/wiretap/config"
	"github.com/pb33f/wiretap/shared"
)

type PreparedRequest struct {
	Config        *shared.WiretapConfiguration
	NewReq        *http.Request
	APIRequest    *http.Request
	BodyBytes     []byte
	DropHeaders   []string
	InjectHeaders map[string]string
	Auth          string
	IsHardError   bool
	UseMock       bool
	TxnConfig     HttpTransactionConfig
}

func (ws *WiretapService) prepareRequest(request *model.Request) *PreparedRequest {
	configStore, _ := ws.controlsStore.Get(shared.ConfigKey)
	config := configStore.(*shared.WiretapConfiguration)

	if config.Headers == nil || len(config.Headers.DropHeaders) == 0 {
		config.Headers = &shared.WiretapHeaderConfig{
			DropHeaders: []string{},
		}
	}

	dropHeaders, injectHeaders, auth := ws.getHeadersAndAuth(config, request)

	// Read body once. The same bytes are reused for display, validation, and
	// upstream request clones so request preparation owns all body rewind work.
	bodyBytes, _ := io.ReadAll(request.HttpRequest.Body)
	_ = request.HttpRequest.Body.Close()
	request.HttpRequest.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// newReq intentionally has no RedirectBasePath; validator and display paths
	// should match the OpenAPI paths instead of the upstream deployment path.
	newReq := CloneExistingRequest(CloneRequest{
		Request:       request.HttpRequest,
		Protocol:      config.RedirectProtocol,
		Host:          config.RedirectHost,
		Port:          config.RedirectPort,
		DropHeaders:   dropHeaders,
		InjectHeaders: injectHeaders,
		Auth:          auth,
		BodyBytes:     bodyBytes,
	})

	// apiRequest includes RedirectBasePath and is the only request sent upstream.
	apiRequest := CloneExistingRequest(CloneRequest{
		Request:       request.HttpRequest,
		Protocol:      config.RedirectProtocol,
		Host:          config.RedirectHost,
		BasePath:      config.RedirectBasePath,
		Port:          config.RedirectPort,
		DropHeaders:   dropHeaders,
		InjectHeaders: injectHeaders,
		Auth:          auth,
		BodyBytes:     bodyBytes,
	})

	if newReq == nil || apiRequest == nil {
		ws.config.Logger.Error("[wiretap] unable to clone API request, failed", "url", request.HttpRequest.URL.String())
		return nil
	}

	isHardError := configModel.IsHardErrorsSet(apiRequest.URL.Path, config)
	txnConfig := HttpTransactionConfig{
		OriginalRequest:   request.HttpRequest,
		NewRequest:        newReq,
		ID:                request.Id,
		TransactionConfig: config,
		DropHeaders:       dropHeaders,
		InjectHeaders:     injectHeaders,
		Auth:              auth,
		BasePath:          config.RedirectBasePath,
		BodyBytes:         bodyBytes,
	}

	return &PreparedRequest{
		Config:        config,
		NewReq:        newReq,
		APIRequest:    apiRequest,
		BodyBytes:     bodyBytes,
		DropHeaders:   dropHeaders,
		InjectHeaders: injectHeaders,
		Auth:          auth,
		IsHardError:   isHardError,
		UseMock:       ws.config.MockMode || configModel.IncludePathOnMockMode(apiRequest.URL.Path, ws.config),
		TxnConfig:     txnConfig,
	}
}

func mergeInjectHeaders(base, override map[string]string) map[string]string {
	merged := make(map[string]string, len(base)+len(override))
	for k, v := range base {
		merged[k] = v
	}
	for k, v := range override {
		merged[k] = v
	}
	return merged
}

func (ws *WiretapService) getHeadersAndAuth(config *shared.WiretapConfiguration, request *model.Request) ([]string, map[string]string, string) {
	var dropHeaders []string
	var injectHeaders map[string]string

	// copy global headers to avoid mutating shared config state
	if config.Headers != nil {
		dropHeaders = append([]string(nil), config.Headers.DropHeaders...)
		injectHeaders = mergeInjectHeaders(nil, config.Headers.InjectHeaders)
	}

	// now add path specific headers.
	matchedPaths := configModel.FindPaths(request.HttpRequest.URL.Path, config)
	auth := ""
	if len(matchedPaths) > 0 {
		var matchedPath *shared.WiretapPathConfig

		// First check if we have a path matching our RewriteId
		matchedPath = configModel.FindPathWithRewriteId(matchedPaths, request.HttpRequest)

		// Get the first matched value in the list, if we don't have a rewriteId that fits
		if matchedPath == nil {
			matchedPath = matchedPaths[0]
		}

		auth = matchedPath.Auth
		if matchedPath.Headers != nil {
			dropHeaders = append(dropHeaders, matchedPath.Headers.DropHeaders...)
			injectHeaders = mergeInjectHeaders(matchedPath.Headers.InjectHeaders, injectHeaders)
		}
	}

	// apply variable substitution so callers receive already-substituted values
	// (matches what CloneExistingRequest sends upstream)
	if len(config.CompiledVariables) > 0 {
		for k, v := range injectHeaders {
			injectHeaders[k] = ReplaceWithVariables(config.CompiledVariables, v)
		}
		if auth != "" {
			auth = ReplaceWithVariables(config.CompiledVariables, auth)
		}
	}

	return dropHeaders, injectHeaders, auth
}
