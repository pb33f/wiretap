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

    // check if we're running static content and create a file-server to handle it.
    if config.StaticDir != "" {
        //wts.fs = gziphandler.GzipHandler(http.FileServer(http.Dir(config.StaticDir)))
        //var err error
        //go func() {
        //
        //	pterm.Info.Println(pterm.LightMagenta(fmt.Sprintf("Serving static content from '%s' on port %s...",
        //		config.StaticDir, config.StaticPort)))
        //
        //	watcher, _ := fsnotify.NewWatcher()
        //	defer watcher.Close()
        //
        //	watchDir := func(path string, fi os.FileInfo, err error) error {
        //		if fi.Mode().IsDir() {
        //			return watcher.Add(path)
        //		}
        //		return nil
        //	}
        //
        //	if wErr := filepath.Walk(config.StaticDir, watchDir); err != nil {
        //		pterm.Fatal.Println(fmt.Sprintf("Error trying to monitor static directory: %s", wErr))
        //	}
        //
        //	go func() {
        //		for {
        //			select {
        //			case event := <-watcher.Events:
        //				if event.Has(fsnotify.Write) {
        //					pterm.Info.Println(pterm.LightMagenta(fmt.Sprintf("[wiretap] reloading static file: %s", event.Name)))
        //				}
        //			case wErr := <-watcher.Errors:
        //				pterm.Error.Println(fmt.Sprintf("[wiretap] static error: %s", wErr.Error()))
        //			}
        //		}
        //	}()
        //
        //	err = http.ListenAndServe(fmt.Sprintf(":%s", config.StaticPort), wts.fs)
        //	if err != nil {
        //		pterm.Fatal.Printf("Fatal error serving static content: %s\n", err.Error())
        //	}
        //
        //}()

    }

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
