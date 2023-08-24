// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package controls

import (
	"github.com/mitchellh/mapstructure"
	"github.com/pb33f/ranch/bus"
	"github.com/pb33f/ranch/model"
	"github.com/pb33f/ranch/service"
	"github.com/pb33f/wiretap/shared"
)

const (
	ControlServiceChan = "controls"
	ChangeDelayRequest = "change-delay-request"
)

type ControlService struct {
	controlsStore bus.BusStore
}

type ChangeGlobalDelayRequest struct {
	Delay int `json:"delay,omitempty"`
}

type ControlResponse struct {
	Config *shared.WiretapConfiguration `json:"config,omitempty"`
}

func NewControlsService() *ControlService {
	storeManager := bus.GetBus().GetStoreManager()
	controlsStore := storeManager.CreateStore(ControlServiceChan)
	return &ControlService{
		controlsStore: controlsStore,
	}
}

func (cs *ControlService) HandleServiceRequest(request *model.Request, core service.FabricServiceCore) {
	switch request.RequestCommand {
	case ChangeDelayRequest:
		cs.changeDelay(request, core)
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
		core.SendResponse(request, &ControlResponse{config})

	} else {
		core.SendErrorResponse(request, 400, "Invalid delay value")
	}
}
