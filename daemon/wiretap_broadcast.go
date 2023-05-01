// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package daemon

import (
    "github.com/google/uuid"
    "github.com/pb33f/libopenapi-validator/errors"
    "github.com/pb33f/ranch/model"
    "net/http"
)

func (ws *WiretapService) broadcastRequestValidationErrors(request *model.Request, errors []*errors.ValidationError) {
    id, _ := uuid.NewUUID()

    ht := buildRequest(request)
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

func (ws *WiretapService) broadcastRequest(request *model.Request) {
    id, _ := uuid.NewUUID()
    ws.broadcastChan.Send(&model.Message{
        Id:            &id,
        DestinationId: request.Id,
        Channel:       WiretapBroadcastChan,
        Destination:   WiretapBroadcastChan,
        Payload:       buildRequest(request),
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
        Payload:       buildResponse(request, response),
        Direction:     model.ResponseDir,
    })
}

func (ws *WiretapService) broadcastResponseError(request *model.Request, response *http.Response, err error) {
    id, _ := uuid.NewUUID()
    ws.broadcastChan.Send(&model.Message{
        Id:            &id,
        DestinationId: request.Id,
        Error:         err,
        Channel:       WiretapBroadcastChan,
        Destination:   WiretapBroadcastChan,
        Payload:       buildResponse(request, response),
        Direction:     model.ResponseDir,
    })
}

func (ws *WiretapService) broadcastResponseValidationErrors(request *model.Request, response *http.Response, errors []*errors.ValidationError) {
    id, _ := uuid.NewUUID()

    ht := buildResponse(request, response)
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
