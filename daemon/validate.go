// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package daemon

import (
	"github.com/pb33f/ranch/model"
	daemonvalidator "github.com/pb33f/wiretap/daemon/validator"
	"github.com/pb33f/wiretap/shared"
	"github.com/pb33f/wiretap/transaction"
	"net/http"
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
