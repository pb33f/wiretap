// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package cmd

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/ranch/bus"
	"github.com/pb33f/ranch/model"
	"github.com/pb33f/ranch/plank/pkg/server"
	"github.com/pb33f/ranch/service"
	"github.com/pb33f/wiretap/controls"
	"github.com/pb33f/wiretap/daemon"
	"github.com/pb33f/wiretap/shared"
	"github.com/pb33f/wiretap/specs"
	"github.com/pterm/pterm"
	"github.com/sirupsen/logrus"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
)

func loadOpenAPISpec(contract string) (libopenapi.Document, error) {
	var specBytes []byte

	if strings.HasPrefix(contract, "http://") || strings.HasPrefix(contract, "https://") {
		if docUrl, err := url.Parse(contract); err == nil {
			logrus.Infof("Fetching OpenAPI Specification from URL: %s", docUrl.String())
			resp, er := http.Get(docUrl.String())
			if er != nil {
				return nil, er
			}
			respBody, e := io.ReadAll(resp.Body)
			if e != nil {
				return nil, e
			}
			if len(respBody) > 0 {
				specBytes = respBody
			}
		}
	} else {

		// not an URL, is it a file?
		var er error
		if _, er = os.Stat(contract); er != nil {
			return nil, er
		}
		specBytes, er = os.ReadFile(contract)
		if er != nil {
			return nil, er
		}
	}
	if len(specBytes) <= 0 {
		return nil, fmt.Errorf("no bytes in OpenAPI Specification")
	}
	return libopenapi.NewDocument(specBytes)
}

func runWiretapService(config *shared.WiretapConfiguration) (server.PlatformServer, error) {

	doc, err := loadOpenAPISpec(config.Contract)
	if err != nil {
		return nil, err
	}
	_, errs := doc.BuildV3Model()
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	serverConfig, _ := server.CreateServerConfig()
	serverConfig.Port, _ = strconv.Atoi(config.Port)
	serverConfig.FabricConfig.EndpointConfig.Heartbeat = 0

	serverConfig.FabricConfig = &server.FabricBrokerConfig{
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

	/* serve monitor */
	go func() {
		fs := http.FileServer(http.Dir("ui/dist/"))
		http.Handle("/", fs)

		log.Print("Monitor UI booting on 9091...")
		err = http.ListenAndServe(":9091", nil)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// create a store and put the config in it.
	ebus := bus.GetBus()
	storeManager := ebus.GetStoreManager()
	controlsStore := storeManager.CreateStoreWithType(controls.ControlServiceChan, reflect.TypeOf(config))
	controlsStore.Put(shared.ConfigKey, config, nil)

	// create an instance of ranch
	platformServer := server.NewPlatformServer(serverConfig)

	// register wiretap service
	if err = platformServer.RegisterService(daemon.NewWiretapService(doc), daemon.WiretapServiceChan); err != nil {
		panic(err)
	}

	// register spec service
	if err = platformServer.RegisterService(specs.NewSpecService(doc), specs.SpecServiceChan); err != nil {
		panic(err)
	}

	// register control service
	if err = platformServer.RegisterService(controls.NewControlsService(), controls.ControlServiceChan); err != nil {
		panic(err)
	}

	// create a new catchall endpoint and listen for all traffic
	platformServer.SetHttpPathPrefixChannelBridge(rbc)

	// start the ranch.
	sysChan := make(chan os.Signal, 1)

	go func() {
		handler, _ := bus.GetBus().ListenStream(server.RANCH_SERVER_ONLINE_CHANNEL)
		seen := false
		handler.Handle(func(message *model.Message) {
			if !seen {
				seen = true
				pterm.Println()
				pterm.Info.Println("Wiretap Service is ready.")
				pterm.Println()
				pterm.Info.Printf("API Gateway: http://localhost:%s\n", config.Port)
				pterm.Info.Printf("Monitor: http://localhost:%s\n", config.MonitorPort)
				pterm.Println()

			}
		}, nil)
	}()

	platformServer.StartServer(sysChan)
	return platformServer, nil
}
