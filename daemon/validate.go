// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package daemon

import (
	"net/http"

	"github.com/pb33f/ranch/model"
	daemonvalidator "github.com/pb33f/wiretap/daemon/validator"
	"github.com/pb33f/wiretap/shared"
	"github.com/pb33f/wiretap/specs"
	"github.com/pb33f/wiretap/transaction"
	wiretapValidation "github.com/pb33f/wiretap/validation"
)

func (ws *WiretapService) getValidatorForRequest(request *model.Request) *daemonvalidator.DocumentValidator {
	if ws.validator == nil {
		return nil
	}
	return ws.validator.GetValidatorForRequest(request)
}

func (ws *WiretapService) getValidatorForHTTPRequest(request *http.Request) *daemonvalidator.DocumentValidator {
	if ws.validator == nil {
		return nil
	}
	return ws.validator.GetValidatorForHTTPRequest(request)
}

func (ws *WiretapService) getValidatorAndRequestForHTTPRequest(request *http.Request) (*daemonvalidator.DocumentValidator, *http.Request) {
	if ws.validator == nil {
		return nil, nil
	}
	return ws.validator.GetValidatorAndRequestForHTTPRequest(request)
}

func (ws *WiretapService) getRouteMatchForHTTPRequest(request *http.Request) *daemonvalidator.RouteMatch {
	if ws.validator == nil {
		return nil
	}
	return ws.validator.GetRouteMatchForHTTPRequest(request)
}

func (ws *WiretapService) ValidateResponse(
	request *model.Request,
	returnedResponse *http.Response,
	preReadBody ...[]byte) []*shared.WiretapValidationError {
	validationRequest := request.HttpRequest
	return ws.ValidateResponseForRequest(request, validationRequest, returnedResponse, preReadBody...)
}

func (ws *WiretapService) ValidateResponseForRequest(
	request *model.Request,
	validationRequest *http.Request,
	returnedResponse *http.Response,
	preReadBody ...[]byte) []*shared.WiretapValidationError {

	var validationErrors, cleanedErrors []*shared.WiretapValidationError
	if ws.validator != nil {
		validationErrors, cleanedErrors = ws.validator.ValidateResponseForRequest(validationRequest, returnedResponse)
	}

	var txn *transaction.HttpTransaction
	if len(preReadBody) > 0 {
		txn = BuildResponseFromBytes(request, returnedResponse, preReadBody[0])
	} else {
		txn = BuildResponse(request, returnedResponse)
	}
	if len(cleanedErrors) > 0 {
		txn.ResponseValidation = cleanedErrors
	}
	ws.transactionStore.Put(request.Id.String(), txn, nil)

	if len(cleanedErrors) > 0 {
		sendToStreamChan(ws, cleanedErrors)
		ws.broadcastResponseValidationErrors(request, txn, cleanedErrors)
	} else {
		ws.broadcastResponse(request, txn)
	}
	return validationErrors
}

// sendToStreamChan delivers validation errors to the stream listener without
// blocking the caller. If the buffered channel is full (listener stalled or
// never started) the send is dropped. Report streaming is best-effort; the
// synchronous hard-error path must never deadlock on a consumer.
func sendToStreamChan(ws *WiretapService, errs []*shared.WiretapValidationError) {
	select {
	case ws.streamChan <- errs:
	default:
		if ws.config != nil && ws.config.Logger != nil {
			ws.config.Logger.Debug("[wiretap] stream channel full; dropping validation errors from stream report")
		}
	}
}

func (ws *WiretapService) ValidateRequest(
	modelRequest *model.Request,
	httpRequest *http.Request,
	txnConfig ...HttpTransactionConfig) []*shared.WiretapValidationError {

	var cleanedErrors []*shared.WiretapValidationError
	if ws.validator != nil {
		cleanedErrors = ws.validator.ValidateRequest(modelRequest, httpRequest)
	}

	// record results
	var buildTransConfig HttpTransactionConfig
	if len(txnConfig) > 0 {
		buildTransConfig = txnConfig[0]
	} else {
		buildTransConfig = HttpTransactionConfig{
			OriginalRequest:   modelRequest.HttpRequest,
			NewRequest:        httpRequest,
			ID:                modelRequest.Id,
			TransactionConfig: ws.config,
		}
	}
	buildTransConfig.SpecConflict = ws.specConflictForRequest(httpRequest)

	txn := BuildHttpTransaction(buildTransConfig)
	if len(cleanedErrors) > 0 {
		txn.RequestValidation = cleanedErrors
	}
	ws.transactionStore.Put(modelRequest.Id.String(), txn, nil)

	// broadcast what we found.
	if len(cleanedErrors) > 0 {
		sendToStreamChan(ws, cleanedErrors)
		ws.broadcastRequestValidationErrors(modelRequest, cleanedErrors, txn)
	} else {
		ws.broadcastRequest(modelRequest, txn)
	}
	return cleanedErrors
}

func (ws *WiretapService) specConflictForRequest(request *http.Request) *transaction.SpecConflict {
	routeMatch := ws.getRouteMatchForHTTPRequest(request)
	if routeMatch == nil || routeMatch.Document == nil || routeMatch.EffectiveRoutePath == "" {
		return nil
	}
	entries := ws.routeConflicts.Lookup(request.Method, routeMatch.EffectiveRoutePath)
	if len(entries) == 0 {
		return nil
	}
	if request == nil || request.URL == nil {
		return nil
	}
	requestPath := request.URL.EscapedPath()

	matchedSpec := routeMatch.Document.DocumentName
	conflictSpecs := make([]string, 0, len(entries))
	seenSpecs := make(map[string]struct{})
	var selected *specs.RouteConflict
	for i := range entries {
		entry := entries[i]
		if entry.MatchedSpec != "" && entry.MatchedSpec != matchedSpec {
			continue
		}
		if entry.ConflictRoutePath != "" && !wiretapValidation.RoutePathMatches(entry.ConflictRoutePath, requestPath) {
			continue
		}
		if selected == nil {
			selected = &entry
		}
		if entry.ConflictSpec == "" {
			continue
		}
		if _, ok := seenSpecs[entry.ConflictSpec]; ok {
			continue
		}
		seenSpecs[entry.ConflictSpec] = struct{}{}
		conflictSpecs = append(conflictSpecs, entry.ConflictSpec)
	}
	if selected == nil {
		return nil
	}
	if len(conflictSpecs) == 0 && selected.ConflictSpec != "" {
		conflictSpecs = append(conflictSpecs, selected.ConflictSpec)
	}

	return &transaction.SpecConflict{
		MatchedSpec:   matchedSpec,
		ConflictSpecs: conflictSpecs,
		Path:          routeMatch.MatchedPath,
		RoutePath:     routeMatch.EffectiveRoutePath,
		Method:        request.Method,
		Kind:          string(selected.Kind),
	}
}
