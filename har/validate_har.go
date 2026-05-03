// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: BUSL-1.1

package har

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/pb33f/harific/motor"
	harModel "github.com/pb33f/harific/motor/model"
	"github.com/pb33f/wiretap/shared"
	"github.com/pb33f/wiretap/validation"
)

type Transaction struct {
	Request  *harModel.Request
	Response *harModel.Response
}

type ValidationResult struct {
	Errors       []*shared.WiretapValidationError
	MessageCount int
	Err          error
}

func ValidateHAR(path string, apiDocumentModels []shared.ApiDocumentModel, configFile *shared.WiretapConfiguration) []*shared.WiretapValidationError {
	return ValidateHARWithResult(path, apiDocumentModels, configFile).Errors
}

func ValidateHARWithResult(path string, apiDocumentModels []shared.ApiDocumentModel, configFile *shared.WiretapConfiguration) ValidationResult {
	logger := slog.Default()
	if configFile.Logger != nil {
		logger = configFile.Logger
	}

	var validationErrors []*shared.WiretapValidationError
	validators := make([]validation.DocumentValidator, 0, len(apiDocumentModels))

	for _, apiDocumentModel := range apiDocumentModels {
		validators = append(validators, validation.DocumentValidator{
			DocumentName: apiDocumentModel.DocumentName,
			DocModel:     &apiDocumentModel.DocumentModel.Model,
			Validator:    validation.NewHttpValidatorWithConfig(&apiDocumentModel.DocumentModel.Model, configFile.StrictMode),
		})
	}
	router := validation.NewSpecRouter(validators)

	streamer, err := NewHARStreamer(path, motor.DefaultStreamerOptions())
	if err != nil {
		logger.Error("error creating HAR streamer", "error", err)
		return ValidationResult{Err: fmt.Errorf("create HAR streamer: %w", err)}
	}
	defer streamer.Close()

	ctx := context.Background()
	if err = streamer.Initialize(ctx); err != nil {
		logger.Error("error initializing HAR streamer", "error", err)
		return ValidationResult{Err: fmt.Errorf("initialize HAR streamer: %w", err)}
	}

	index := streamer.GetIndex()
	if index == nil || index.TotalEntries == 0 {
		return ValidationResult{Errors: validationErrors}
	}
	if len(configFile.HARPathAllowList) == 0 {
		return ValidationResult{Errors: validationErrors}
	}

	results, err := streamAllowedHAREntries(ctx, streamer, configFile.HARPathAllowList)
	if err != nil {
		logger.Error("error streaming HAR file", "error", err)
		return ValidationResult{Err: fmt.Errorf("stream HAR file: %w", err)}
	}

	messageCount := 0
	for result := range results {
		if result.Error != nil {
			logger.Error("error streaming HAR entry", "error", result.Error)
			return ValidationResult{
				Errors:       validationErrors,
				MessageCount: messageCount,
				Err:          fmt.Errorf("stream HAR entry: %w", result.Error),
			}
		}
		if result.Entry == nil {
			continue
		}

		httpRequest, err := harModel.ConvertRequestIntoHttpRequest(result.Entry.Request)

		if err != nil {
			logger.Error("error converting request", "error", err)
			return ValidationResult{Err: fmt.Errorf("convert HAR request: %w", err)}
		}

		path, ok := rewriteHARPath(httpRequest.URL.Path, configFile.HARPathAllowList)
		if !ok {
			logger.Debug("[HAR] skipping request", "path", httpRequest.URL.Path)
			continue
		}
		httpRequest.URL.Path = path

		if result.Entry.Request.Method != "" {
			messageCount++
		}
		if result.Entry.Response.StatusCode > 0 {
			messageCount++
		}

		docValidator := router.Resolve(httpRequest)
		if docValidator == nil {
			logger.Error("no validators available; a valid specification must be provided in order to perform HAR validation")
			return ValidationResult{Err: errors.New("no validators available")}
		}

		validRequest, requestValidationErrors := docValidator.Validator.ValidateHttpRequest(httpRequest)
		if !validRequest {
			validationErrors = append(validationErrors, shared.ConvertValidationErrors(docValidator.DocumentName, requestValidationErrors)...)
		} else {
			configFile.Logger.Debug("[HAR] valid request", "path", httpRequest.URL.Path)
		}

		httpResponse := harModel.ConvertResponseIntoHttpResponse(result.Entry.Response)
		validResponse, responseValidationErrors := docValidator.Validator.ValidateHttpResponse(httpRequest, httpResponse)
		if !validResponse {
			validationErrors = append(validationErrors, shared.ConvertValidationErrors(docValidator.DocumentName, responseValidationErrors)...)
		} else {
			configFile.Logger.Debug("[HAR] valid response", "path", httpRequest.URL.Path)
		}
	}

	return ValidationResult{Errors: validationErrors, MessageCount: messageCount}
}

func streamAllowedHAREntries(
	ctx context.Context,
	streamer motor.HARStreamer,
	allowList []string,
) (<-chan motor.StreamResult, error) {
	return streamer.StreamFiltered(ctx, func(metadata *motor.EntryMetadata) bool {
		if metadata == nil {
			return false
		}
		path := metadata.URL
		if parsed, err := url.Parse(metadata.URL); err == nil && parsed.Path != "" {
			path = parsed.Path
		}
		_, ok := rewriteHARPath(path, allowList)
		return ok
	})
}

func rewriteHARPath(path string, allowList []string) (string, bool) {
	if allowList == nil {
		return path, true
	}
	for _, allow := range allowList {
		if strings.HasPrefix(path, allow) {
			return strings.Replace(path, allow, "", 1), true
		}
	}
	return path, false
}
