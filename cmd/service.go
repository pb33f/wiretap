// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package cmd

import (
    "errors"
    "fmt"
    "github.com/google/uuid"
    "github.com/pb33f/libopenapi"
    "github.com/pb33f/ranch/model"
    "github.com/pb33f/ranch/plank/pkg/server"
    "github.com/pb33f/ranch/service"
    "github.com/pb33f/wiretap/daemon"
    "github.com/sirupsen/logrus"
    "io"
    "net/http"
    "net/url"
    "os"
    "strconv"
)

type WiretapServiceConfiguration struct {
    Contract     string
    RedirectHost string
    Port         string
}

func loadOpenAPISpec(contract string) (libopenapi.Document, error) {
    var specBytes []byte
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
    } else {

        // not an URL, is it a file?
        if _, er := os.Stat(contract); er == nil {
            return nil, er
        }
        specBytes, err = os.ReadFile(contract)
        if err != nil {
            return nil, err
        }
    }
    if len(specBytes) <= 0 {
        return nil, fmt.Errorf("no bytes in OpenAPI Specification")
    }
    return libopenapi.NewDocument(specBytes)
}

func runWiretapService(config *WiretapServiceConfiguration) (server.PlatformServer, error) {

    doc, err := loadOpenAPISpec(config.Contract)
    if err != nil {
        return nil, err
    }
    _, errs := doc.BuildV3Model()
    if len(errs) > 0 {
        return nil, errors.Join(errs...)
    }

    // configure daemon.
    serverConfig, _ := server.CreateServerConfig()
    serverConfig.Port, _ = strconv.Atoi(config.Port)
    serverConfig.SpaConfig = &server.SpaConfig{
        RootFolder: "docs",
        BaseUri:    "/",
    }
    serverConfig.FabricConfig.EndpointConfig.Heartbeat = 0
    serverConfig.StaticDir = []string{"/static"}

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

    // create an instance of plank.
    platformServer := server.NewPlatformServer(serverConfig)
    //platformServer.SetStaticRoute("change-report", "change-report")

    // boot what-changed html report service.
    if err = platformServer.RegisterService(daemon.NewWiretapService(doc),
        daemon.WiretapServiceChan); err != nil {
        panic(err)
    }

    // create a new catchall endpoint and listen for all traffic
    platformServer.SetHttpPathPrefixChannelBridge(rbc)

    // start the ranch.
    sysChan := make(chan os.Signal, 1)
    platformServer.StartServer(sysChan)
    return platformServer, nil
}
