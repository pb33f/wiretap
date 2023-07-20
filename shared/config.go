// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package shared

import (
	"embed"
	"fmt"
	"github.com/gobwas/glob"
	"regexp"
)

type WiretapConfiguration struct {
	Contract            string                        `json:"-" yaml:"-"`
	RedirectHost        string                        `json:"redirectHost,omitempty" yaml:"redirectHost,omitempty"`
	RedirectPort        string                        `json:"redirectPort,omitempty" yaml:"redirectPort,omitempty"`
	RedirectBasePath    string                        `json:"redirectBasePath,omitempty" yaml:"redirectBasePath,omitempty"`
	RedirectProtocol    string                        `json:"redirectProtocol,omitempty" yaml:"redirectProtocol,omitempty"`
	RedirectURL         string                        `json:"redirectURL,omitempty" yaml:"redirectURL,omitempty"`
	Port                string                        `json:"port,omitempty" yaml:"port,omitempty"`
	MonitorPort         string                        `json:"monitorPort,omitempty" yaml:"monitorPort,omitempty"`
	WebSocketPort       string                        `json:"webSocketPort,omitempty" yaml:"webSocketPort,omitempty"`
	GlobalAPIDelay      int                           `json:"globalAPIDelay,omitempty" yaml:"globalAPIDelay,omitempty"`
	StaticDir           string                        `json:"staticDir,omitempty" yaml:"staticDir,omitempty"`
	StaticIndex         string                        `json:"staticIndex,omitempty" yaml:"staticIndex,omitempty"`
	PathConfigurations  map[string]*WiretapPathConfig `json:"paths,omitempty" yaml:"paths,omitempty"`
	Headers             *WiretapHeaderConfig          `json:"headers,omitempty" yaml:"headers,omitempty"`
	StaticPaths         []string                      `json:"staticPaths,omitempty" yaml:"staticPaths,omitempty"`
	Variables           map[string]string             `json:"variables,omitempty" yaml:"variables,omitempty"`
	Spec                string                        `json:"contract,omitempty" yaml:"contract,omitempty"`
	CompiledVariables   map[string]*CompiledVariable  `json:"-" yaml:"-"`
	StaticPathsCompiled []glob.Glob                   `json:"-" yaml:"-"`
	CompiledPaths       map[string]*CompiledPath      `json:"-"`
	FS                  embed.FS                      `json:"-"`
}

func (wtc *WiretapConfiguration) CompilePaths() {
	wtc.CompiledPaths = make(map[string]*CompiledPath)
	for x := range wtc.PathConfigurations {
		wtc.CompiledPaths[x] = wtc.PathConfigurations[x].Compile(x)
	}
	if len(wtc.StaticPaths) > 0 {
		comp := make([]glob.Glob, len(wtc.StaticPaths))
		for x, path := range wtc.StaticPaths {
			comp[x] = glob.MustCompile(path)
		}
		wtc.StaticPathsCompiled = comp
	}
}

func (wtc *WiretapConfiguration) CompileVariables() {
	wtc.CompiledVariables = make(map[string]*CompiledVariable)
	for x := range wtc.Variables {
		compiled := &CompiledVariable{
			CompiledVariable: regexp.MustCompile(fmt.Sprintf("\\${(%s)}", x)),
			VariableValue:    wtc.Variables[x],
		}
		wtc.CompiledVariables[x] = compiled
	}
}

func (wtc *WiretapConfiguration) ReplaceWithVariables(input string) string {
	for x := range wtc.Variables {
		if wtc.Variables[x] != "" && wtc.CompiledVariables[x] != nil {
			input = wtc.CompiledVariables[x].CompiledVariable.
				ReplaceAllString(input, wtc.CompiledVariables[x].VariableValue)
		}
	}
	return input
}

type WiretapPathConfig struct {
	Target       string               `json:"target,omitempty" yaml:"target,omitempty"`
	PathRewrite  map[string]string    `json:"pathRewrite,omitempty" yaml:"pathRewrite,omitempty"`
	ChangeOrigin bool                 `json:"changeOrigin,omitempty" yaml:"changeOrigin,omitempty"`
	Headers      *WiretapHeaderConfig `json:"headers,omitempty" yaml:"headers,omitempty"`
	Secure       bool                 `json:"secure,omitempty" yaml:"secure,omitempty"`
	Auth         string               `json:"auth,omitempty" yaml:"auth,omitempty"`
	CompiledPath *CompiledPath        `json:"-"`
}

type CompiledPath struct {
	PathConfig          *WiretapPathConfig
	CompiledKey         glob.Glob
	CompiledTarget      glob.Glob
	CompiledPathRewrite map[string]*regexp.Regexp
}

type CompiledVariable struct {
	CompiledVariable *regexp.Regexp
	VariableValue    string
}

type CompiledPathRewrite struct {
	PathConfig     *WiretapPathConfig
	Key            string
	CompiledKey    glob.Glob
	CompiledTarget glob.Glob
}

type WiretapHeaderConfig struct {
	DropHeaders    []string          `json:"drop,omitempty" yaml:"drop,omitempty"`
	InjectHeaders  map[string]string `json:"inject,omitempty" yaml:"inject,omitempty"`
	RewriteHeaders map[string]string `json:"rewrite,omitempty" yaml:"rewrite,omitempty"`
}

func (wpc *WiretapPathConfig) Compile(key string) *CompiledPath {
	cp := &CompiledPath{
		PathConfig:     wpc,
		CompiledKey:    glob.MustCompile(key),
		CompiledTarget: glob.MustCompile(wpc.Target),
	}
	wpc.CompiledPath = cp
	cp.CompiledPathRewrite = make(map[string]*regexp.Regexp)
	for x := range wpc.PathRewrite {
		cp.CompiledPathRewrite[x] = regexp.MustCompile(x)
	}
	return cp
}

const ConfigKey = "config"
const WiretapPortPlaceholder = "%WIRETAP_PORT%"
const IndexFile = "index.html"
const UILocation = "ui/dist"
const UIAssetsLocation = "ui/dist/assets"
