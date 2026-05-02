// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package daemon

import (
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"text/template"

	"github.com/pb33f/ranch/model"
	"github.com/pb33f/wiretap/daemon/mockproxy"
	"github.com/pb33f/wiretap/daemon/proxy"
	"github.com/pb33f/wiretap/shared"
)

//go:embed templates/socket-include.html
var staticTemplate string

var parsedStaticTemplate = template.Must(template.New("index").Parse(staticTemplate))

type staticTemplateModel struct {
	OriginalContent string
	WebSocketPort   string
}

func (ws *WiretapService) handleHttpRequest(request *model.Request) {

	// determine if this is a request for a file or not.
	if ws.config.StaticDir != "" {
		fp := filepath.Join(ws.config.StaticDir, request.HttpRequest.URL.Path)

		isRoot := false
		// check if this is a static path catch-all
		if len(ws.config.StaticPathsCompiled) > 0 {
			for key := range ws.config.StaticPathsCompiled {
				if ws.config.StaticPathsCompiled[key].Match(request.HttpRequest.URL.Path) {
					fp = filepath.Join(ws.config.StaticDir, ws.config.StaticIndex)
					isRoot = true
					break
				}
			}
		}

		// check if this is a root request
		if fp == ws.config.StaticDir {
			isRoot = true
			fp = filepath.Join(ws.config.StaticDir, "index.html")
		}
		localStat, _ := os.Stat(fp)
		if localStat != nil {

			if isRoot {

				// if this root, we need to modify the index to inject some JS.
				tmpFile, _ := os.CreateTemp("", "index.html")
				defer os.Remove(tmpFile.Name())

				indexBytes, _ := os.ReadFile(fp)

				// prep a model
				m := staticTemplateModel{
					OriginalContent: string(indexBytes),
					WebSocketPort:   ws.config.WebSocketPort,
				}

				// execute the new template
				_ = parsedStaticTemplate.Execute(tmpFile, m)

				ws.config.Logger.Info("[wiretap] static file request", "url", request.HttpRequest.URL.String(), "code", 200)

				// serve it.
				http.ServeFile(request.HttpResponseWriter, request.HttpRequest, tmpFile.Name())
				return
			}

			if !localStat.IsDir() {

				ws.config.Logger.Info("[wiretap] static file request", "url", request.HttpRequest.URL.String(), "code", 200)

				http.ServeFile(request.HttpResponseWriter, request.HttpRequest, fp)
				return
			}
		}
	}
	prep := ws.prepareRequest(request)
	if prep == nil {
		return
	}

	ws.config.Logger.Info("[wiretap] handling API request", "url", request.HttpRequest.URL.String())

	// short-circuit if we're using mock mode, there is no API call to make.
	if prep.UseMock {
		ws.config.Logger.Info("MockMode enabled; skipping validation")
		if ws.mock == nil {
			ws.mock = mockproxy.NewHandler()
		}
		ws.mock.Handle(request, &mockproxy.PreparedRequest{
			Config:      prep.Config,
			NewReq:      prep.NewReq,
			IsHardError: prep.IsHardError,
			ValidateRequest: func() []*shared.WiretapValidationError {
				return ws.ValidateRequest(request, prep.NewReq, prep.TxnConfig)
			},
			GenerateMock: func(httpReq *http.Request) ([]byte, int, error) {
				docValidator, mockReq := ws.getValidatorAndRequestForHTTPRequest(httpReq)
				if docValidator != nil {
					return docValidator.MockEngine.GenerateResponse(mockReq)
				}
				return nil, http.StatusInternalServerError,
					fmt.Errorf("mock engine has not been initialized; configure an OpenAPI specification to use this option")
			},
			BroadcastResponse: func(response *http.Response) {
				ws.broadcastResponse(request, BuildResponse(request, response))
			},
		})
		return
	}

	if ws.proxy == nil {
		ws.proxy = proxy.NewHandler(ws.transport)
	}
	ws.proxy.Handle(request, &proxy.PreparedRequest{
		Config:      prep.Config,
		NewReq:      prep.NewReq,
		APIRequest:  prep.APIRequest,
		BodyBytes:   prep.BodyBytes,
		ControlPath: prep.ControlPath,
		IsHardError: prep.IsHardError,
		Validator: proxyValidator{
			validateRequest: func() []*shared.WiretapValidationError {
				return ws.ValidateRequest(request, prep.NewReq, prep.TxnConfig)
			},
			validateResponse: func(response *http.Response, body []byte) []*shared.WiretapValidationError {
				return ws.ValidateResponseForRequest(request, prep.NewReq, response, body)
			},
		},
		BroadcastResponseError: func(response *http.Response, err error) {
			ws.broadcastResponseError(request, CloneExistingResponse(response), err)
		},
	})
}

type proxyValidator struct {
	validateRequest  func() []*shared.WiretapValidationError
	validateResponse func(*http.Response, []byte) []*shared.WiretapValidationError
}

func (v proxyValidator) ValidateRequest() []*shared.WiretapValidationError {
	if v.validateRequest == nil {
		return nil
	}
	return v.validateRequest()
}

func (v proxyValidator) ValidateResponse(response *http.Response, body []byte) []*shared.WiretapValidationError {
	if v.validateResponse == nil {
		return nil
	}
	return v.validateResponse(response, body)
}
