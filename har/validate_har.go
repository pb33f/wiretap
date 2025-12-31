// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package har

import (
	"github.com/pb33f/harhar"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi-validator/paths"
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

type harValidator struct {
	documentName string
	docModel     *libopenapi.DocumentModel[v3.Document]
	validator    validation.HttpValidator
}

func ValidateHAR(har *harhar.HAR, apiDocumentModels []shared.ApiDocumentModel, configFile *shared.WiretapConfiguration) []*shared.WiretapValidationError {

	var validationErrors []*shared.WiretapValidationError
	validators := make([]harValidator, 0)

	for _, apiDocumentModel := range apiDocumentModels {
		validators = append(validators, harValidator{
			documentName: apiDocumentModel.DocumentName,
			docModel:     apiDocumentModel.DocumentModel,
			validator:    validation.NewHttpValidatorWithConfig(&apiDocumentModel.DocumentModel.Model, configFile.StrictMode),
		})
	}

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

					var httpValidator validation.HttpValidator
					var validatorSpec string

					if len(validators) == 1 {
						httpValidator = validators[0].validator
						validatorSpec = validators[0].documentName
					} else {
						pathFound := false
						for _, hValidator := range validators {
							// Find the first path match between all provided specifications
							pathItem, _, _ := paths.FindPath(httpRequest, &hValidator.docModel.Model, nil)

							if pathItem != nil {
								httpValidator = hValidator.validator
								validatorSpec = hValidator.documentName
								pathFound = true
								break
							}
						}

						// If we haven't found a path, let's pick the first validator to validate against
						if !pathFound && len(validators) > 0 {
							httpValidator = validators[0].validator
							validatorSpec = validators[0].documentName
						} else {
							pterm.Error.Printf("no validators available; a valid specification must be provided in order to perform HAR validation")
							return nil
						}
					}

					validRequest, requestValidationErrors := httpValidator.ValidateHttpRequest(httpRequest)
					if !validRequest {
						validationErrors = append(validationErrors, shared.ConvertValidationErrors(validatorSpec, requestValidationErrors)...)
					} else {
						configFile.Logger.Debug("[HAR] valid request", "path", httpRequest.URL.Path)
					}

					httpResponse := harhar.ConvertResponseIntoHttpResponse(entry.Response)
					validResponse, responseValidationErrors := httpValidator.ValidateHttpResponse(httpRequest, httpResponse)
					if !validResponse {
						validationErrors = append(validationErrors, shared.ConvertValidationErrors(validatorSpec, responseValidationErrors)...)
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
