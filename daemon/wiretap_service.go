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
    "github.com/pb33f/wiretap/shared"
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
    config           *shared.WiretapConfiguration
    fs               http.Handler
}

func NewWiretapService(document libopenapi.Document, config *shared.WiretapConfiguration) *WiretapService {
    storeManager := bus.GetBus().GetStoreManager()
    controlsStore := storeManager.CreateStore(controls.ControlServiceChan)
    transactionStore := storeManager.CreateStore(WiretapServiceChan)

    tr := &http.Transport{
        MaxIdleConns:    20,
        IdleConnTimeout: 30 * time.Second,
    }

    wts := &WiretapService{

        transport:        tr,
        controlsStore:    controlsStore,
        transactionStore: transactionStore,
    }
    if document != nil {
        m, _ := document.BuildV3Model()
        wts.document = document
        wts.docModel = &m.Model
    }

    // hard-wire the config, change this later if needed.
    wts.config = config
    
    return wts

}

func (ws *WiretapService) HandleServiceRequest(request *model.Request, core service.FabricServiceCore) {
    switch request.RequestCommand {
    case IncomingHttpRequest:
        ws.handleHttpRequest(request)
    default:
        core.HandleUnknownRequest(request)
    }
}

func (ws *WiretapService) HandleHttpRequest(request *model.Request) {

    ws.handleHttpRequest(request)
}
