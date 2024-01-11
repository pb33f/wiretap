// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package har

import (
	"github.com/pb33f/harhar"
	"github.com/pb33f/libopenapi-validator/errors"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/wiretap/shared"
	"github.com/pb33f/wiretap/validation"
	"github.com/pterm/pterm"
	"strings"
)

type Transaction struct {
	Request  *harhar.Request
	Response *harhar.Response
}

func ValidateHAR(har *harhar.HAR, doc *v3.Document, configFile *shared.WiretapConfiguration) []*errors.ValidationError {

	var validationErrors []*errors.ValidationError

	validator := validation.NewHttpValidator(doc)

	for _, entry := range har.Log.Entries {

		httpRequest, err := harhar.ConvertRequestIntoHttpRequest(entry.Request)

		if err != nil {
			pterm.Error.Printf("error converting request: %s", err.Error())
			return nil
		}

		if configFile.HARPathAllowList != nil {
			for _, allow := range configFile.HARPathAllowList {

				if strings.HasPrefix(httpRequest.URL.Path, allow) {

					path := strings.Replace(httpRequest.URL.Path, allow, "", 1)
					httpRequest.URL.Path = path

					validRequest, requestValidationErrors := validator.ValidateHttpRequest(httpRequest)
					if !validRequest {
						validationErrors = append(validationErrors, requestValidationErrors...)
					} else {
						configFile.Logger.Debug("[HAR] valid request", "path", httpRequest.URL.Path)
					}

					httpResponse := harhar.ConvertResponseIntoHttpResponse(entry.Response)
					validResponse, responseValidationErrors := validator.ValidateHttpResponse(httpRequest, httpResponse)
					if !validResponse {
						validationErrors = append(validationErrors, responseValidationErrors...)
					} else {
						configFile.Logger.Debug("[HAR] valid response", "path", httpRequest.URL.Path)
					}
					break
				}
				pterm.Debug.Printf("[HAR] skipping request: %s\n", httpRequest.URL.Path)
			}
		}
	}

	return validationErrors

}
