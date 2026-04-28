// Copyright 2023-2024 Princess Beef Heavy Industries, LLC / Dave Shanley
// https://pb33f.io
//
// SPDX-License-Identifier: AGPL

package har

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pb33f/harific/motor"
	harModel "github.com/pb33f/harific/motor/model"
	"github.com/pb33f/ranch/bus"
	"github.com/pb33f/ranch/model"
	"github.com/pb33f/ranch/service"
	"github.com/pb33f/wiretap/daemon"
	"github.com/pb33f/wiretap/shared"
	"log/slog"
)

const (
	HARServiceChan     = "har-service"
	StartTheHARRequest = "start-the-har"
)

type HARService struct {
	harStore       bus.BusStore
	logger         *slog.Logger
	wiretapService *daemon.WiretapService
	replayDelay    time.Duration
}

type ControlResponse struct {
	Config *shared.WiretapConfiguration `json:"config,omitempty"`
}

func NewHARService(wiretapService *daemon.WiretapService, logger *slog.Logger, replayDelayMillis int) *HARService {
	harStore := bus.GetBus().GetStoreManager().CreateStore(HARServiceChan)
	return &HARService{
		harStore:       harStore,
		logger:         logger,
		wiretapService: wiretapService,
		replayDelay:    time.Duration(replayDelayMillis) * time.Millisecond,
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
	if hs.harStore == nil {
		return
	}

	harValue := hs.harStore.GetValue(shared.HARKey)
	harPath, ok := harValue.(string)
	if !ok || harPath == "" {
		return
	}

	streamer, err := NewHARStreamer(harPath, motor.StreamerOptions{WorkerCount: 1})
	if err != nil {
		hs.logger.Error("error creating HAR streamer", "error", err.Error())
		return
	}
	defer streamer.Close()

	ctx := context.Background()
	if err = streamer.Initialize(ctx); err != nil {
		hs.logger.Error("error initializing HAR streamer", "error", err.Error())
		return
	}

	index := streamer.GetIndex()
	if index == nil || index.TotalEntries == 0 {
		return
	}

	results, err := streamer.StreamRange(ctx, 0, index.TotalEntries)
	if err != nil {
		hs.logger.Error("error streaming HAR", "error", err.Error())
		return
	}

	for result := range results {
		if result.Error != nil {
			hs.logger.Error("error streaming HAR entry", "error", result.Error.Error())
			continue
		}
		if result.Entry == nil {
			continue
		}

		httpRequest, err := harModel.ConvertRequestIntoHttpRequest(result.Entry.Request)
		if err != nil {
			hs.logger.Error("error converting request", "error", err.Error())
			continue
		}

		id, _ := uuid.NewUUID()
		request.HttpRequest = httpRequest
		request.Id = &id
		hs.wiretapService.ValidateRequest(request, httpRequest)

		if hs.replayDelay > 0 {
			time.Sleep(hs.replayDelay)
		}

		httpResponse := harModel.ConvertResponseIntoHttpResponse(result.Entry.Response)
		hs.wiretapService.ValidateResponse(request, httpResponse)
	}
}
