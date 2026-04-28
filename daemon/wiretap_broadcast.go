// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package daemon

import (
	"encoding/json"
	"fmt"
	"github.com/pb33f/ranch/bus"
	"github.com/pb33f/ranch/model"
	daemonbroadcast "github.com/pb33f/wiretap/daemon/broadcast"
	"github.com/pb33f/wiretap/shared"
	"github.com/pb33f/wiretap/transaction"
	"net/http"
)

func (ws *WiretapService) setBroadcastChannel(channel *bus.Channel) {
	ws.broadcastChan = channel
	if ws.broadcaster == nil {
		ws.broadcaster = daemonbroadcast.NewLazyBroadcaster()
	}
	ws.broadcaster.Set(daemonbroadcast.NewBroadcaster(channel, WiretapBroadcastChan))
}

func (ws *WiretapService) activeBroadcaster() daemonbroadcast.Broadcaster {
	if ws.broadcaster == nil {
		ws.broadcaster = daemonbroadcast.NewLazyBroadcaster()
	}
	if !ws.broadcaster.Ready() && ws.broadcastChan != nil {
		ws.broadcaster.Set(daemonbroadcast.NewBroadcaster(ws.broadcastChan, WiretapBroadcastChan))
	}
	return ws.broadcaster
}

func (ws *WiretapService) broadcastRequestValidationErrors(request *model.Request,
	errors []*shared.WiretapValidationError, txn *transaction.HttpTransaction) {
	ws.activeBroadcaster().RequestValidationErrors(request, errors, txn)
}

func (ws *WiretapService) broadcastRequest(request *model.Request, txn *transaction.HttpTransaction) {
	ws.activeBroadcaster().Request(request, txn)
}

func (ws *WiretapService) broadcastResponse(request *model.Request, txn *transaction.HttpTransaction) {
	ws.activeBroadcaster().Response(request, txn)
}

func (ws *WiretapService) broadcastResponseError(request *model.Request, response *http.Response, err error) {
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

	ws.activeBroadcaster().ResponseError(request, resp, err)
}

func (ws *WiretapService) broadcastResponseValidationErrors(request *model.Request, txn *transaction.HttpTransaction, errors []*shared.WiretapValidationError) {
	ws.activeBroadcaster().ResponseValidationErrors(request, txn, errors)
}
