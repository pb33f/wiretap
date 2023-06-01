// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package config

import (
	"github.com/mitchellh/mapstructure"
	"github.com/pb33f/ranch/bus"
	"github.com/pb33f/ranch/model"
	"github.com/pb33f/ranch/service"
	"github.com/pb33f/wiretap/controls"
	"github.com/pb33f/wiretap/shared"
)

const (
	ConfigurationServiceChan = "configuration"
	GetConfigurationRequest  = "config-request"
)

type ConfigurationService struct {
	configStore bus.BusStore
}

type RequestConfiguration struct {
}

type Configuration struct {
	Configuration *shared.WiretapConfiguration `json:"configuration,omitempty"`
}

func NewConfigurationService() *ConfigurationService {
	storeManager := bus.GetBus().GetStoreManager()
	controlsStore := storeManager.CreateStore(controls.ControlServiceChan)
	return &ConfigurationService{
		configStore: controlsStore,
	}
}

func (cs *ConfigurationService) HandleServiceRequest(request *model.Request, core service.FabricServiceCore) {
	switch request.RequestCommand {
	case GetConfigurationRequest:
		cs.returnConfig(request, core)
	default:
		core.HandleUnknownRequest(request)
	}
}

func (cs *ConfigurationService) returnConfig(request *model.Request, core service.FabricServiceCore) {

	if dl, ok := request.Payload.(map[string]interface{}); ok {

		// decode the object into a request
		var r RequestConfiguration
		_ = mapstructure.Decode(dl, &r)

		// extract state from store.
		storeData, _ := cs.configStore.Get(shared.ConfigKey)
		core.SendResponse(request, storeData)

	} else {
		core.SendErrorResponse(request, 400, "Invalid config request")
	}
}
