// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package proxy

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	configModel "github.com/pb33f/wiretap/config"
	"github.com/pb33f/wiretap/shared"
	"github.com/pterm/pterm"
)

type wiretapTransport struct {
	capturedCookieHeaders []string
	originalTransport     http.RoundTripper
}

func (h *Handler) newWiretapTransport() *wiretapTransport {
	return &wiretapTransport{
		originalTransport: h.transport,
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

func (h *Handler) callAPI(req *http.Request, wiretapConfig ...*shared.WiretapConfiguration) (*http.Response, error) {
	if len(wiretapConfig) == 0 || wiretapConfig[0] == nil {
		return nil, fmt.Errorf("wiretap configuration is required to call upstream API")
	}
	cfg := wiretapConfig[0]

	replaced := configModel.RewritePath(req.URL.Path, req, cfg)
	if replaced.RewrittenPath != req.URL.Path {
		newURL, _ := url.Parse(replaced.RewrittenPath)
		if req.URL.RawQuery != "" {
			newURL.RawQuery = req.URL.RawQuery
		}
		if replaced.PathConfiguration != nil && replaced.PathConfiguration.RewriteId != "" {
			rewriteID := replaced.PathConfiguration.RewriteId
			pterm.Info.Printf("[wiretap] Re-writing path '%s' to '%s' with identifier '%s'\n", req.URL.String(), newURL.String(), rewriteID)
		} else {
			pterm.Info.Printf("[wiretap] Re-writing path '%s' to '%s'\n", req.URL.String(), newURL.String())
		}
		req.URL = newURL
	}

	tr := h.newWiretapTransport()
	var client *http.Client
	if configModel.IgnoreRedirectOnPath(req.URL.Path, cfg) &&
		!configModel.PathRedirectAllowListed(req.URL.Path, cfg) {
		client = &http.Client{
			Transport: tr,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	} else {
		client = &http.Client{Transport: tr}
	}

	if req.Header.Get("Referer") != "" {
		req.Header.Set("X-Original-Referer", req.Header.Get("Referer"))
		req.Header.Set("Referer", reconstructURL(
			req,
			cfg.RedirectProtocol,
			cfg.RedirectHost,
			cfg.RedirectBasePath,
			cfg.RedirectPort,
		))
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if len(tr.capturedCookieHeaders) > 0 && resp.Header.Get("Set-Cookie") == "" {
		resp.Header.Set("Set-Cookie", tr.capturedCookieHeaders[0])
	}
	return resp, nil
}

func reconstructURL(r *http.Request, protocol, host, basepath string, port string) string {
	if host == "" {
		host = r.Host
	}
	if protocol == "" {
		protocol = "http"
	}
	var b strings.Builder
	b.Grow(len(protocol) + 3 + len(host) + 1 + len(port) + len(basepath) + len(r.URL.Path) + 1 + len(r.URL.RawQuery))
	b.WriteString(protocol)
	b.WriteString("://")
	b.WriteString(host)
	if port != "" {
		b.WriteByte(':')
		b.WriteString(port)
	}
	if basepath != "" {
		b.WriteString(basepath)
	}
	if r.URL.Path != "" {
		b.WriteString(r.URL.Path)
	}
	if r.URL.RawQuery != "" {
		b.WriteByte('?')
		b.WriteString(r.URL.RawQuery)
	}
	return b.String()
}
