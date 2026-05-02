// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package cmd

import (
	"fmt"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel"
	"github.com/pb33f/wiretap/shared"
	"github.com/pb33f/wiretap/specs"
	"github.com/pterm/pterm"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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
				pterm.Debug.Printf("Setting OpenAPI reference base URL to: '%s'\n", u.String())
				docConfig.BaseURL = u
			}
		} else {
			pterm.Debug.Printf("Setting OpenAPI reference base path to: '%s'\n", base)
			docConfig.BasePath = base
		}
	}

	handler := pterm.NewSlogHandler(&pterm.DefaultLogger)
	docConfig.Logger = slog.New(handler)
	pterm.DefaultLogger.Level = pterm.LogLevelError

	return libopenapi.NewDocumentWithConfiguration(specBytes, docConfig)
}

func loadAllSpecs(paths []string, base string) ([]shared.ApiDocument, []specs.LoadError) {
	docs := make([]shared.ApiDocument, 0, len(paths))
	var loadErrors []specs.LoadError

	for _, contract := range paths {
		specBase := base
		if specBase == "" && !strings.HasPrefix(contract, "http://") && !strings.HasPrefix(contract, "https://") {
			specBase = filepath.Dir(contract)
		}
		doc, err := loadOpenAPISpec(contract, specBase)
		if err != nil {
			loadErrors = append(loadErrors, specs.LoadError{Spec: contract, Error: err})
			continue
		}

		docModel, docErr := doc.BuildV3Model()
		if docErr != nil && docModel != nil {
			pterm.Warning.Printf("OpenAPI Specification loaded, but there was an issue detected...\n")
			pterm.Warning.Printf("--> %s\n", docErr.Error())
		}
		if docErr != nil && docModel == nil {
			loadErrors = append(loadErrors, specs.LoadError{Spec: contract, Error: docErr})
			continue
		}

		docs = append(docs, shared.ApiDocument{
			DocumentName:  contract,
			Document:      doc,
			DocumentModel: docModel,
		})
	}

	return docs, loadErrors
}
