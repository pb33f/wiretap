// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package cmd

import (
	"errors"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/ranch/bus"
	"github.com/pb33f/ranch/plank/pkg/server"
	"github.com/pb33f/ranch/plank/utils"
	"github.com/pb33f/wiretap/config"
	"github.com/pb33f/wiretap/controls"
	"github.com/pb33f/wiretap/daemon"
	"github.com/pb33f/wiretap/report"
	"github.com/pb33f/wiretap/shared"
	"github.com/pb33f/wiretap/specs"
	"github.com/pterm/pterm"
	"os"
	"reflect"
	"strconv"
)

func runWiretapService(wiretapConfig *shared.WiretapConfiguration) (server.PlatformServer, error) {

	var doc libopenapi.Document
	var err error
	// load the openapi spec
	if wiretapConfig.Contract != "" {
		doc, err = loadOpenAPISpec(wiretapConfig.Contract)
		if err != nil {
			return nil, err
		}

		// build a model
		_, errs := doc.BuildV3Model()
		if len(errs) > 0 {
			return nil, errors.Join(errs...)
		}
	}

	if doc != nil {
		pterm.Info.Printf("OpenAPI Specification: '%s' parsed and read\n", wiretapConfig.Contract)
	}

	// create a store and put the wiretapConfig in it.
	storeManager := bus.GetBus().GetStoreManager()
	controlsStore := storeManager.CreateStoreWithType(controls.ControlServiceChan, reflect.TypeOf(wiretapConfig))
	controlsStore.Put(shared.ConfigKey, wiretapConfig, nil)

	// create a new ranch config.
	ranchConfig, _ := server.CreateServerConfig()
	ranchConfig.Port, _ = strconv.Atoi(wiretapConfig.WebSocketPort)
	ranchConfig.FabricConfig.EndpointConfig.Heartbeat = 0
	ranchConfig.LogConfig.FormatOptions = &utils.LogFormatOption{
		DisableTimestamp: true,
	}

	// running TLS?
	if wiretapConfig.CertificateKey != "" && wiretapConfig.Certificate != "" {
		tlsConfig := &server.TLSCertConfig{
			CertFile:                  wiretapConfig.Certificate,
			KeyFile:                   wiretapConfig.CertificateKey,
			SkipCertificateValidation: true,
		}
		ranchConfig.TLSCertConfig = tlsConfig
	}

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

	// create an instance of ranch
	platformServer := server.NewPlatformServer(ranchConfig)

	// create wiretap service
	wtService := daemon.NewWiretapService(doc, wiretapConfig)

	// register wiretap service
	if err = platformServer.RegisterService(wtService, daemon.WiretapServiceChan); err != nil {
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

	// create a new chan and listen for interrupt signals
	sysChan := make(chan os.Signal, 1)

	// hook in booted message
	bootedMessage(wiretapConfig)

	// boot the http handler
	handleHttpTraffic(wiretapConfig, wtService)

	// boot the monitor
	serveMonitor(wiretapConfig)

	// if static dir is configured, monitor static content
	if wiretapConfig.StaticDir != "" {
		daemon.MonitorStatic(wiretapConfig)
	}

	// boot wiretap
	platformServer.StartServer(sysChan)
	return platformServer, nil
}
