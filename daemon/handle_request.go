// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package daemon

import (
    "github.com/pb33f/libopenapi-validator/parameters"
    "github.com/pb33f/libopenapi-validator/requests"
    "github.com/pb33f/libopenapi-validator/responses"
    "github.com/pb33f/ranch/model"
    "github.com/pb33f/ranch/service"
    "io"
    "net/http"
)

func (ws *WiretapService) handleHttpRequest(request *model.Request, core service.FabricServiceCore) {

    lowResponseChan := make(chan *http.Response)
    lowErrorChan := make(chan error)
    var returnedResponse *http.Response
    var returnedError error

    // create validators.
    requestValidator := requests.NewRequestBodyValidator(ws.docModel)
    paramValidator := parameters.NewParameterValidator(ws.docModel)
    responseValidator := responses.NewResponseBodyValidator(ws.docModel)

    // validate the request
    go ws.validateRequest(request, requestValidator, paramValidator, responseValidator)

    // call the API being requested.
    go ws.callAPI(request.HttpRequest, lowResponseChan, lowErrorChan)

doneWaitingForResponse:
    for {
        select {
        case resp, ok := <-lowResponseChan:
            if ok {
                returnedResponse = resp
            }
            break doneWaitingForResponse
        case err := <-lowErrorChan:
            returnedError = err
            break doneWaitingForResponse
        }
    }

    if returnedResponse == nil && returnedError != nil {
        go ws.broadcastResponseError(request, returnedResponse, returnedError)
        core.SendErrorResponse(request, 500, returnedError.Error())
        return
    } else {
        // validate response
        go ws.validateResponse(request, responseValidator, returnedResponse)
    }

    body, _ := io.ReadAll(returnedResponse.Body)
    headers := extractHeaders(returnedResponse)

    // todo: deal with headers and correct raw values
    if returnedResponse.StatusCode >= 400 {
        core.SendErrorResponseAsStringWithHeadersAndPayload(request, returnedResponse.StatusCode,
            "HTTP Request failed", string(body), headers)
    } else {
        core.SendResponseAsStringWithHeaders(request, string(body), headers)
    }
}
