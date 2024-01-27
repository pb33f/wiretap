// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package daemon

import (
	"crypto/tls"
	"net/http"
	"net/url"

	"github.com/pb33f/wiretap/config"
	"github.com/pterm/pterm"

	"github.com/pb33f/wiretap/shared"
)

type wiretapTransport struct {
	capturedCookieHeaders []string
	originalTransport     http.RoundTripper
}

func newWiretapTransport() *wiretapTransport {
	// Disable ssl cert checks
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
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

func (ws *WiretapService) callAPI(req *http.Request) (*http.Response, error) {

	tr := newWiretapTransport()
	client := &http.Client{Transport: tr}

	configStore, _ := ws.controlsStore.Get(shared.ConfigKey)

	// create a new request from the original request, but replace the path
	wiretapConfig := configStore.(*shared.WiretapConfiguration)

	// lookup path and determine if we need to redirect it.
	replaced := config.RewritePath(req.URL.Path, wiretapConfig)
	if replaced != req.URL.Path {
		newUrl, _ := url.Parse(replaced)
		if req.URL.RawQuery != "" {
			newUrl.RawQuery = req.URL.RawQuery
		}
		pterm.Info.Printf("[wiretap] Re-writing path '%s' to '%s'\n", req.URL.String(), newUrl.String())
		req.URL = newUrl
	}

	// re-write referer
	if req.Header.Get("Referer") != "" {
		// retain original referer for logging
		req.Header.Set("X-Original-Referer", req.Header.Get("Referer"))
		req.Header.Set("Referer", ReconstructURL(req,
			wiretapConfig.RedirectProtocol,
			wiretapConfig.RedirectHost,
			wiretapConfig.RedirectBasePath,
			wiretapConfig.RedirectPort))
	}
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	if len(tr.capturedCookieHeaders) > 0 {
		if resp.Header.Get("Set-Cookie") == "" {
			resp.Header.Set("Set-Cookie", tr.capturedCookieHeaders[0])
		}
	}
	return resp, nil
}
