// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package daemon

import (
	"fmt"
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

	// create a new request from the original request, but replace the path
	resp, err := client.Do(cloneRequest(req))
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
