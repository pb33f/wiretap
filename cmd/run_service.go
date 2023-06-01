// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package cmd

import (
	"errors"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/pb33f/ranch/bus"
	"github.com/pb33f/ranch/model"
	"github.com/pb33f/ranch/plank/pkg/server"
	"github.com/pb33f/ranch/service"
	"github.com/pb33f/wiretap/config"
	"github.com/pb33f/wiretap/controls"
	"github.com/pb33f/wiretap/daemon"
	"github.com/pb33f/wiretap/report"
	"github.com/pb33f/wiretap/shared"
	"github.com/pb33f/wiretap/specs"
	"net/http"
	"os"
	"reflect"
	"strconv"
)

func runWiretapService(wiretapConfig *shared.WiretapConfiguration) (server.PlatformServer, error) {

	// load the openapi spec
	doc, err := loadOpenAPISpec(wiretapConfig.Contract)
	if err != nil {
		return nil, err
	}

	// build a model
	_, errs := doc.BuildV3Model()
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	// create a store and put the wiretapConfig in it.
	storeManager := bus.GetBus().GetStoreManager()
	controlsStore := storeManager.CreateStoreWithType(controls.ControlServiceChan, reflect.TypeOf(wiretapConfig))
	controlsStore.Put(shared.ConfigKey, wiretapConfig, nil)

	// create a new ranch config.
	ranchConfig, _ := server.CreateServerConfig()
	ranchConfig.Port, _ = strconv.Atoi(wiretapConfig.Port)
	ranchConfig.FabricConfig.EndpointConfig.Heartbeat = 0

	// create an application fabric configuration for the ranch server.
	ranchConfig.FabricConfig = &server.FabricBrokerConfig{
		FabricEndpoint: "/ranch",
		EndpointConfig: &bus.EndpointConfig{
			Heartbeat:             0,
			UserQueuePrefix:       "/queue",
			TopicPrefix:           "/topic",
			AppRequestPrefix:      "/pub",
			AppRequestQueuePrefix: "/pub/queue",
		},
	}

	// create a REST bridge configuration and set it for a prefix for all our requests off root.
	rbc := &service.RESTBridgeConfig{
		ServiceChannel: daemon.WiretapServiceChan,
		Uri:            "/",
		FabricRequestBuilder: func(w http.ResponseWriter, r *http.Request) model.Request {
			id := uuid.New()
			return model.Request{
				Id:                 &id,
				RequestCommand:     daemon.IncomingHttpRequest,
				HttpRequest:        r,
				HttpResponseWriter: w,
			}
		},
	}

	// create an instance of ranch
	platformServer := server.NewPlatformServer(ranchConfig)

	// register wiretap service
	if err = platformServer.RegisterService(
		daemon.NewWiretapService(doc), daemon.WiretapServiceChan); err != nil {
		panic(err)
	}

	// register spec service
	if err = platformServer.RegisterService(
		specs.NewSpecService(doc), specs.SpecServiceChan); err != nil {
		panic(err)
	}

	// register control service
	if err = platformServer.RegisterService(
		controls.NewControlsService(), controls.ControlServiceChan); err != nil {
		panic(err)
	}

	// register report service
	if err = platformServer.RegisterService(
		report.NewReportService(), report.ReportServiceChan); err != nil {
		panic(err)
	}

	// register wiretapConfig service
	if err = platformServer.RegisterService(
		config.NewConfigurationService(), config.ConfigurationServiceChan); err != nil {
		panic(err)
	}

	// create a new catchall endpoint and listen for all traffic
	platformServer.SetHttpPathPrefixChannelBridge(rbc)

	// add global CORS middleware
	middlewareManager := platformServer.GetMiddlewareManager()
	_ = middlewareManager.SetGlobalMiddleware([]mux.MiddlewareFunc{daemon.CORSMiddleware()})

	// create a new chan and listen for interrupt signals
	sysChan := make(chan os.Signal, 1)

	// hook in booted message
	bootedMessage(wiretapConfig)

	// boot the monitor
	serveMonitor(wiretapConfig)

	// boot wiretap
	platformServer.StartServer(sysChan)
	return platformServer, nil
}
