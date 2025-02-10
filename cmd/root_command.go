// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package cmd

import (
	"embed"
	"encoding/json"
	"errors"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"

	"github.com/pb33f/harhar"
	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/pb33f/wiretap/har"
	"github.com/pb33f/wiretap/shared"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	Version string
	Commit  string
	Date    string
	FS      embed.FS

	rootCmd = &cobra.Command{
		SilenceUsage: true,
		Use:          "wiretap",
		Short:        "wiretap is a tool for detecting API compliance against an OpenAPI contract, by sniffing network traffic.",
		Long:         `wiretap is a tool for detecting API compliance against an OpenAPI contract, by sniffing network traffic.`,
		RunE: func(cmd *cobra.Command, args []string) error {

			PrintBanner()

			configFlag, _ := cmd.Flags().GetString("config")

			var spec string
			var port string
			var monitorPort string
			var wsPort string
			var staticDir string
			var staticIndex string
			var redirectHost string
			var redirectPort string
			var redirectScheme string
			var redirectBasePath string
			var redirectURL string
			var globalAPIDelay int
			var wsHost string

			// certs
			var cert string
			var certKey string

			// hard errors
			var hardError bool
			var hardErrorCode int
			var hardErrorReturnCode int

			// mock mode
			var mockMode bool
			var useAllMockResponseFields bool

			certFlag, _ := cmd.Flags().GetString("cert")
			if certFlag != "" {
				cert = certFlag
			}

			keyFlag, _ := cmd.Flags().GetString("key")
			if keyFlag != "" {
				certKey = keyFlag
			}
			base, _ := cmd.Flags().GetString("base")
			reportFilename, _ := cmd.Flags().GetString("report-filename")

			harFlag, _ := cmd.Flags().GetString("har")
			harValidate, _ := cmd.Flags().GetBool("har-validate")
			harWhiteList, _ := cmd.Flags().GetStringArray("har-allow")

			debug, _ := cmd.Flags().GetBool("debug")
			mockMode, _ = cmd.Flags().GetBool("mock-mode")
			useAllMockResponseFields, _ = cmd.Flags().GetBool("enable-all-mock-response-fields")
			hardError, _ = cmd.Flags().GetBool("hard-validation")
			hardErrorCode, _ = cmd.Flags().GetInt("hard-validation-code")
			hardErrorReturnCode, _ = cmd.Flags().GetInt("hard-validation-return-code")
			streamReport, _ := cmd.Flags().GetBool("stream-report")
			strictRedirectLocation, _ := cmd.Flags().GetBool("strict-redirect-location")

			portFlag, _ := cmd.Flags().GetString("port")
			if portFlag != "" {
				port = portFlag
			} else {
				port = "9090" // default
			}

			specFlag, _ := cmd.Flags().GetString("spec")
			if specFlag != "" {
				spec = specFlag
			}

			monitorPortFlag, _ := cmd.Flags().GetString("monitor-port")
			if monitorPortFlag != "" {
				monitorPort = monitorPortFlag
			} else {
				monitorPort = "9091" // default
			}

			staticDirFlag, _ := cmd.Flags().GetString("static")
			if staticDirFlag != "" {
				staticDir = staticDirFlag
			}

			staticIndex, _ = cmd.Flags().GetString("static-index")

			wsHostFlag, _ := cmd.Flags().GetString("ws-host")
			if wsHostFlag != "" {
				wsHost = wsHostFlag
			} else {
				wsHost = "localhost"
			}

			wsPortFlag, _ := cmd.Flags().GetString("ws-port")
			if wsPortFlag != "" {
				wsPort = wsPortFlag
			} else {
				wsPort = "9092" // default
			}

			redirectURLFlag, _ := cmd.Flags().GetString("url")
			if redirectURLFlag != "" {
				redirectURL = redirectURLFlag
			}

			globalAPIDelayFlag, _ := cmd.Flags().GetInt("delay")
			if globalAPIDelayFlag > 0 {
				globalAPIDelay = globalAPIDelayFlag
			}

			var config shared.WiretapConfiguration

			// set version.
			config.Version = Version

			if configFlag == "" {
				// see if a configuration file exists in the current directory or in the user's home directory.
				local, _ := os.Stat("wiretap.yaml")
				home, _ := os.Stat(filepath.Join(os.Getenv("HOME"), "wiretap.yaml"))
				if home != nil {
					configFlag = filepath.Join(os.Getenv("HOME"), "wiretap.yaml")
				}
				if local != nil {
					configFlag = local.Name()
				}
			}

			if configFlag != "" {

				cBytes, err := os.ReadFile(configFlag)
				if err != nil {
					pterm.Error.Printf("Failed to read wiretap configuration '%s': %s\n", configFlag, err.Error())
					return err
				}
				err = yaml.Unmarshal(cBytes, &config)
				if err != nil {
					pterm.Error.Printf("Failed to parse wiretap configuration '%s': %s\n", configFlag, err.Error())
					return err
				}
				pterm.Info.Printf("Loaded wiretap configuration '%s'...\n\n", configFlag)
				if config.RedirectURL != "" {
					redirectURL = config.RedirectURL
				}
				if config.StaticIndex == "" {
					config.StaticIndex = staticIndex
				}
				if config.Spec != "" {
					spec = config.Spec
				}
				if mockMode {
					if !config.MockMode {
						config.MockMode = true
					}
				}
				if useAllMockResponseFields {
					if !config.UseAllMockResponseFields {
						config.UseAllMockResponseFields = true
					}
				}
				if streamReport {
					if !config.StreamReport {
						config.StreamReport = true
					}
				}
				if strictRedirectLocation {
					if !config.StrictRedirectLocation {
						config.StrictRedirectLocation = true
					}
				}

				if reportFilename != "" {
					config.ReportFile = reportFilename
				}

				if base != config.Base {
					config.Base = base
				}
				if harFlag != "" && harFlag != config.HAR {
					config.HAR = harFlag
				}
				if harValidate && !config.HARValidate {
					config.HARValidate = harValidate
				}
				if len(harWhiteList) > 0 {
					config.HARPathAllowList = harWhiteList
				}

			} else {

				pterm.Info.Println("No wiretap configuration located. Using defaults")
				config.StaticIndex = staticIndex
				if mockMode {
					config.MockMode = true
				}
				if useAllMockResponseFields {
					config.UseAllMockResponseFields = true
				}
				if streamReport {
					config.StreamReport = true
				}
				if strictRedirectLocation {
					config.StrictRedirectLocation = true
				}
				if base != "" {
					config.Base = base
				}
				config.ReportFile = reportFilename
				config.HAR = harFlag
				config.HARValidate = harValidate
				config.HARPathAllowList = harWhiteList
			}

			if spec == "" {
				pterm.Println()
				pterm.Warning.Println("No OpenAPI specification provided. " +
					"Please provide a path to an OpenAPI specification using the --spec or -s flags. \n" +
					"Without an OpenAPI specification, wiretap will not be able to validate " +
					"requests and responses")
				pterm.Println()
			}

			if mockMode && spec == "" {
				pterm.Println()
				pterm.Error.Println("Cannot enable mock mode, no OpenAPI specification provided!\n" +
					"Please provide a path to an OpenAPI specification using the --spec or -s flags.\n" +
					"Without an OpenAPI specification, wiretap will not be able to generate mock responses")
				pterm.Println()
				return nil
			}

			if !mockMode && redirectURL == "" && harFlag == "" {
				pterm.Println()
				pterm.Error.Println("No redirect URL provided. " +
					"Please provide a URL to redirect API traffic to using the --url or -u flags.")
				pterm.Println()
				return nil
			}

			if redirectURL != "" {

				parsedURL, e := url.Parse(redirectURL)
				if e != nil {
					pterm.Println()
					pterm.Error.Printf("URL is not valid. "+
						"Please provide a valid URL to redirect to. %s cannot be parsed\n\n", redirectURL)
					pterm.Println()
					return nil
				}
				if parsedURL.Scheme == "" || parsedURL.Host == "" {
					pterm.Println()
					pterm.Error.Printf("URL is not valid. "+
						"Please provide a valid URL to redirect to. %s cannot be parsed\n\n", redirectURL)
					pterm.Println()
					return nil
				}
				redirectHost = parsedURL.Hostname()
				redirectPort = parsedURL.Port()
				redirectScheme = parsedURL.Scheme
				redirectBasePath = parsedURL.Path
			}

			if spec != "" {
				config.Contract = spec
			}
			config.RedirectURL = redirectURL
			config.RedirectHost = redirectHost
			config.RedirectBasePath = redirectBasePath
			config.RedirectPort = redirectPort
			config.RedirectProtocol = redirectScheme

			if config.Port == "" {
				config.Port = port
			}
			if config.MonitorPort == "" {
				config.MonitorPort = monitorPort
			}
			if config.WebSocketPort == "" {
				config.WebSocketPort = wsPort
			}
			if config.WebSocketHost == "" {
				config.WebSocketHost = wsHost
			}
			if config.GlobalAPIDelay == 0 {
				config.GlobalAPIDelay = globalAPIDelay
			}
			if config.StaticDir == "" {
				config.StaticDir = staticDir
			}
			config.FS = FS

			if config.HardErrors || hardError {
				config.HardErrors = true
			}

			// configure hard errors if set
			if (config.HardErrors || len(config.HardErrorsList) > 0) && config.HardErrorCode <= 0 {
				config.HardErrorCode = hardErrorCode
			}
			if (config.HardErrors || len(config.HardErrorsList) > 0) && config.HardErrorReturnCode <= 0 {
				config.HardErrorReturnCode = hardErrorReturnCode
			}

			// certs
			if config.Certificate == "" && config.CertificateKey == "" {
				config.Certificate = cert
				config.CertificateKey = certKey
			}

			// variables
			if len(config.Variables) > 0 {
				config.CompileVariables()
				printLoadedVariables(config.Variables)
			}

			// paths
			if config.PathConfigurations != nil && config.PathConfigurations.Len() > 0 || len(config.StaticPaths) > 0 || len(config.HARPathAllowList) > 0 || len(config.IgnorePathRewrite) > 0 {
				config.CompilePaths()
				if len(config.IgnorePathRewrite) > 0 {
					printLoadedIgnorePathRewrite(config.IgnorePathRewrite)
				}

				if config.PathConfigurations.Len() > 0 {
					printLoadedPathConfigurations(config.PathConfigurations)
				}
			}

			// path delays
			if len(config.PathDelays) > 0 {
				config.CompilePathDelays()
				printLoadedPathDelayConfigurations(config.PathDelays)
			}

			if len(config.IgnoreRedirects) > 0 {
				config.CompileIgnoreRedirects()
				printLoadedIgnoreRedirectPaths(config.IgnoreRedirects)
			}

			if len(config.RedirectAllowList) > 0 {
				config.CompileRedirectAllowList()
				printLoadedRedirectAllowList(config.RedirectAllowList)
			}

			if len(config.WebsocketConfigs) > 0 {
				for _, config := range config.WebsocketConfigs {
					if config.VerifyCert == nil {
						config.VerifyCert = func() *bool { b := true; return &b }()
					}
				}

				printLoadedWebsockets(config.WebsocketConfigs)
			}

			if len(config.MockModeList) > 0 && !config.MockMode {
				config.CompileMockModeList()
				printLoadedMockModeList(config.MockModeList)
			}

			if len(config.HardErrorsList) > 0 && !config.HardErrors {
				config.CompileHardErrorList()
				printLoadedHardErrorList(config.MockModeList)
			}

			if len(config.IgnoreValidation) > 0 {
				config.CompileIgnoreValidations()
				printLoadedIgnoreValidationPaths(config.IgnoreValidation)
			}

			if len(config.ValidationAllowList) > 0 {
				config.CompileValidationAllowList()
				printLoadedValidationAllowList(config.ValidationAllowList)
			}

			// static headers
			if config.Headers != nil && len(config.Headers.DropHeaders) > 0 {
				pterm.Info.Printf("Dropping the following %d %s globally:\n", len(config.Headers.DropHeaders),
					shared.Pluralize(len(config.Headers.DropHeaders), "header", "headers"))
				for _, header := range config.Headers.DropHeaders {
					pterm.Printf("🗑️  %s\n", pterm.LightRed(header))
				}
				pterm.Println()
			}

			// static paths
			if len(config.StaticPaths) > 0 && config.StaticDir != "" {
				staticPath := filepath.Join(config.StaticDir, config.StaticIndex)
				pterm.Info.Printf("Mapping %d static %s to '%s':\n", len(config.StaticPaths),
					shared.Pluralize(len(config.StaticPaths), "path", "paths"), staticPath)
				for _, path := range config.StaticPaths {
					pterm.Printf("⛱️  %s\n", pterm.LightMagenta(path))
				}
				pterm.Println()
			}

			// hard errors
			if config.HardErrors {
				pterm.Printf("❌  Hard validation mode enabled. HTTP error %s for requests and error %s for responses that "+
					"fail to pass validation.\n",
					pterm.LightRed(config.HardErrorCode), pterm.LightRed(config.HardErrorReturnCode))
				pterm.Println()
			}

			// mock mode
			if config.MockMode {
				pterm.Printf("Ⓜ️ %s. All responses will be mocked and no traffic will be sent to the target API.\n",
					pterm.LightCyan("Mock mode enabled"))
				pterm.Println()
			}

			// using TLS?
			if config.CertificateKey != "" && config.Certificate != "" {
				pterm.Printf("🔐 Running over %s using certificate: %s and key: %s\n",
					pterm.LightYellow("TLS/HTTPS & HTTP/2"), pterm.LightMagenta(config.Certificate), pterm.LightCyan(config.CertificateKey))
				pterm.Println()
			}

			// streaming violations?
			if config.StreamReport {
				pterm.Printf("⏩  Streaming API violations to file: %s\n", pterm.LightMagenta(config.ReportFile))
				pterm.Println()
			}

			var harBytes []byte
			var harFile *harhar.HAR

			// check if we're using a HAR file
			if config.HAR != "" {
				pterm.Println()
				pterm.Printf("📦 Loading HAR file: %s\n", pterm.LightMagenta(config.HAR))
				// can we read the har file?
				_, err := os.Stat(config.HAR)
				if err != nil {
					pterm.Error.Printf("Cannot read HAR file: %s (%s)\n", config.HAR, err.Error())
					return nil
				}

				var fErr error
				harBytes, fErr = os.ReadFile(config.HAR)
				if fErr != nil {
					pterm.Error.Printf("Cannot read HAR file: %s (%s)\n", config.HAR, fErr.Error())
					return nil
				}

				harFile, fErr = har.BuildHAR(harBytes)
				if fErr != nil {
					pterm.Error.Printf("Cannot parse HAR file: %s (%s)\n", config.HAR, fErr.Error())
					return nil
				}
				pterm.Println()
				config.HARFile = harFile
			}
			// let's create a logger first.
			logLevel := pterm.LogLevelWarn
			if debug {
				logLevel = pterm.LogLevelDebug
			}

			// check if we want to validate the HAR file against the OpenAPI spec.
			// but only if we're not in mock mode and there is a spec provided
			if config.HARValidate && !config.MockMode && config.Contract != "" {
				pterm.Printf("🔍 Validating HAR file against OpenAPI specification: %s\n", pterm.LightMagenta(config.Contract))

				// check if whitelist is empty, if so, fail.
				if len(config.HARPathAllowList) == 0 {
					pterm.Error.Println("Cannot validate HAR file against OpenAPI specification, no paths provided to whitelist, use '-j' / '--har-whitelist' to define whitelist paths")
					pterm.Println()
					return nil
				}
				printLoadedHarWhitelist(config.HARPathAllowList)

			} else {
				// we can't use this mode, print an error and return
				if config.HARValidate && config.MockMode {
					pterm.Error.Println("Cannot validate HAR file against OpenAPI specification in mock mode!")
					pterm.Println()
					return nil
				}

				// if there is no spec, print an error
				if config.HARValidate && config.Contract == "" {
					pterm.Error.Println("Cannot validate HAR file against OpenAPI specification, no specification provided, use '-s'")
					pterm.Println()
					return nil
				}

			}

			ptermLog := &pterm.Logger{
				Formatter:  pterm.LogFormatterColorful,
				Writer:     os.Stdout,
				Level:      logLevel,
				ShowTime:   true,
				TimeFormat: "2006-01-02 15:04:05",
				MaxWidth:   180,
				KeyStyles: map[string]pterm.Style{
					"error":  *pterm.NewStyle(pterm.FgRed, pterm.Bold),
					"err":    *pterm.NewStyle(pterm.FgRed, pterm.Bold),
					"caller": *pterm.NewStyle(pterm.FgGray, pterm.Bold),
				},
			}

			handler := pterm.NewSlogHandler(ptermLog)
			config.Logger = slog.New(handler)

			// if we have a HAR file, we need to load it into the config
			if harFile != nil {
				config.HARFile = harFile
			}

			// load the openapi spec
			var doc libopenapi.Document
			var docModel *libopenapi.DocumentModel[v3.Document]
			var err error
			if config.Contract != "" {
				doc, err = loadOpenAPISpec(config.Contract, config.Base)
				if err != nil {
					return err
				}

				// build a model
				var errs []error
				docModel, errs = doc.BuildV3Model()
				if len(errs) > 0 && docModel != nil {
					pterm.Warning.Printf("OpenAPI Specification loaded, but there %s %d %s detected...\n",
						shared.Pluralize(len(errs), "was", "were"),
						len(errs),
						shared.Pluralize(len(errs), "issue", "issues"))
					for _, e := range errs {
						pterm.Warning.Printf("--> %s\n", e.Error())
					}
				}
				if len(errs) > 0 && docModel == nil {
					pterm.Error.Printf("Failed to load / read OpenAPI specification.")
					return errors.Join(errs...)
				}
			}

			if doc != nil {
				pterm.Info.Printf("OpenAPI Specification: '%s' parsed and read\n", config.Contract)
			}

			if !config.HARValidate {

				// ready to boot, let's go!
				_, pErr := runWiretapService(&config, doc)

				if pErr != nil {
					pterm.Println()
					pterm.Error.Printf("Cannot start wiretap: %s\n", pErr.Error())
					pterm.Println()
					return nil
				}
			} else {

				if harFile != nil && config.HARValidate {

					count := 0
					for _, entry := range harFile.Log.Entries {
						if entry.Request.Method != "" {
							count++
						}
						if entry.Response.StatusCode > 0 {
							count++
						}
					}

					validationErrors := har.ValidateHAR(harFile, &docModel.Model, &config)
					if len(validationErrors) > 0 {
						pterm.Println()
						pterm.Error.Printf("HAR file failed validation against OpenAPI specification: %s\n", config.Contract)
						pterm.Println()

						for _, e := range validationErrors {

							location := pterm.Sprintf("Violation location: %s:%d:%d", pterm.LightCyan(config.Contract), e.SpecLine, e.SpecCol)
							var items []pterm.BulletListItem
							items = append(items, pterm.BulletListItem{
								Level: 0, Text: pterm.LightRed(e.Message),
							})
							if e.Reason != e.Message {
								items = append(items, pterm.BulletListItem{
									Level: 1, Text: pterm.Sprintf("Reason: %s", pterm.Gray(e.Reason)),
								})
							}
							if len(e.SchemaValidationErrors) > 0 {

								items = append(items, pterm.BulletListItem{
									Level: 2, Text: pterm.Gray(e.SchemaValidationErrors[0].Reason),
								})
								if e.SchemaValidationErrors[0].Line >= 1 {
									items = append(items, pterm.BulletListItem{
										Level: 3, Text: pterm.Sprintf("Schema violation Location: %s:%d:%d",
											pterm.LightCyan(config.Contract), e.SchemaValidationErrors[0].Line,
											e.SchemaValidationErrors[0].Column),
									})
								}

							}
							if e.SpecLine >= 1 {
								items = append(items, pterm.BulletListItem{
									Level: 1, Text: location,
								})
							}
							pterm.DefaultBulletList.WithItems(items).Render()
						}

						if len(validationErrors) > 0 {
							// render validationErrors to JSON
							b, _ := json.MarshalIndent(validationErrors, "", "  ")
							os.WriteFile(reportFilename, b, 0644)
							pterm.Printf("Report generated and saved to: %s", pterm.LightMagenta(reportFilename))
							pterm.Println()

							pterm.Println()
							pterm.Error.Printf("Wiretap detected %d contract violations against %d requests and responses",
								len(validationErrors), count)
							pterm.Println()

						}

						return nil
					} else {
						pterm.Println()
						pterm.Success.Printf("HAR file passed validation against %d requests and responses", count)
						pterm.Println()

					}

				}

			}

			return nil

		},
	}
)

func Execute(version, commit, date string, fs embed.FS) {
	Version = version
	Commit = commit
	Date = date
	FS = fs

	rootCmd.Flags().StringP("url", "u", "", "Set the redirect URL for wiretap to send traffic to")
	rootCmd.Flags().IntP("delay", "d", 0, "Set a global delay for all API requests")
	rootCmd.Flags().StringP("port", "p", "", "Set port on which to listen for HTTP traffic (default is 9090)")
	rootCmd.Flags().StringP("monitor-port", "m", "", "Set port on which to serve the monitor UI (default is 9091)")
	rootCmd.Flags().StringP("ws-port", "w", "", "Set port on which to serve the monitor UI websocket (default is 9092)")
	rootCmd.Flags().StringP("ws-host", "v", "localhost", "Set the backend hostname for wiretap, for remotely deployed service")
	rootCmd.Flags().StringP("spec", "s", "", "Set the path to the OpenAPI specification to use")
	rootCmd.Flags().StringP("static", "t", "", "Set the path to a directory of static files to serve")
	rootCmd.Flags().StringP("static-index", "i", "index.html", "Set the index filename for static file serving (default is index.html)")
	rootCmd.Flags().StringP("cert", "n", "", "Set the path to the TLS certificate to use for TLS/HTTPS")
	rootCmd.Flags().StringP("key", "k", "", "Set the path to the TLS certificate key to use for TLS/HTTPS")
	rootCmd.Flags().BoolP("hard-validation", "e", false, "Return a HTTP error for non-compliant request/response")
	rootCmd.Flags().IntP("hard-validation-code", "q", 400, "Set a custom http error code for non-compliant requests when using the hard-error flag")
	rootCmd.Flags().IntP("hard-validation-return-code", "y", 502, "Set a custom http error code for non-compliant responses when using the hard-error flag")
	rootCmd.Flags().BoolP("mock-mode", "x", false, "Run in mock mode, responses are mocked and no traffic is sent to the target API (requires OpenAPI spec)")
	rootCmd.Flags().BoolP("enable-all-mock-response-fields", "o", true, "Enable usage of all property examples in mock responses. When set to false, only required field examples will be used.")
	rootCmd.Flags().StringP("config", "c", "", "Location of wiretap configuration file to use (default is .wiretap in current directory)")
	rootCmd.Flags().StringP("base", "b", "", "Set a base path to resolve relative file references from, or a overriding base URL to resolve remote references from")
	rootCmd.Flags().BoolP("debug", "l", false, "Enable debug logging")
	rootCmd.Flags().StringP("har", "z", "", "Load a HAR file instead of sniffing traffic")
	rootCmd.Flags().BoolP("har-validate", "g", false, "Load a HAR file instead of sniffing traffic, and validate against the OpenAPI specification (requires -s)")
	rootCmd.Flags().StringArrayP("har-allow", "j", nil, "Add a path to the HAR allow list, can use arg multiple times")
	rootCmd.Flags().StringP("report-filename", "f", "wiretap-report.json", "Filename for any headless report generation output")
	rootCmd.Flags().BoolP("stream-report", "a", false, "Stream violations to report JSON file as they occur (headless mode)")
	rootCmd.Flags().BoolP("strict-redirect-location", "r", false, "Rewrite the redirect `Location` header on redirect responses to wiretap's API Gateway Host")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func printLoadedIgnorePathRewrite(ignoreRewritePaths []*shared.IgnoreRewriteConfig) {
	pterm.Info.Printf("Loaded %d %s on which to globally ignore rewriting", len(ignoreRewritePaths),
		shared.Pluralize(len(ignoreRewritePaths), "path", "paths"))
	pterm.Println()

	for _, ignoreRewrite := range ignoreRewritePaths {
		pterm.Printf("🙅 Paths matching '%s' will not be re-written regardless of Path Rewrite configuration\n", pterm.LightCyan(ignoreRewrite.Path))
	}
	pterm.Println()
}

func printLoadedPathConfigurations(configs *orderedmap.Map[string, *shared.WiretapPathConfig]) {
	pterm.Info.Printf("Loaded %d path %s:\n", configs.Len(),
		shared.Pluralize(configs.Len(), "configuration", "configurations"))
	pterm.Println()

	for x := configs.First(); x != nil; x = x.Next() {
		k, v := x.Key(), x.Value()
		pterm.Printf("%s --> %s\n", pterm.LightMagenta(k), pterm.LightCyan(v.Target))
		for ka, p := range v.PathRewrite {
			pterm.Printf("✏️  '%s' re-written to '%s'\n", pterm.LightCyan(ka), pterm.LightGreen(p))
		}
		if v.Headers != nil {
			for kb, h := range v.Headers.InjectHeaders {
				pterm.Printf("💉 '%s' injected with '%s'\n", pterm.LightCyan(kb), pterm.LightGreen(h))
			}
			for _, h := range v.Headers.DropHeaders {
				pterm.Printf("🗑️  '%s' is being %s\n", pterm.LightCyan(h), pterm.LightRed("dropped"))
			}
		}
		for _, ignoreRewrite := range v.IgnoreRewrite {
			pterm.Printf("🙅 Paths matching '%s' will not be re-written for this path configuration\n", pterm.LightCyan(ignoreRewrite.Path))
		}

		if v.Auth != "" {
			pterm.Printf("🔒 Basic authentication implemented for '%s'\n", pterm.LightMagenta(k))
		}

		if v.RewriteId != "" {
			pterm.Printf("💳  Identifier '%s' registered for this configuration\n", pterm.LightCyan(v.RewriteId))
		}

		pterm.Println()
	}
}

func printLoadedPathDelayConfigurations(pathDelays map[string]int) {
	pterm.Info.Printf("Loaded %d path %s:\n", len(pathDelays),
		shared.Pluralize(len(pathDelays), "delay", "delays"))

	for k, v := range pathDelays {
		pterm.Printf("⏱️ %sms --> %s\n", pterm.LightCyan(v), pterm.LightMagenta(k))
	}
	pterm.Println()

}

func printLoadedVariables(variables map[string]string) {
	pterm.Info.Printf("Loaded %d %s:\n", len(variables),
		shared.Pluralize(len(variables), "variable", "variables"))
	for k, v := range variables {
		pterm.Printf("📌 Variable '${%s}' points to '%s'\n", pterm.LightCyan(k), pterm.LightGreen(v))
	}
	pterm.Println()
}

func printLoadedHarWhitelist(variables []string) {
	pterm.Println()
	pterm.Info.Printf("Loaded %d %s:\n", len(variables),
		shared.Pluralize(len(variables), "HAR whitelist path", "HAR whitelist paths"))
	for _, v := range variables {
		pterm.Printf("📝 '%s' has been whitelisted for HAR analysis\n", pterm.LightCyan(v))
	}
	pterm.Println()
}

func printLoadedIgnoreRedirectPaths(ignoreRedirects []string) {
	pterm.Info.Printf("Loaded %d %s to ignore for redirects:\n", len(ignoreRedirects),
		shared.Pluralize(len(ignoreRedirects), "path", "paths"))

	for _, x := range ignoreRedirects {
		pterm.Printf("🙈 Paths matching '%s' will be ignored for resolving redirects\n", pterm.LightCyan(x))
	}
	pterm.Println()
}

func printLoadedRedirectAllowList(allowRedirects []string) {
	pterm.Info.Printf("Loaded %d allow listed redirect %s:\n", len(allowRedirects),
		shared.Pluralize(len(allowRedirects), "path", "paths"))

	for _, x := range allowRedirects {
		pterm.Printf("🐵 Paths matching '%s' will always follow redirects, regardless of ignoreRedirect settings\n", pterm.LightCyan(x))
	}
	pterm.Println()
}

func printLoadedWebsockets(websockets map[string]*shared.WiretapWebsocketConfig) {
	pterm.Info.Printf("Loaded %d %s: \n", len(websockets), shared.Pluralize(len(websockets), "websocket", "websockets"))

	for websocket := range websockets {
		pterm.Printf("🔌 Paths prefixed '%s' will be managed as a websocket\n", pterm.LightCyan(websocket))
	}
	pterm.Println()
}

func printLoadedIgnoreValidationPaths(ignoreValidations []string) {
	pterm.Info.Printf("Loaded %d %s to ignore validation:\n", len(ignoreValidations),
		shared.Pluralize(len(ignoreValidations), "path", "paths"))

	for _, x := range ignoreValidations {
		pterm.Printf("⚖️ Paths matching '%s' will not have requests validated\n", pterm.LightCyan(x))
	}
	pterm.Println()
}

func printLoadedValidationAllowList(validationAllowList []string) {
	pterm.Info.Printf("Loaded %d allow listed validation paths %s :\n", len(validationAllowList),
		shared.Pluralize(len(validationAllowList), "path", "paths"))

	for _, x := range validationAllowList {
		pterm.Printf("👮 Paths matching '%s' will always have validation run, regardless of ignoreValidation settings\n", pterm.LightCyan(x))
	}
	pterm.Println()
}

func printLoadedMockModeList(mockModeList []string) {
	pterm.Info.Printf("Loaded %d %s from mock mode list:\n", len(mockModeList),
		shared.Pluralize(len(mockModeList), "path", "paths"))

	for _, x := range mockModeList {
		pterm.Printf("️Ⓜ️  Paths matching '%s' will have all responses %s.\n", x, pterm.LightMagenta("generated as mocks/simulations"))
	}
	pterm.Println()
}

func printLoadedHardErrorList(HardErrorList []string) {
	pterm.Info.Printf("Loaded %d %s from hard validation list:\n", len(HardErrorList),
		shared.Pluralize(len(HardErrorList), "path", "paths"))

	for _, x := range HardErrorList {
		pterm.Printf("️️❌ Paths matching '%s' will create HTTP %s errors for requests and %s errors for responses that fail to pass validation.\n", x, pterm.LightRed("400"), pterm.LightRed("502"))
	}
	pterm.Println()
}
