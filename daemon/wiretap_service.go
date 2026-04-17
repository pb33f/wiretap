// Copyright 2023 Princess Beef Heavy Industries LLC
// SPDX-License-Identifier: AGPL

package daemon

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/ranch/bus"
	"github.com/pb33f/ranch/model"
	"github.com/pb33f/ranch/service"
	"github.com/pb33f/wiretap/controls"
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

type documentValidator struct {
	documentName string
	document     libopenapi.Document
	docModel     *v3.Document
	validator    validation.HttpValidator
	mockEngine   *mock.ResponseMockEngine
}
type WiretapService struct {
	transport          *http.Transport
	serviceCore        service.FabricServiceCore
	broadcastChan      *bus.Channel
	bus                bus.EventBus
	controlsStore      bus.BusStore
	transactionStore   bus.BusStore
	config             *shared.WiretapConfiguration
	fs                 http.Handler
	documentValidators []documentValidator
	stream             bool
	streamChan   chan []*shared.WiretapValidationError
	reportFile   string
	StaticMockDir      string
}

func NewWiretapService(documents []shared.ApiDocument, config *shared.WiretapConfiguration) *WiretapService {
	storeManager := bus.GetBus().GetStoreManager()
	controlsStore := storeManager.CreateStore(controls.ControlServiceChan)
	transactionStore := storeManager.CreateStore(WiretapServiceChan)

	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.MaxIdleConns = 20
	tr.IdleConnTimeout = 30 * time.Second
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	wts := &WiretapService{
		stream:           config.StreamReport,
		reportFile:       config.ReportFile,
		// Buffered so short stalls in the stream listener don't block the proxy's
		// hard-error sync path. The non-blocking sends at validate.go drop excess
		// once the buffer fills — report streaming is best-effort, proxying is not.
		streamChan:       make(chan []*shared.WiretapValidationError, 256),
		transport:        tr,
		controlsStore:    controlsStore,
		transactionStore: transactionStore,
		StaticMockDir:    config.StaticMockDir,
	}

	for _, document := range documents {
		m, _ := document.Document.BuildV3Model()
		docModel := &m.Model

		wts.documentValidators = append(wts.documentValidators, documentValidator{
			documentName: document.DocumentName,
			document:     document.Document,
			docModel:     docModel,
			validator:    validation.NewHttpValidatorWithConfig(docModel, config.StrictMode),
			mockEngine: mock.NewMockEngineWithConfig(
				docModel,
				config.MockModePretty,
				config.UseAllMockResponseFields,
				config.StrictMode,
				config.HardErrors,
				config.MockBypassValidation),
		})
	}

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
	ws.handleStaticMockResponse(request, response)
}

func (ws *WiretapService) HandleWebsocketRequest(request *model.Request) {
	ws.handleWebsocketRequest(request)
}
