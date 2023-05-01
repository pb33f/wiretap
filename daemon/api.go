// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package daemon

import "net/http"

func (ws *WiretapService) callAPI(req *http.Request, responseChan chan *http.Response, errorChan chan error) {
    // create a new request from the original request, but replace the path
    resp, err := ws.client.Do(cloneRequest(req))
    if err != nil {
        errorChan <- err
        close(errorChan)
    }
    responseChan <- resp
    close(responseChan)
}
