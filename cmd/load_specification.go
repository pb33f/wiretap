// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package cmd

import (
	"fmt"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel"
	"github.com/pterm/pterm"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func loadOpenAPISpec(contract, base string) (libopenapi.Document, error) {
	var specBytes []byte

	if strings.HasPrefix(contract, "http://") || strings.HasPrefix(contract, "https://") {
		if docUrl, err := url.Parse(contract); err == nil {
			pterm.Info.Printf("Fetching OpenAPI Specification from URL: '%s'\n", docUrl.String())
			resp, er := http.Get(docUrl.String())
			if er != nil {
				return nil, er
			}
			respBody, e := io.ReadAll(resp.Body)
			if e != nil {
				return nil, e
			}
			if len(respBody) > 0 {
				specBytes = respBody
			}
		}
	} else {

		// not a URL, is it a file?
		var er error
		if _, er = os.Stat(contract); er != nil {
			return nil, er
		}
		specBytes, er = os.ReadFile(contract)
		if er != nil {
			return nil, er
		}
	}
	if len(specBytes) <= 0 {
		return nil, fmt.Errorf("no bytes in OpenAPI Specification")
	}

	docConfig := datamodel.NewDocumentConfiguration()
	docConfig.AllowFileReferences = true
	docConfig.AllowRemoteReferences = true
	if base != "" {
		if strings.HasPrefix(base, "http") {
			u, _ := url.Parse(base)
			if u != nil {
				pterm.Info.Printf("Setting OpenAPI reference base URL to: '%s'\n", u.String())
				docConfig.BaseURL = u
			}
		} else {
			pterm.Info.Printf("Setting OpenAPI reference base path to: '%s'\n", base)
			docConfig.BasePath = base
		}
	}

	handler := pterm.NewSlogHandler(&pterm.DefaultLogger)
	docConfig.Logger = slog.New(handler)
	pterm.DefaultLogger.Level = pterm.LogLevelError

	return libopenapi.NewDocumentWithConfiguration(specBytes, docConfig)
}
