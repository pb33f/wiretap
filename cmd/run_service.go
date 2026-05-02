// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package cmd

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strconv"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/ranch/plank/pkg/server"
	ranchService "github.com/pb33f/ranch/service"
	"github.com/pb33f/ranch/transport/fabric"
	"github.com/pb33f/wiretap/config"
	"github.com/pb33f/wiretap/controls"
	"github.com/pb33f/wiretap/daemon"
	"github.com/pb33f/wiretap/har"
	"github.com/pb33f/wiretap/report"
	"github.com/pb33f/wiretap/shared"
	"github.com/pb33f/wiretap/specs"
	staticMock "github.com/pb33f/wiretap/static-mock"
)

func runWiretapService(wiretapConfig *shared.WiretapConfiguration, docs []shared.ApiDocument, primaryDoc libopenapi.Document, conflictReports ...*specs.ConflictReport) (server.PlatformServer, error) {
	// create a new ranch config.
	ranchConfig, err := server.CreateServerConfig()
	if err != nil {
		return nil, fmt.Errorf("create ranch server config: %w", err)
	}
	ranchPort, err := strconv.Atoi(wiretapConfig.WebSocketPort)
	if err != nil {
		return nil, fmt.Errorf("parse websocket port %q: %w", wiretapConfig.WebSocketPort, err)
	}
	ranchConfig.Port = ranchPort
	ranchConfig.Logger = wiretapConfig.Logger

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
		EndpointConfig: &fabric.EndpointConfig{
			Heartbeat:             0,
			UserQueuePrefix:       "/queue",
			TopicPrefix:           "/topic",
			AppRequestPrefix:      "/pub",
			AppRequestQueuePrefix: "/pub/queue",
		},
	}

	// create an instance of ranch
	platformServer := server.NewPlatformServer(ranchConfig)
	storeManager := platformServer.StoreManager()

	// create stores and seed runtime configuration.
	controlsStore := storeManager.CreateStoreWithType(controls.ControlServiceChan, reflect.TypeOf(wiretapConfig))
	controlsStore.Put(shared.ConfigKey, wiretapConfig, nil)

	harStore := storeManager.CreateStoreWithType(har.HARServiceChan, reflect.TypeOf(""))
	harStore.Put(shared.HARKey, wiretapConfig.HAR, nil)

	// create wiretap service
	var conflictReport *specs.ConflictReport
	if len(conflictReports) > 0 {
		conflictReport = conflictReports[0]
	}
	wtService := daemon.NewWiretapService(docs, wiretapConfig, storeManager, conflictReport)

	// register wiretap service
	if err := registerPlatformService(platformServer, "wiretap", daemon.WiretapServiceChan, wtService); err != nil {
		return platformServer, err
	}

	staticMockService := staticMock.NewStaticMockService(wtService, wiretapConfig.Logger)
	// register Static-Mock Service
	if err := registerPlatformService(platformServer, "static mock", staticMock.StaticMockServiceChan, staticMockService); err != nil {
		return platformServer, err
	}

	// register spec service
	if err := registerPlatformService(platformServer, "spec", specs.SpecServiceChan, specs.NewSpecService(primaryDoc)); err != nil {
		return platformServer, err
	}

	// register control service
	if err := registerPlatformService(platformServer, "control", controls.ControlServiceChan, controls.NewControlsService(storeManager)); err != nil {
		return platformServer, err
	}

	// register report service
	if err := registerPlatformService(platformServer, "report", report.ReportServiceChan, report.NewReportService(storeManager)); err != nil {
		return platformServer, err
	}

	// register wiretapConfig service
	if err := registerPlatformService(platformServer, "configuration", config.ConfigurationServiceChan, config.NewConfigurationService(storeManager)); err != nil {
		return platformServer, err
	}

	// register HAR Service
	if err := registerPlatformService(platformServer, "HAR", har.HARServiceChan, har.NewHARService(wtService, wiretapConfig.Logger, wiretapConfig.HARReplayDelay, storeManager)); err != nil {
		return platformServer, err
	}

	// Start watcher to look for changes to static mock definitions.
	staticMockService.StartWatcher()

	// create a new chan and listen for interrupt signals
	sysChan := make(chan os.Signal, 1)

	// hook in booted message
	bootedMessage(wiretapConfig, platformServer.Bus())

	// boot the http handler
	hht := HandleHttpTraffic{
		WiretapConfig:     wiretapConfig,
		WiretapService:    wtService,
		StaticMockService: staticMockService,
	}
	handleHttpTraffic(&hht)

	// boot the monitor
	serveMonitor(wiretapConfig)

	// if static dir is configured, monitor static content
	if wiretapConfig.StaticDir != "" {
		daemon.MonitorStatic(wiretapConfig, platformServer.Bus())
	}

	// boot wiretap
	if err := platformServer.StartServer(context.Background(), sysChan); err != nil {
		return platformServer, err
	}
	return platformServer, nil
}

func registerPlatformService(
	platformServer server.PlatformServer,
	name string,
	channel string,
	svc ranchService.FabricService,
) error {
	if err := platformServer.RegisterService(svc, channel); err != nil {
		return fmt.Errorf("register %s service: %w", name, err)
	}
	return nil
}
