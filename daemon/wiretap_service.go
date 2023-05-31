// Copyright 2023 Princess Beef Heavy Industries LLC
// SPDX-License-Identifier: MIT

package daemon

import (
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/ranch/bus"
	"github.com/pb33f/ranch/model"
	"github.com/pb33f/ranch/service"
	"github.com/pb33f/wiretap/controls"
	"net/http"
	"time"
)

const (
	WiretapServiceChan   = "wiretap"
	WiretapBroadcastChan = "wiretap-broadcast"
	IncomingHttpRequest  = "incoming-http-request"
)

type WiretapService struct {
	transport        *http.Transport
	document         libopenapi.Document
	docModel         *v3.Document
	serviceCore      service.FabricServiceCore
	broadcastChan    *bus.Channel
	bus              bus.EventBus
	controlsStore    bus.BusStore
	transactionStore bus.BusStore
}

func NewWiretapService(document libopenapi.Document) *WiretapService {
	storeManager := bus.GetBus().GetStoreManager()
	controlsStore := storeManager.CreateStore(controls.ControlServiceChan)
	transactionStore := storeManager.CreateStore(WiretapServiceChan)

	tr := &http.Transport{
		MaxIdleConns:    20,
		IdleConnTimeout: 30 * time.Second,
	}
	m, _ := document.BuildV3Model()
	return &WiretapService{
		document:         document,
		docModel:         &m.Model,
		transport:        tr,
		controlsStore:    controlsStore,
		transactionStore: transactionStore,
	}
}

func (ws *WiretapService) HandleServiceRequest(request *model.Request, core service.FabricServiceCore) {
	switch request.RequestCommand {
	case IncomingHttpRequest:
		ws.handleHttpRequest(request, core)
	default:
		core.HandleUnknownRequest(request)
	}
}
