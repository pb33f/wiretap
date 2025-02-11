// Copyright 2023-2024 Princess Beef Heavy Industries, LLC / Dave Shanley
// https://pb33f.io
//
// SPDX-License-Identifier: AGPL

package staticMock

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/pb33f/ranch/model"
	"github.com/pb33f/ranch/service"
	"github.com/pb33f/wiretap/daemon"
)

const (
	StaticMockServiceChan = "static-mock-service"
	IncomingHttpRequest   = "incoming-http-request"
)

type StaticMockDefinitionRequest struct {
	Method      string       `json:"method,omitempty"`
	UrlPath     string       `json:"urlPath,omitempty"`
	URL         *url.URL     `json:"url,omitempty"`
	Header      *http.Header `json:"header,omitempty"`
	Body        string       `json:"body,omitempty"`
	QueryParams url.Values   `json:"queryParams,omitempty"`
}

type StaticMockDefinitionResponse struct {
	Timestamp    int64                         `json:"timestamp,omitempty"`
	Headers      map[string]any                `json:"headers,omitempty"`
	StatusCode   int                           `json:"statusCode,omitempty"`
	Body         string                        `json:"responseBody,omitempty"`
	Cookies      map[string]*daemon.HttpCookie `json:"cookies,omitempty"`
	Time         time.Time                     `json:"-"`
	BodyJsonPath string                        `json:"bodyJsonPath,omitempty"`
}

type StaticMockDefinition struct {
	Request  StaticMockDefinitionRequest  `json:"request,omitempty"`
	Response StaticMockDefinitionResponse `json:"response,omitempty"`
}

type StaticMockService struct {
	logger          *slog.Logger
	wiretapService  *daemon.WiretapService
	mockDefinitions []StaticMockDefinition
}

func NewStaticMockService(wiretapService *daemon.WiretapService, logger *slog.Logger) *StaticMockService {
	mockDefinitions := loadStaticMockRequestsAndResponses(wiretapService, logger)

	return &StaticMockService{
		logger:          logger,
		wiretapService:  wiretapService,
		mockDefinitions: mockDefinitions,
	}
}

func getDefinitionFromJson(mockInterface map[string]interface{}) (StaticMockDefinition, error) {
	var mockDefinition StaticMockDefinition

	mockInterfaceJson, err := json.Marshal(mockInterface)
	if err != nil {
		return StaticMockDefinition{}, err
	}

	err = json.Unmarshal(mockInterfaceJson, &mockDefinition)
	if err != nil {
		return StaticMockDefinition{}, err
	}

	return mockDefinition, nil
}

func loadStaticMockRequestsAndResponses(wiretapService *daemon.WiretapService, logger *slog.Logger) []StaticMockDefinition {
	var staticMockDefinitions []StaticMockDefinition
	mocksPath := wiretapService.StaticMockDir + "/mock-definitions"

	files, err := os.ReadDir(mocksPath)
	if err != nil {
		logger.Error(err.Error())
	}

	// Loop through & read each mock definition file
	for _, file := range files {
		// Check if it's a regular file (not a directory)
		if !file.IsDir() {
			filePath := mocksPath + "/" + file.Name()
			data, err := os.ReadFile(filePath)
			if err != nil {
				logger.Error("Error reading file %s: %v\n", filePath, err)
				continue
			}

			var mockDefinitions interface{}

			err = json.Unmarshal(data, &mockDefinitions)
			if err != nil {
				logger.Error(err.Error())
				continue
			}

			switch mdJson := mockDefinitions.(type) {
			// If the content of the file is a JSON object (key-value pairs)
			case map[string]interface{}:
				mockDefinition, err := getDefinitionFromJson(mdJson)
				if err != nil {
					logger.Error(err.Error())
					continue
				}
				staticMockDefinitions = append(staticMockDefinitions, mockDefinition)

			// If the content of the file is a JSON array (array of requests)
			case []interface{}:
				// You can iterate over the array
				for _, item := range mdJson {
					mockDefinition, err := getDefinitionFromJson(item.(map[string]interface{}))
					if err != nil {
						logger.Error(err.Error())
						continue
					}
					staticMockDefinitions = append(staticMockDefinitions, mockDefinition)
				}

			default:
				// If it's neither an object nor an array
				logger.Error("JSON not in the right format.")
			}
		}
	}

	return staticMockDefinitions
}

func (sms *StaticMockService) HandleServiceRequest(request *model.Request, core service.FabricServiceCore) {
	switch request.RequestCommand {
	case IncomingHttpRequest:
		sms.HandleStaticMockRequest(request)
	default:
		core.HandleUnknownRequest(request)
	}
}

func (sms *StaticMockService) HandleStaticMockRequest(request *model.Request) {
	sms.handleStaticMockRequest(request)
}
