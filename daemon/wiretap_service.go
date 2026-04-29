// Copyright 2023 Princess Beef Heavy Industries LLC
// SPDX-License-Identifier: AGPL

package daemon

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/pb33f/ranch/bus"
	"github.com/pb33f/ranch/model"
	"github.com/pb33f/ranch/service"
	"github.com/pb33f/ranch/store"
	"github.com/pb33f/wiretap/controls"
	"github.com/pb33f/wiretap/daemon/broadcast"
	"github.com/pb33f/wiretap/daemon/mockproxy"
	"github.com/pb33f/wiretap/daemon/proxy"
	daemonvalidator "github.com/pb33f/wiretap/daemon/validator"
	"github.com/pb33f/wiretap/mock"
	"github.com/pb33f/wiretap/shared"
	"github.com/pb33f/wiretap/validation"
)

const (
	WiretapServiceChan      = "wiretap"
	WiretapBroadcastChan    = "wiretap-broadcast"
	WiretapStaticChangeChan = "wiretap-static-change"
	IncomingHttpRequest     = "incoming-http-request"
)

type WiretapService struct {
	transport        *http.Transport
	serviceCore      service.FabricServiceCore
	broadcastChan    *bus.Channel
	bus              bus.EventBus
	controlsStore    store.BusStore
	transactionStore store.BusStore
	config           *shared.WiretapConfiguration
	fs               http.Handler
	broadcaster      *broadcast.LazyBroadcaster
	validator        *daemonvalidator.Validator
	proxy            *proxy.Handler
	mock             *mockproxy.Handler
	stream           bool
	streamChan       chan []*shared.WiretapValidationError
	reportFile       string
	StaticMockDir    string
}

func NewWiretapService(documents []shared.ApiDocument, config *shared.WiretapConfiguration, storeManager store.Manager) *WiretapService {
	controlsStore := storeManager.CreateStore(controls.ControlServiceChan)
	transactionStore := storeManager.CreateStore(WiretapServiceChan)

	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.MaxIdleConns = 20
	tr.IdleConnTimeout = 30 * time.Second
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	wts := &WiretapService{
		stream:     config.StreamReport,
		reportFile: config.ReportFile,
		// Buffered so short stalls in the stream listener don't block the proxy's
		// hard-error sync path. The non-blocking sends at validate.go drop excess
		// once the buffer fills — report streaming is best-effort, proxying is not.
		streamChan:       make(chan []*shared.WiretapValidationError, 256),
		transport:        tr,
		controlsStore:    controlsStore,
		transactionStore: transactionStore,
		broadcaster:      broadcast.NewLazyBroadcaster(),
		proxy:            proxy.NewHandler(tr),
		mock:             mockproxy.NewHandler(),
		StaticMockDir:    config.StaticMockDir,
	}

	documentValidators := make([]daemonvalidator.DocumentValidator, 0, len(documents))
	for _, document := range documents {
		m, _ := document.Document.BuildV3Model()
		docModel := &m.Model

		documentValidators = append(documentValidators, daemonvalidator.DocumentValidator{
			DocumentName: document.DocumentName,
			Document:     document.Document,
			DocModel:     docModel,
			Validator:    validation.NewHttpValidatorWithConfig(docModel, config.StrictMode),
			MockEngine: mock.NewMockEngineWithConfig(
				docModel,
				config.MockModePretty,
				config.UseAllMockResponseFields,
				config.StrictMode,
				config.HardErrors,
				config.MockBypassValidation),
		})
	}
	wts.validator = daemonvalidator.New(documentValidators)

	// hard-wire the config, change this later if needed.
	wts.config = config

	// listen for violations
	wts.listenForValidationErrors()

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

func (ws *WiretapService) HandleStaticMockResponse(request *model.Request, response *http.Response) {
	if ws.mock == nil {
		ws.mock = mockproxy.NewHandler()
	}
	ws.mock.HandleStaticResponse(request, response, func(resp *http.Response) {
		ws.broadcastResponse(request, BuildResponse(request, resp))
	})
}

func (ws *WiretapService) HandleWebsocketRequest(request *model.Request) {
	ws.handleWebsocketRequest(request)
}
