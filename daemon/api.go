// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package daemon

import (
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

func (ws *WiretapService) newWiretapTransport() *wiretapTransport {
	return &wiretapTransport{
		originalTransport: ws.transport,
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

func (ws *WiretapService) callAPI(req *http.Request, wiretapConfig ...*shared.WiretapConfiguration) (*http.Response, error) {

	var cfg *shared.WiretapConfiguration
	if len(wiretapConfig) > 0 && wiretapConfig[0] != nil {
		cfg = wiretapConfig[0]
	} else {
		configStore, _ := ws.controlsStore.Get(shared.ConfigKey)
		cfg = configStore.(*shared.WiretapConfiguration)
	}

	// lookup path and determine if we need to redirect it.
	replaced := config.RewritePath(req.URL.Path, req, cfg)
	if replaced.RewrittenPath != req.URL.Path {
		newUrl, _ := url.Parse(replaced.RewrittenPath)
		if req.URL.RawQuery != "" {
			newUrl.RawQuery = req.URL.RawQuery
		}
		if replaced.PathConfiguration != nil && replaced.PathConfiguration.RewriteId != "" {
			rewriteId := replaced.PathConfiguration.RewriteId
			pterm.Info.Printf("[wiretap] Re-writing path '%s' to '%s' with identifier '%s'\n", req.URL.String(), newUrl.String(), rewriteId)
		} else {
			pterm.Info.Printf("[wiretap] Re-writing path '%s' to '%s'\n", req.URL.String(), newUrl.String())
		}
		req.URL = newUrl
	}

	tr := ws.newWiretapTransport()
	var client *http.Client

	// create a client based on if wiretap should redirect on the path or not
	if config.IgnoreRedirectOnPath(req.URL.Path, cfg) && !config.PathRedirectAllowListed(req.URL.Path, cfg) {
		client = &http.Client{
			Transport: tr,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	} else {
		client = &http.Client{Transport: tr}
	}

	// re-write referer
	if req.Header.Get("Referer") != "" {
		// retain original referer for logging
		req.Header.Set("X-Original-Referer", req.Header.Get("Referer"))
		req.Header.Set("Referer", ReconstructURL(req,
			cfg.RedirectProtocol,
			cfg.RedirectHost,
			cfg.RedirectBasePath,
			cfg.RedirectPort))
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
