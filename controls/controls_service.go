// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package controls

import (
	"github.com/go-viper/mapstructure/v2"
	"github.com/pb33f/ranch/model"
	"github.com/pb33f/ranch/service"
	"github.com/pb33f/ranch/store"
	"github.com/pb33f/wiretap/shared"
)

const (
	ControlServiceChan = "controls"
	ChangeDelayRequest = "change-delay-request"
	ResetStateRequest  = "reset-state-request"
)

type ControlService struct {
	controlsStore    store.BusStore
	transactionStore store.BusStore
	harStore         store.BusStore
}

type ChangeGlobalDelayRequest struct {
	Delay int `json:"delay,omitempty"`
}

type ControlResponse struct {
	Config *shared.WiretapConfiguration `json:"config,omitempty"`
	Reset  bool                         `json:"reset,omitempty"`
}

func NewControlsService(storeManager store.Manager) *ControlService {
	controlsStore := storeManager.CreateStore(ControlServiceChan)
	transactionStore := storeManager.GetStore(shared.WiretapServiceChan)
	harStore := storeManager.GetStore(shared.HARServiceChan)
	return &ControlService{
		controlsStore:    controlsStore,
		transactionStore: transactionStore,
		harStore:         harStore,
	}
}

func (cs *ControlService) HandleServiceRequest(request *model.Request, core service.FabricServiceCore) {
	switch request.RequestCommand {
	case ChangeDelayRequest:
		cs.changeDelay(request, core)
	case ResetStateRequest:
		cs.resetState(request, core)
	default:
		core.HandleUnknownRequest(request)
	}
}

func (cs *ControlService) changeDelay(request *model.Request, core service.FabricServiceCore) {

	if dl, ok := request.Payload.(map[string]interface{}); ok {

		// decode the object into a request
		var r ChangeGlobalDelayRequest
		_ = mapstructure.Decode(dl, &r)

		// extract state from store.
		controls := cs.controlsStore.GetValue(shared.ConfigKey)
		config := controls.(*shared.WiretapConfiguration)

		// update if valid.
		if r.Delay >= 0 {
			config.GlobalAPIDelay = r.Delay
			cs.controlsStore.Put(shared.ConfigKey, config, nil)
		}
		core.SendResponse(request, &ControlResponse{Config: config})

	} else {
		core.SendErrorResponse(request, 400, "Invalid delay value")
	}
}

func (cs *ControlService) resetState(request *model.Request, core service.FabricServiceCore) {
	config := cs.resetRuntimeState()
	core.SendResponse(request, &ControlResponse{
		Config: config,
		Reset:  true,
	})
}

func (cs *ControlService) resetRuntimeState() *shared.WiretapConfiguration {
	if cs.transactionStore != nil {
		cs.transactionStore.Reset()
		cs.transactionStore.Initialize()
	}
	if cs.harStore != nil {
		cs.harStore.Reset()
		cs.harStore.Initialize()
	}

	controls := cs.controlsStore.GetValue(shared.ConfigKey)
	config, ok := controls.(*shared.WiretapConfiguration)
	if !ok || config == nil {
		return nil
	}
	config.GlobalAPIDelay = 0
	cs.controlsStore.Put(shared.ConfigKey, config, nil)
	return config
}
