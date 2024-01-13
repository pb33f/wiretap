// Copyright 2023-2024 Princess Beef Heavy Industries, LLC / Dave Shanley
// https://pb33f.io
//
// SPDX-License-Identifier: AGPL

package har

import (
	"github.com/google/uuid"
	"github.com/pb33f/harhar"
	"github.com/pb33f/ranch/bus"
	"github.com/pb33f/ranch/model"
	"github.com/pb33f/ranch/service"
	"github.com/pb33f/wiretap/daemon"
	"github.com/pb33f/wiretap/shared"
	"log/slog"
	"time"
)

const (
	HARServiceChan     = "har-service"
	StartTheHARRequest = "start-the-har"
)

type HARService struct {
	harStore       bus.BusStore
	logger         *slog.Logger
	wiretapService *daemon.WiretapService
}

type ControlResponse struct {
	Config *shared.WiretapConfiguration `json:"config,omitempty"`
}

func NewHARService(wiretapService *daemon.WiretapService, logger *slog.Logger) *HARService {
	harStore := bus.GetBus().GetStoreManager().CreateStore(HARServiceChan)
	return &HARService{
		harStore:       harStore,
		logger:         logger,
		wiretapService: wiretapService,
	}
}

func (hs *HARService) HandleServiceRequest(request *model.Request, core service.FabricServiceCore) {
	switch request.RequestCommand {
	case StartTheHARRequest:
		hs.startTheHAR(request)
	default:
		core.HandleUnknownRequest(request)
	}
}

func (hs *HARService) startTheHAR(request *model.Request) {
	if hs.harStore != nil {
		har := hs.harStore.GetValue(shared.HARKey)
		if har != nil {
			harFile := har.(*harhar.HAR)
			if harFile != nil {
				for _, entry := range harFile.Log.Entries {
					httpRequest, err := harhar.ConvertRequestIntoHttpRequest(entry.Request)
					if err != nil {
						hs.logger.Error("error converting request", "error", err.Error())
						continue
					}
					id, _ := uuid.NewUUID()
					request.HttpRequest = httpRequest
					request.Id = &id
					hs.wiretapService.ValidateRequest(request, httpRequest)

					time.Sleep(10 * time.Millisecond)

					httpResponse := harhar.ConvertResponseIntoHttpResponse(entry.Response)
					hs.wiretapService.ValidateResponse(request, httpResponse)
				}
			}
		}
	}
}
