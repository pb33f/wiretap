// Copyright 2023-2024 Princess Beef Heavy Industries, LLC / Dave Shanley
// https://pb33f.io
//
// SPDX-License-Identifier: AGPL

package staticMock

import (
	"encoding/json"
	"log/slog"
	"os"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/pb33f/ranch/model"
	"github.com/pb33f/ranch/service"
	"github.com/pb33f/wiretap/daemon"
)

const (
	StaticMockServiceChan = "static-mock-service"
	IncomingHttpRequest   = "incoming-http-request"
	MockDefinitionsPath   = "/mock-definitions"
	MockBodyJsonsPath     = "/body-jsons/"
)

type StaticMockDefinitionRequest struct {
	Method      string          `json:"method,omitempty"`
	UrlPath     string          `json:"urlPath,omitempty"`
	Host        string          `json:"host,omitempty"`
	Header      *map[string]any `json:"header,omitempty"`
	Body        interface{}     `json:"body,omitempty"`
	QueryParams *map[string]any `json:"queryParams,omitempty"`
}

type StaticMockDefinitionResponse struct {
	Header           map[string]any `json:"header,omitempty"`
	StatusCode       int            `json:"statusCode,omitempty"`
	Body             string         `json:"body,omitempty"`
	BodyJsonFilename string         `json:"bodyJsonFilename,omitempty"`
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

// getDefinitionFromJson converts a JSON object to a StaticMockDefinition
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

// loadStaticMockRequestsAndResponses loads the static mock definitions from the JSON files
func loadStaticMockRequestsAndResponses(wiretapService *daemon.WiretapService, logger *slog.Logger) []StaticMockDefinition {
	var staticMockDefinitions []StaticMockDefinition
	mocksPath := wiretapService.StaticMockDir + MockDefinitionsPath

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
				logger.Error("Error parsing json file %s: %v\n", filePath, err)
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
				logger.Error("JSON not in the right format. \nFile => %s\n JSON => \n%s", file.Name(), string(data))
			}
		}
	}

	return staticMockDefinitions
}

// StartWatcher Function to start a watcher on mock-definitions folder
func (sms *StaticMockService) StartWatcher() {
	pathToWatch := sms.wiretapService.StaticMockDir + MockDefinitionsPath

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		sms.logger.Error("Error when creating fsnotify.NewWatcher '%s'", err.Error(), err)
	}

	err = watcher.Add(pathToWatch)
	if err != nil {
		sms.logger.Error("Error adding path to watch. path => '%s'", pathToWatch, err)
	}

	go func(sms *StaticMockService) {
		// Event loop
		for {
			select {
			case event, ok := <-watcher.Events:
				// If files are created, deleted or modified, call the handler
				if !ok {
					return
				}
				eventsToWatch := event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename)
				fileTypeIsJson := strings.HasSuffix(strings.ToLower(event.Name), ".json")
				if eventsToWatch && fileTypeIsJson {
					sms.handleStaticMockChange()
				}
			case err := <-watcher.Errors:
				if err != nil {
					sms.logger.Error("Error while watching files => '%s'", err.Error(), err)
				}
				watcher.Close()
			}
		}
	}(sms)
}

// handleStaticMockChange Function to handle changes to mock definitions. It loads the mock definitions in realtime
// so that the entire wiretap service doesn't need a restart
func (sms *StaticMockService) handleStaticMockChange() {
	sms.logger.Info("Mock definitions modified. Rebuilding mocks...")
	sms.mockDefinitions = loadStaticMockRequestsAndResponses(sms.wiretapService, sms.logger)
	sms.logger.Info("New mock definitions loaded")
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
