// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package cmd

import (
	"fmt"
	"github.com/pb33f/libopenapi"
	"github.com/pterm/pterm"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func loadOpenAPISpec(contract string) (libopenapi.Document, error) {
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
	return libopenapi.NewDocument(specBytes)
}
