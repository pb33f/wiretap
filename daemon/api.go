// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package daemon

import (
    "fmt"
    "github.com/pb33f/wiretap/shared"
    "net/http"
)

type wiretapTransport struct {
    capturedCookieHeaders []string
    originalTransport     http.RoundTripper
}

func newWiretapTransport() *wiretapTransport {
    return &wiretapTransport{
        originalTransport: http.DefaultTransport,
    }
}

func (c *wiretapTransport) RoundTrip(r *http.Request) (*http.Response, error) {
    resp, err := c.originalTransport.RoundTrip(r)
    if resp != nil {
        cookie := resp.Header.Get("Set-Cookie")
        if cookie != "" {
            c.capturedCookieHeaders = append(c.capturedCookieHeaders, cookie)
        }
    }
    return resp, err
}

func (ws *WiretapService) callAPI(req *http.Request, responseChan chan *http.Response, errorChan chan error) {

    tr := newWiretapTransport()
    client := &http.Client{Transport: tr}

    configStore, _ := ws.controlsStore.Get(shared.ConfigKey)

    // create a new request from the original request, but replace the path

    config := configStore.(*shared.WiretapConfiguration)

    resp, err := client.Do(cloneRequest(req,
        config.RedirectProtocol,
        config.RedirectHost,
        config.RedirectPort))

    if err != nil {
        errorChan <- err
        close(errorChan)
    }

    fmt.Print(tr.capturedCookieHeaders)

    if len(tr.capturedCookieHeaders) > 0 {
        if resp.Header.Get("Set-Cookie") == "" {
            resp.Header.Set("Set-Cookie", tr.capturedCookieHeaders[0])
        }
    }

    responseChan <- resp
    close(responseChan)
}
