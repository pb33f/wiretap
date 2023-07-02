// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package shared

import (
	"embed"
	"github.com/gobwas/glob"
)

type WiretapConfiguration struct {
	Contract           string                        `json:"-"`
	RedirectHost       string                        `json:"redirectHost,omitempty"`
	RedirectPort       string                        `json:"redirectPort,omitempty"`
	RedirectBasePath   string                        `json:"redirectBasePath,omitempty"`
	RedirectProtocol   string                        `json:"redirectProtocol,omitempty"`
	RedirectURL        string                        `json:"redirectURL,omitempty"`
	Port               string                        `json:"port,omitempty"`
	MonitorPort        string                        `json:"monitorPort,omitempty"`
	WebSocketPort      string                        `json:"webSocketPort,omitempty"`
	GlobalAPIDelay     int                           `json:"globalAPIDelay,omitempty"`
	StaticDir          string                        `json:"staticDir,omitempty"`
	StaticPort         string                        `json:"staticPort,omitempty"`
	PathConfigurations map[string]*WiretapPathConfig `json:"paths,omitempty"`
	CompiledPaths      map[string]*CompiledPath      `json:"-"`
	FS                 embed.FS                      `json:"-"`
}

func (wtc *WiretapConfiguration) CompilePaths() {
	wtc.CompiledPaths = make(map[string]*CompiledPath)
	for x := range wtc.PathConfigurations {
		wtc.CompiledPaths[x] = wtc.PathConfigurations[x].Compile(x)
	}
}

type WiretapPathConfig struct {
	Target      string            `json:"target,omitempty"`
	PathRewrite map[string]string `json:"pathRewrite,omitempty"`
	Secure      bool              `json:"secure,omitempty"`
}

type CompiledPath struct {
	PathConfig     *WiretapPathConfig
	CompiledKey    glob.Glob
	CompiledTarget glob.Glob
}

func (wpc *WiretapPathConfig) Compile(key string) *CompiledPath {
	cp := &CompiledPath{
		PathConfig:     wpc,
		CompiledKey:    glob.MustCompile(key),
		CompiledTarget: glob.MustCompile(wpc.Target),
	}
	return cp
}

const ConfigKey = "config"
const WiretapPortPlaceholder = "%WIRETAP_PORT%"
const IndexFile = "index.html"
const UILocation = "ui/dist"
const UIAssetsLocation = "ui/dist/assets"
