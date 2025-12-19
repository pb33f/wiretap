// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package shared

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"log/slog"
	"regexp"
	"strings"

	"github.com/pb33f/libopenapi/orderedmap"
	"go.yaml.in/yaml/v4"

	"github.com/gobwas/glob"
	"github.com/pb33f/harhar"
)

type ApiDocument struct {
	DocumentName string
	Document     libopenapi.Document
}
type ApiDocumentModel struct {
	DocumentName  string
	DocumentModel *libopenapi.DocumentModel[v3.Document]
}
type WiretapConfiguration struct {
	Contracts                   []string                                    `json:"-" yaml:"-"`
	PrimaryContract             string                                      `json:"-" yaml:"-"`
	RedirectHost                string                                      `json:"redirectHost,omitempty" yaml:"redirectHost,omitempty"`
	RedirectPort                string                                      `json:"redirectPort,omitempty" yaml:"redirectPort,omitempty"`
	RedirectBasePath            string                                      `json:"redirectBasePath,omitempty" yaml:"redirectBasePath,omitempty"`
	RedirectProtocol            string                                      `json:"redirectProtocol,omitempty" yaml:"redirectProtocol,omitempty"`
	RedirectURL                 string                                      `json:"redirectURL,omitempty" yaml:"redirectURL,omitempty"`
	Port                        string                                      `json:"port,omitempty" yaml:"port,omitempty"`
	MonitorPort                 string                                      `json:"monitorPort,omitempty" yaml:"monitorPort,omitempty"`
	WebSocketHost               string                                      `json:"webSocketHost,omitempty" yaml:"webSocketHost,omitempty"`
	WebSocketPort               string                                      `json:"webSocketPort,omitempty" yaml:"webSocketPort,omitempty"`
	GlobalAPIDelay              int                                         `json:"globalAPIDelay,omitempty" yaml:"globalAPIDelay,omitempty"`
	StaticDir                   string                                      `json:"staticDir,omitempty" yaml:"staticDir,omitempty"`
	StaticIndex                 string                                      `json:"staticIndex,omitempty" yaml:"staticIndex,omitempty"`
	PathConfigurations          *orderedmap.Map[string, *WiretapPathConfig] `json:"paths,omitempty" yaml:"paths,omitempty"`
	Headers                     *WiretapHeaderConfig                        `json:"headers,omitempty" yaml:"headers,omitempty"`
	StaticPaths                 []string                                    `json:"staticPaths,omitempty" yaml:"staticPaths,omitempty"`
	Variables                   map[string]string                           `json:"variables,omitempty" yaml:"variables,omitempty"`
	Spec                        string                                      `json:"contract,omitempty" yaml:"contract,omitempty"`
	Specs                       []string                                    `json:"contracts,omitempty" yaml:"contracts,omitempty"`
	Certificate                 string                                      `json:"certificate,omitempty" yaml:"certificate,omitempty"`
	CertificateKey              string                                      `json:"certificateKey,omitempty" yaml:"certificateKey,omitempty"`
	HardErrors                  bool                                        `json:"hardValidation,omitempty" yaml:"hardValidation,omitempty"`
	HardErrorCode               int                                         `json:"hardValidationCode,omitempty" yaml:"hardValidationCode,omitempty"`
	HardErrorReturnCode         int                                         `json:"hardValidationReturnCode,omitempty" yaml:"hardValidationReturnCode,omitempty"`
	HardErrorsList              []string                                    `json:"hardValidationList,omitempty" yaml:"hardValidationList,omitempty"`
	PathDelays                  map[string]int                              `json:"pathDelays,omitempty" yaml:"pathDelays,omitempty"`
	MockMode                    bool                                        `json:"mockMode,omitempty" yaml:"mockMode,omitempty"`
	MockModeList                []string                                    `json:"mockModeList,omitempty" yaml:"mockModeList,omitempty"`
	StaticMockDir               string                                      `json:"staticMockDir,omitempty" yaml:"staticMockDir,omitempty"`
	UseAllMockResponseFields    bool                                        `json:"useAllMockResponseFields,omitempty" yaml:"useAllMockResponseFields,omitempty"`
	MockModePretty              bool                                        `json:"mockModePretty,omitempty" yaml:"mockModePretty,omitempty"`
	Base                        string                                      `json:"base,omitempty" yaml:"base,omitempty"`
	HAR                         string                                      `json:"har,omitempty" yaml:"har,omitempty"`
	HARValidate                 bool                                        `json:"harValidate,omitempty" yaml:"harValidate,omitempty"`
	HARPathAllowList            []string                                    `json:"harPathAllowList,omitempty" yaml:"harPathAllowList,omitempty"`
	StreamReport                bool                                        `json:"streamReport,omitempty" yaml:"streamReport,omitempty"`
	ReportFile                  string                                      `json:"reportFilename,omitempty" yaml:"reportFilename,omitempty"`
	IgnoreRedirects             []string                                    `json:"ignoreRedirects,omitempty" yaml:"ignoreRedirects,omitempty"`
	RedirectAllowList           []string                                    `json:"redirectAllowList,omitempty" yaml:"redirectAllowList,omitempty"`
	WebsocketConfigs            map[string]*WiretapWebsocketConfig          `json:"websockets" yaml:"websockets"`
	IgnoreValidation            []string                                    `json:"ignoreValidation,omitempty" yaml:"ignoreValidation,omitempty"`
	ValidationAllowList         []string                                    `json:"validationAllowList,omitempty" yaml:"validationAllowList,omitempty"`
	StrictRedirectLocation      bool                                        `json:"strictRedirectLocation,omitempty" yaml:"strictRedirectLocation,omitempty"`
	IgnorePathRewrite           []*IgnoreRewriteConfig                      `json:"ignorePathRewrite,omitempty" yaml:"ignorePathRewrite,omitempty"`
	HARFile                     *harhar.HAR                                 `json:"-" yaml:"-"`
	CompiledMockModeList        []glob.Glob                                 `json:"-" yaml:"-"`
	CompiledPathDelays          map[string]*CompiledPathDelay               `json:"-" yaml:"-"`
	CompiledVariables           map[string]*CompiledVariable                `json:"-" yaml:"-"`
	Version                     string                                      `json:"-" yaml:"-"`
	StaticPathsCompiled         []glob.Glob                                 `json:"-" yaml:"-"`
	CompiledHardErrorList       []glob.Glob                                 `json:"-" yaml:"-"`
	CompiledPaths               *orderedmap.Map[string, *CompiledPath]      `json:"-"`
	CompiledIgnoreRedirects     []*CompiledRedirect                         `json:"-" yaml:"-"`
	CompiledRedirectAllowList   []*CompiledRedirect                         `json:"-" yaml:"-"`
	CompiledIgnoreValidations   []*CompiledRedirect                         `json:"-" yaml:"-"`
	CompiledValidationAllowList []*CompiledRedirect                         `json:"-" yaml:"-"`
	CompiledIgnorePathRewrite   []*CompiledIgnoreRewrite                    `json:"-" yaml:"-"`
	FS                          embed.FS                                    `json:"-"`
	Logger                      *slog.Logger
}

// UnmarshalJSON In order to initialize our ordered maps, we need to create custom un-marshallers.
// This implements json unmarshalling.
func (wtc *WiretapConfiguration) UnmarshalJSON(data []byte) error {
	wtc.PathConfigurations = orderedmap.New[string, *WiretapPathConfig]()
	wtc.CompiledPaths = orderedmap.New[string, *CompiledPath]()

	// Create a type alias to remove all methods, including the custom UnmarshalJSON method
	// This ensures that no infinite recursion happens
	type wiretapConfigAlias WiretapConfiguration
	aliasedConfig := (*wiretapConfigAlias)(wtc)

	return json.Unmarshal(data, aliasedConfig)
}

// UnmarshalYAML In order to initialize our ordered maps, we need to create custom un-marshallers.
// This implements yaml unmarshalling.
func (wtc *WiretapConfiguration) UnmarshalYAML(node *yaml.Node) error {
	wtc.PathConfigurations = orderedmap.New[string, *WiretapPathConfig]()
	wtc.CompiledPaths = orderedmap.New[string, *CompiledPath]()

	// Create a type alias to remove all methods, including the custom UnmarshalYAML method
	// This ensures that no infinite recursion happens
	type wiretapConfigAlias WiretapConfiguration
	aliasedConfig := (*wiretapConfigAlias)(wtc)

	return node.Decode(&aliasedConfig)
}

func (wtc *WiretapConfiguration) CompilePaths() {
	wtc.CompiledPaths = orderedmap.New[string, *CompiledPath]()
	for x := wtc.PathConfigurations.First(); x != nil; x = x.Next() {
		wtc.CompiledPaths.Set(x.Key(), x.Value().Compile(x.Key()))
	}
	if len(wtc.StaticPaths) > 0 {
		comp := make([]glob.Glob, len(wtc.StaticPaths))
		for x, path := range wtc.StaticPaths {
			comp[x] = glob.MustCompile(path)
		}
		wtc.StaticPathsCompiled = comp
	}
	if len(wtc.IgnorePathRewrite) > 0 {
		compiledPaths := make([]*CompiledIgnoreRewrite, len(wtc.IgnorePathRewrite))
		for x, ignoreRewrite := range wtc.IgnorePathRewrite {
			compiledPaths[x] = &CompiledIgnoreRewrite{
				RewriteTarget:         ignoreRewrite.RewriteTarget,
				CompiledIgnoreRewrite: glob.MustCompile(ignoreRewrite.Path),
			}
		}
		wtc.CompiledIgnorePathRewrite = compiledPaths
	}
}

func (wtc *WiretapConfiguration) CompilePathDelays() {
	wtc.CompiledPathDelays = make(map[string]*CompiledPathDelay)
	for k, v := range wtc.PathDelays {
		compiled := &CompiledPathDelay{
			CompiledPathDelay: glob.MustCompile(wtc.ReplaceWithVariables(k)),
			PathDelayValue:    v,
		}
		wtc.CompiledPathDelays[k] = compiled
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

func (wtc *WiretapConfiguration) CompileMockModeList() {
	wtc.CompiledMockModeList = make([]glob.Glob, 0)
	for _, x := range wtc.MockModeList {
		wtc.CompiledMockModeList = append(wtc.CompiledMockModeList, glob.MustCompile(wtc.ReplaceWithVariables((x))))
	}
}

func (wtc *WiretapConfiguration) CompileHardErrorList() {
	wtc.CompiledHardErrorList = make([]glob.Glob, 0)
	for _, x := range wtc.HardErrorsList {
		wtc.CompiledHardErrorList = append(wtc.CompiledHardErrorList, glob.MustCompile(wtc.ReplaceWithVariables((x))))
	}
}

func (wtc *WiretapConfiguration) CompileIgnoreRedirects() {
	wtc.CompiledIgnoreRedirects = make([]*CompiledRedirect, 0)
	for _, x := range wtc.IgnoreRedirects {
		compiled := &CompiledRedirect{
			CompiledPath: glob.MustCompile(wtc.ReplaceWithVariables(x)),
		}
		wtc.CompiledIgnoreRedirects = append(wtc.CompiledIgnoreRedirects, compiled)
	}
}

func (wtc *WiretapConfiguration) CompileRedirectAllowList() {
	wtc.CompiledRedirectAllowList = make([]*CompiledRedirect, 0)
	for _, x := range wtc.RedirectAllowList {
		compiled := &CompiledRedirect{
			CompiledPath: glob.MustCompile(wtc.ReplaceWithVariables(x)),
		}
		wtc.CompiledRedirectAllowList = append(wtc.CompiledRedirectAllowList, compiled)
	}
}

func (wtc *WiretapConfiguration) CompileIgnoreValidations() {
	wtc.CompiledIgnoreValidations = make([]*CompiledRedirect, 0)
	for _, x := range wtc.IgnoreValidation {
		compiled := &CompiledRedirect{
			CompiledPath: glob.MustCompile(wtc.ReplaceWithVariables(x)),
		}
		wtc.CompiledIgnoreValidations = append(wtc.CompiledIgnoreValidations, compiled)
	}
}

func (wtc *WiretapConfiguration) CompileValidationAllowList() {
	wtc.CompiledValidationAllowList = make([]*CompiledRedirect, 0)
	for _, x := range wtc.ValidationAllowList {
		compiled := &CompiledRedirect{
			CompiledPath: glob.MustCompile(wtc.ReplaceWithVariables(x)),
		}
		wtc.CompiledValidationAllowList = append(wtc.CompiledValidationAllowList, compiled)
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

func (wtc *WiretapConfiguration) GetHttpProtocol() string {
	protocol := "http"

	if wtc.CertificateKey != "" && wtc.Certificate != "" {
		protocol = "https"
	}

	return protocol
}

func (wtc *WiretapConfiguration) GetApiGateway() string {
	return fmt.Sprintf("%s://%s", wtc.GetHttpProtocol(), wtc.GetApiGatewayHost())
}

func (wtc *WiretapConfiguration) GetApiGatewayHost() string {
	return fmt.Sprintf("localhost:%s", wtc.Port)
}

func (wtc *WiretapConfiguration) GetMonitorUI() string {
	return fmt.Sprintf("%s://localhost:%s", wtc.GetHttpProtocol(), wtc.MonitorPort)
}

func (wtc *WiretapConfiguration) GetContractList() string {
	return strings.Join(wtc.Contracts, ", ")
}

type WiretapWebsocketConfig struct {
	VerifyCert  *bool    `json:"verifyCert" yaml:"verifyCert"`
	DropHeaders []string `json:"dropHeaders" yaml:"dropHeaders"`
}

type WiretapPathConfig struct {
	Target                string                   `json:"target,omitempty" yaml:"target,omitempty"`
	PathRewrite           map[string]string        `json:"pathRewrite,omitempty" yaml:"pathRewrite,omitempty"`
	ChangeOrigin          bool                     `json:"changeOrigin,omitempty" yaml:"changeOrigin,omitempty"`
	Headers               *WiretapHeaderConfig     `json:"headers,omitempty" yaml:"headers,omitempty"`
	Secure                bool                     `json:"secure,omitempty" yaml:"secure,omitempty"`
	Auth                  string                   `json:"auth,omitempty" yaml:"auth,omitempty"`
	RewriteId             string                   `json:"rewriteId,omitempty" yaml:"rewriteId,omitempty"`
	IgnoreRewrite         []*IgnoreRewriteConfig   `json:"ignoreRewrite,omitempty" yaml:"ignoreRewrite,omitempty"`
	CompiledPath          *CompiledPath            `json:"-"`
	CompiledIgnoreRewrite []*CompiledIgnoreRewrite `json:"-"`
}

type IgnoreRewriteConfig struct {
	RewriteTarget bool   `json:"rewriteTarget,omitempty" yaml:"rewriteTarget,omitempty"`
	Path          string `json:"path,omitempty" yaml:"path,omitempty"`
}

type CompiledIgnoreRewrite struct {
	RewriteTarget         bool
	CompiledIgnoreRewrite glob.Glob
}

type CompiledPath struct {
	PathConfig          *WiretapPathConfig
	CompiledKey         glob.Glob
	CompiledTarget      glob.Glob
	CompiledPathRewrite map[string]*regexp.Regexp
}

type CompiledPathDelay struct {
	CompiledPathDelay glob.Glob
	PathDelayValue    int
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

type CompiledRedirect struct {
	CompiledPath glob.Glob
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

	wpc.CompiledIgnoreRewrite = make([]*CompiledIgnoreRewrite, len(wpc.IgnoreRewrite))
	for i, ignoreRewrite := range wpc.IgnoreRewrite {
		wpc.CompiledIgnoreRewrite[i] = &CompiledIgnoreRewrite{
			RewriteTarget:         ignoreRewrite.RewriteTarget,
			CompiledIgnoreRewrite: glob.MustCompile(ignoreRewrite.Path),
		}
	}
	return cp
}

const ConfigKey = "config"
const HARKey = "har"
const WiretapHostPlaceholder = "%WIRETAP_HOST%"
const WiretapPortPlaceholder = "%WIRETAP_PORT%"
const WiretapTLSPlaceholder = "%WIRETAP_TLS%"
const WiretapVersionPlaceholder = "%WIRETAP_VERSION%"
const IndexFile = "index.html"
const UILocation = "ui/dist"
const UIAssetsLocation = "ui/dist/assets"
