// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package daemon

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/pb33f/ranch/model"
	"github.com/pb33f/wiretap/shared"
	"net/http"
)

func (ws *WiretapService) broadcastRequestValidationErrors(request *model.Request,
	errors []*shared.WiretapValidationError, transaction *HttpTransaction) {
	id, _ := uuid.NewUUID()
	ht := transaction
	ht.RequestValidation = errors

	ws.broadcastChan.Send(&model.Message{
		Id:            &id,
		DestinationId: request.Id,
		Channel:       WiretapBroadcastChan,
		Destination:   WiretapBroadcastChan,
		Payload:       ht,
		Direction:     model.ResponseDir,
	})
}

func (ws *WiretapService) broadcastRequest(request *model.Request, transaction *HttpTransaction) {
	id, _ := uuid.NewUUID()
	ws.broadcastChan.Send(&model.Message{
		Id:            &id,
		DestinationId: request.Id,
		Channel:       WiretapBroadcastChan,
		Destination:   WiretapBroadcastChan,
		Payload:       transaction,
		Direction:     model.ResponseDir,
	})
}

func (ws *WiretapService) broadcastResponse(request *model.Request, response *http.Response) {
	id, _ := uuid.NewUUID()
	ws.broadcastChan.Send(&model.Message{
		Id:            &id,
		DestinationId: request.Id,
		Channel:       WiretapBroadcastChan,
		Destination:   WiretapBroadcastChan,
		Payload:       BuildResponse(request, response),
		Direction:     model.ResponseDir,
	})
}

func (ws *WiretapService) broadcastResponseError(request *model.Request, response *http.Response, err error) {
	id, _ := uuid.NewUUID()
	title := "Response Error"
	code := 500
	if response != nil {
		title = fmt.Sprintf("Response Error %d", response.StatusCode)
		code = response.StatusCode
	}

	respBodyString, _ := json.Marshal(&shared.WiretapError{
		Title:  title,
		Status: code,
		Detail: err.Error(),
	})

	resp := BuildResponse(request, response)
	resp.Response.Body = string(respBodyString)

	ws.broadcastChan.Send(&model.Message{
		Id:            &id,
		DestinationId: request.Id,
		Error:         err,
		Channel:       WiretapBroadcastChan,
		Destination:   WiretapBroadcastChan,
		Payload:       resp,
		Direction:     model.ResponseDir,
	})
}

func (ws *WiretapService) broadcastResponseValidationErrors(request *model.Request, response *http.Response, errors []*shared.WiretapValidationError) {
	id, _ := uuid.NewUUID()

	ht := BuildResponse(request, response)
	ht.ResponseValidation = errors

	ws.broadcastChan.Send(&model.Message{
		Id:            &id,
		DestinationId: request.Id,
		Channel:       WiretapBroadcastChan,
		Destination:   WiretapBroadcastChan,
		Payload:       ht,
		Direction:     model.ResponseDir,
	})
}
