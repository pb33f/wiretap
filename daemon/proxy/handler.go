// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/pb33f/ranch/model"
	configModel "github.com/pb33f/wiretap/config"
	"github.com/pb33f/wiretap/daemon/problems"
	"github.com/pb33f/wiretap/shared"
)

type APICaller func(*http.Request, ...*shared.WiretapConfiguration) (*http.Response, error)
type RequestValidator func() []*shared.WiretapValidationError
type ResponseValidator func(*http.Response, []byte) []*shared.WiretapValidationError
type ResponseErrorBroadcaster func(*http.Response, error)

type PreparedRequest struct {
	Config                 *shared.WiretapConfiguration
	NewReq                 *http.Request
	APIRequest             *http.Request
	BodyBytes              []byte
	ControlPath            string
	IsHardError            bool
	CallAPI                APICaller
	ValidateRequest        RequestValidator
	ValidateResponse       ResponseValidator
	BroadcastResponseError ResponseErrorBroadcaster
}

type Handler struct {
	transport     http.RoundTripper
	validationSem chan struct{}
}

const defaultValidationConcurrency = 8

func NewHandler(transport ...http.RoundTripper) *Handler {
	tr := http.DefaultTransport
	if len(transport) > 0 && transport[0] != nil {
		tr = transport[0]
	}
	return &Handler{
		transport:     tr,
		validationSem: make(chan struct{}, defaultValidationConcurrency),
	}
}

func (h *Handler) Handle(request *model.Request, prep *PreparedRequest) {
	config := prep.Config
	var requestErrors []*shared.WiretapValidationError
	var responseErrors []*shared.WiretapValidationError
	controlPath := prepControlPath(prep)

	if configModel.IgnoreValidationOnPath(controlPath, config) &&
		!configModel.PathValidationAllowListed(controlPath, config) {
		config.Logger.Info(
			fmt.Sprintf("Request on validation ignored path: %s ; skipping validation", controlPath))
	} else if prep.IsHardError {
		requestErrors = prep.ValidateRequest()
	} else {
		h.runValidationBounded(func() {
			prep.ValidateRequest()
		})
	}

	callAPI := prep.CallAPI
	if callAPI == nil {
		callAPI = h.callAPI
	}
	returnedResponse, returnedError := callAPI(prep.APIRequest, config)

	if returnedResponse == nil && returnedError != nil {
		config.Logger.Info("[wiretap] request failed", "url", prep.APIRequest.URL.String(), "code", 500,
			"error", returnedError.Error())
		go prep.BroadcastResponseError(returnedResponse, returnedError)
		request.HttpResponseWriter.WriteHeader(http.StatusInternalServerError)
		wtError := shared.GenerateError("Unable to call API", http.StatusInternalServerError, returnedError.Error(), "", returnedResponse)
		_, _ = request.HttpResponseWriter.Write(shared.MarshalError(wtError))
		return
	}

	respBody, _ := io.ReadAll(returnedResponse.Body)
	_ = returnedResponse.Body.Close()
	returnedResponse.Body = io.NopCloser(bytes.NewBuffer(respBody))

	if prep.IsHardError {
		responseErrors = prep.ValidateResponse(returnedResponse, respBody)
	} else {
		// Clone headers for async validation; http.Header is a map and the main
		// goroutine continues to read and rewrite returnedResponse.Header below.
		clonedResp := &http.Response{
			StatusCode: returnedResponse.StatusCode,
			Header:     returnedResponse.Header.Clone(),
			Body:       io.NopCloser(bytes.NewBuffer(respBody)),
		}
		h.runValidationBounded(func() {
			prep.ValidateResponse(clonedResp, respBody)
		})
	}

	delay := configModel.FindPathDelay(request.HttpRequest.URL.Path, config)
	if delay > 0 {
		time.Sleep(time.Duration(delay) * time.Millisecond)
	} else if config.GlobalAPIDelay > 0 {
		time.Sleep(time.Duration(config.GlobalAPIDelay) * time.Millisecond)
	}

	headers := extractHeaders(returnedResponse)
	shared.SetCORSHeaders(headers)

	if config.StrictRedirectLocation && is3xxStatusCode(returnedResponse.StatusCode) {
		setStrictLocationHeader(config, headers)
	}

	for k, v := range headers {
		for _, j := range v {
			responseHeaders := request.HttpResponseWriter.Header()
			if responseHeaders.Get(k) == "" {
				responseHeaders.Set(k, j)
			} else {
				responseHeaders.Add(k, j)
			}
		}
	}
	config.Logger.Info("[wiretap] request completed", "url", request.HttpRequest.URL.String(), "code", returnedResponse.StatusCode)

	statusCode := problems.PickHardErrorStatus(prep.IsHardError, requestErrors, responseErrors, config, returnedResponse.StatusCode)

	if prep.IsHardError && problems.ShouldReturnValidationProblem(config, requestErrors, responseErrors) {
		problems.WriteValidationProblemResponse(
			request.HttpResponseWriter,
			statusCode,
			request.HttpRequest.URL.Path,
			requestErrors,
			responseErrors,
		)
		return
	}

	request.HttpResponseWriter.WriteHeader(statusCode)
	_, _ = request.HttpResponseWriter.Write(respBody)
}

func prepControlPath(prep *PreparedRequest) string {
	if prep.ControlPath != "" {
		return prep.ControlPath
	}
	if prep.APIRequest != nil && prep.APIRequest.URL != nil {
		return prep.APIRequest.URL.Path
	}
	return ""
}

func (h *Handler) runValidationBounded(work func()) {
	if work == nil {
		return
	}
	select {
	case h.validationSem <- struct{}{}:
		go func() {
			defer func() { <-h.validationSem }()
			work()
		}()
	default:
		work()
	}
}

func extractHeaders(resp *http.Response) map[string][]string {
	headers := make(map[string][]string)
	for k, v := range resp.Header {
		headers[k] = v
	}
	return headers
}

func setStrictLocationHeader(config *shared.WiretapConfiguration, headers map[string][]string) {
	if locations, ok := headers["Location"]; ok {
		newLocations := make([]string, 0)
		apiGatewayHost := config.GetApiGatewayHost()

		for _, location := range locations {
			parsedLocation, parseErr := url.Parse(location)
			if parseErr != nil {
				config.Logger.Warn(fmt.Sprintf("Unable to parse `Location` header URL: %s", location))
				newLocations = append(newLocations, location)
			} else if parsedLocation.Host != "" && parsedLocation.Host != apiGatewayHost {
				parsedLocation.Host = apiGatewayHost
				newLocation := parsedLocation.String()
				config.Logger.Info(fmt.Sprintf("Rewrote `Location` header from %s to %s", location, newLocation))
				newLocations = append(newLocations, newLocation)
			} else {
				newLocations = append(newLocations, location)
			}
		}
		headers["Location"] = newLocations
	}
}

func is3xxStatusCode(statusCode int) bool {
	return 300 <= statusCode && statusCode < 400
}
