// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package cmd

import (
	"embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/pb33f/wiretap/har"
	"github.com/pb33f/wiretap/shared"
	wiretapSpecs "github.com/pb33f/wiretap/specs"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v4"
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

			flags := cmd.Flags()
			configFlag, _ := flags.GetString("config")

			specs := make([]string, 0)
			specDirs := make([]string, 0)
			specIgnore := make([]string, 0)
			var primarySpec string
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

			// static mock dir
			var staticMockDir string

			// mock mode
			var mockMode bool
			var useAllMockResponseFields bool

			certFlag, _ := flags.GetString("cert")
			if certFlag != "" {
				cert = certFlag
			}

			keyFlag, _ := flags.GetString("key")
			if keyFlag != "" {
				certKey = keyFlag
			}
			base, _ := flags.GetString("base")
			reportFilename, _ := flags.GetString("report-filename")
			reportFilenameChanged := flags.Changed("report-filename")

			harFlag, _ := flags.GetString("har")
			harValidate, _ := flags.GetBool("har-validate")
			harWhiteList, _ := flags.GetStringArray("har-allow")
			harReplayDelay, _ := flags.GetInt("har-replay-delay")

			debug, _ := flags.GetBool("debug")
			staticMockDir, _ = flags.GetString("static-mock-dir")
			mockMode, _ = flags.GetBool("mock-mode")
			mockBypassValidation, _ := flags.GetBool("mock-bypass-validation")
			useAllMockResponseFields, _ = flags.GetBool("enable-all-mock-response-fields")
			hardError, _ = flags.GetBool("hard-validation")
			hardErrorCode, _ = flags.GetInt("hard-validation-code")
			hardErrorReturnCode, _ = flags.GetInt("hard-validation-return-code")
			hardErrorReturnProblem, _ := flags.GetBool("hard-error-return-problem")
			streamReport, _ := flags.GetBool("stream-report")
			strictRedirectLocation, _ := flags.GetBool("strict-redirect-location")
			strictMode, _ := flags.GetBool("strict-mode")
			dryRunFlag, _ := flags.GetBool("dry-run")
			ignoreClashingOperationIDFlag, _ := flags.GetBool("ignore-clashing-operationid")

			portFlag, _ := flags.GetString("port")
			if portFlag != "" {
				port = portFlag
			} else {
				port = "9090" // default
			}

			specFlag, _ := flags.GetString("spec")
			if len(specFlag) != 0 && specFlag != "" {
				specs = append(specs, specFlag)
				if primarySpec == "" {
					primarySpec = specFlag
				}
			}

			specsFlag, _ := flags.GetStringSlice("specs")
			if len(specsFlag) > 0 && specsFlag[0] != "" {
				specs = append(specs, specsFlag...)
				if primarySpec == "" {
					primarySpec = specsFlag[0]
				}
			}
			specDirsFlag, _ := flags.GetStringSlice("spec-dir")
			if len(specDirsFlag) > 0 && specDirsFlag[0] != "" {
				specDirs = append(specDirs, specDirsFlag...)
			}
			specIgnoreFlag, _ := flags.GetStringSlice("ignore")
			if len(specIgnoreFlag) > 0 && specIgnoreFlag[0] != "" {
				specIgnore = append(specIgnore, specIgnoreFlag...)
			}

			monitorPortFlag, _ := flags.GetString("monitor-port")
			if monitorPortFlag != "" {
				monitorPort = monitorPortFlag
			} else {
				monitorPort = "9091" // default
			}

			staticDirFlag, _ := flags.GetString("static")
			if staticDirFlag != "" {
				staticDir = staticDirFlag
			}

			staticIndex, _ = flags.GetString("static-index")

			wsHostFlag, _ := flags.GetString("ws-host")
			if wsHostFlag != "" {
				wsHost = wsHostFlag
			} else {
				wsHost = "localhost"
			}

			wsPortFlag, _ := flags.GetString("ws-port")
			if wsPortFlag != "" {
				wsPort = wsPortFlag
			} else {
				wsPort = "9092" // default
			}

			redirectURLFlag, _ := flags.GetString("url")
			if redirectURLFlag != "" {
				redirectURL = redirectURLFlag
			}

			globalAPIDelayFlag, _ := flags.GetInt("delay")
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
				if len(config.Specs) != 0 {
					for _, spec := range config.Specs {
						if spec != "" {
							specs = append(specs, spec)
						}
					}
				}
				if config.Spec != "" {
					specs = append(specs, config.Spec)
					if primarySpec == "" {
						primarySpec = config.Spec
					}
				}
				if len(config.SpecDirs) > 0 {
					specDirs = append(config.SpecDirs, specDirs...)
				}
				if len(config.SpecIgnore) > 0 {
					specIgnore = append(config.SpecIgnore, specIgnore...)
				}
				if dryRunFlag {
					config.DryRun = true
				}
				if ignoreClashingOperationIDFlag {
					config.IgnoreClashingOperationID = true
				}
				if len(staticMockDir) != 0 {
					if len(config.StaticMockDir) == 0 {
						config.StaticMockDir = staticMockDir
					}
				}
				if mockMode {
					if !config.MockMode {
						config.MockMode = true
					}
				}
				if mockBypassValidation {
					if !config.MockBypassValidation {
						config.MockBypassValidation = true
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
				if strictMode {
					if !config.StrictMode {
						config.StrictMode = true
					}
				}

				if reportFilenameChanged {
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
				if harReplayDelay > 0 {
					config.HARReplayDelay = harReplayDelay
				}

			} else {

				pterm.Info.Println("No wiretap configuration located. Using defaults")
				config.StaticIndex = staticIndex
				if len(staticMockDir) != 0 {
					if len(config.StaticMockDir) == 0 {
						config.StaticMockDir = staticMockDir
					}
				}
				if mockMode {
					config.MockMode = true
				}
				if mockBypassValidation {
					config.MockBypassValidation = true
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
				if strictMode {
					config.StrictMode = true
				}
				if base != "" {
					config.Base = base
				}
				config.ReportFile = reportFilename
				config.HAR = harFlag
				config.HARValidate = harValidate
				config.HARPathAllowList = harWhiteList
				config.HARReplayDelay = harReplayDelay
				config.SpecDirs = specDirs
				config.SpecIgnore = specIgnore
				config.DryRun = dryRunFlag
				config.IgnoreClashingOperationID = ignoreClashingOperationIDFlag
			}

			config.SpecDirs = specDirs
			config.SpecIgnore = specIgnore
			dryRun := config.DryRun || dryRunFlag

			discoveredSpecs, discoveryErr := wiretapSpecs.DiscoverSpecs(specs, specDirs, specIgnore)
			if discoveryErr != nil {
				pterm.Error.Printf("Failed to discover OpenAPI specifications: %s\n", discoveryErr.Error())
				return discoveryErr
			}
			specs = discoveredSpecs
			primarySpec, discoveryErr = resolvePrimarySpec(primarySpec, specs, specIgnore)
			if discoveryErr != nil {
				pterm.Error.Printf("Failed to resolve primary OpenAPI specification: %s\n", discoveryErr.Error())
				return discoveryErr
			}

			// If no primary specification has been provided, then we'll default to the first specification in the list
			// Priority for primary specification is (in order):
			// 1. -spec option on the CLI
			// 2. First -specs option on the CLI if defined
			// 3. `contract` key in the yaml config
			// 4. First specification in the list final specification list
			if primarySpec == "" && len(specs) != 0 {
				primarySpec = specs[0]
			}

			if len(specs) == 0 {
				pterm.Println()
				pterm.Warning.Println("No OpenAPI specification provided. " +
					"Please provide a path to at least one OpenAPI specification using the --spec or -s flags. \n" +
					"Without an OpenAPI specification, wiretap will not be able to validate " +
					"requests and responses")
				pterm.Println()
			}

			wantsMockMode := config.MockMode || mockMode || len(config.MockModeList) > 0
			if !dryRun && wantsMockMode && len(specs) == 0 {
				pterm.Println()
				pterm.Error.Println("Cannot enable mock mode, no OpenAPI specification provided!\n" +
					"Please provide a path to an OpenAPI specification using the --spec or -s flags.\n" +
					"Without an OpenAPI specification, wiretap will not be able to generate mock responses")
				pterm.Println()
				return fmt.Errorf("cannot enable mock mode: no OpenAPI specification provided")
			}

			if !dryRun && !config.MockMode && redirectURL == "" && config.HAR == "" && !config.HARValidate {
				pterm.Println()
				pterm.Error.Println("No redirect URL provided. " +
					"Please provide a URL to redirect API traffic to using the --url or -u flags.")
				pterm.Println()
				return nil
			}

			if !dryRun && redirectURL != "" {

				parsedURL, e := url.Parse(redirectURL)
				if e != nil {
					pterm.Println()
					pterm.Error.Printf("URL is not valid. "+
						"Please provide a valid URL to redirect to. %s cannot be parsed\n\n", redirectURL)
					pterm.Println()
					return fmt.Errorf("invalid redirect URL %q: %w", redirectURL, e)
				}
				if parsedURL.Scheme == "" || parsedURL.Host == "" {
					pterm.Println()
					pterm.Error.Printf("URL is not valid. "+
						"Please provide a valid URL to redirect to. %s cannot be parsed\n\n", redirectURL)
					pterm.Println()
					return fmt.Errorf("invalid redirect URL %q: missing scheme or host", redirectURL)
				}
				redirectHost = parsedURL.Hostname()
				redirectPort = parsedURL.Port()
				redirectScheme = parsedURL.Scheme
				redirectBasePath = parsedURL.Path
			}

			if len(specs) != 0 {
				config.Contracts = specs
				config.PrimaryContract = primarySpec
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
			if config.ReportFile == "" {
				config.ReportFile = reportFilename
			}
			reportFilename = config.ReportFile
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
			if !config.HardErrorReturnProblem && hardErrorReturnProblem {
				config.HardErrorReturnProblem = true
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
			hasPathConfigurations := config.PathConfigurations != nil && config.PathConfigurations.Len() > 0
			if hasPathConfigurations || len(config.StaticPaths) > 0 || len(config.HARPathAllowList) > 0 || len(config.IgnorePathRewrite) > 0 {
				config.CompilePaths()
				if len(config.IgnorePathRewrite) > 0 {
					printLoadedIgnorePathRewrite(config.IgnorePathRewrite)
				}

				if hasPathConfigurations {
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

			// static mock dir
			if len(config.StaticMockDir) != 0 {
				pterm.Printf("Ⓜ️ %s. Requests matching mock definitions in the static-mock-dir will return mocked responses.\n",
					pterm.LightCyan("Static mock directory defined"))
				pterm.Println()
			}

			// mock mode
			if config.MockMode {
				pterm.Printf("Ⓜ️ %s. All responses will be mocked and no traffic will be sent to the target API.\n",
					pterm.LightCyan("Mock mode enabled"))
				pterm.Println()
			}

			// strict mode
			if config.StrictMode {
				pterm.Printf("🔬 %s. Undeclared properties, parameters, headers, and cookies will be reported as validation errors.\n",
					pterm.LightCyan("Strict validation mode enabled"))
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

			// check if we're using a HAR file
			if !dryRun && config.HAR != "" {
				pterm.Println()
				pterm.Printf("📦 Loading HAR file: %s\n", pterm.LightMagenta(config.HAR))
				info, err := os.Stat(config.HAR)
				if err != nil {
					pterm.Error.Printf("Cannot read HAR file: %s (%s)\n", config.HAR, err.Error())
					return fmt.Errorf("cannot read HAR file %q: %w", config.HAR, err)
				}
				if info.IsDir() {
					pterm.Error.Printf("Cannot read HAR file: %s is a directory\n", config.HAR)
					return fmt.Errorf("cannot read HAR file %q: is a directory", config.HAR)
				}
				pterm.Println()
			}
			// let's create a logger first.
			logLevel := pterm.LogLevelWarn
			if debug {
				logLevel = pterm.LogLevelDebug
			}

			// check if we want to validate the HAR file against the OpenAPI spec.
			if !dryRun && config.HARValidate {
				if config.MockMode {
					pterm.Error.Println("Cannot validate HAR file against OpenAPI specification in mock mode!")
					pterm.Println()
					return fmt.Errorf("cannot validate HAR file against OpenAPI specification in mock mode")
				}

				if len(config.Contracts) == 0 {
					pterm.Error.Println("Cannot validate HAR file against OpenAPI specification, no specification provided, use '-s'")
					pterm.Println()
					return fmt.Errorf("cannot validate HAR file against OpenAPI specification: no specification provided")
				}

				if config.HAR == "" {
					pterm.Error.Println("Cannot validate HAR file against OpenAPI specification, no HAR file provided, use '-z' / '--har'")
					pterm.Println()
					return fmt.Errorf("cannot validate HAR file against OpenAPI specification: no HAR file provided")
				}

				pterm.Printf("🔍 Validating HAR file against OpenAPI specification(s): %s\n", pterm.LightMagenta(config.GetContractList()))

				// check if whitelist is empty, if so, fail.
				if len(config.HARPathAllowList) == 0 {
					pterm.Error.Println("Cannot validate HAR file against OpenAPI specification, no paths provided to allow list, use '-j' / '--har-allow' to define allow-list paths")
					pterm.Println()
					return fmt.Errorf("cannot validate HAR file against OpenAPI specification: no HAR allow-list paths provided")
				}
				printLoadedHarWhitelist(config.HARPathAllowList)
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

			// load the openapi specs and analyze conflicts
			var primaryDoc libopenapi.Document
			docs, loadErrors := loadAllSpecs(config.Contracts, config.Base)
			docModels := make([]shared.ApiDocumentModel, 0, len(docs))
			for _, doc := range docs {
				docModels = append(docModels, shared.ApiDocumentModel{
					DocumentName:  doc.DocumentName,
					DocumentModel: doc.DocumentModel,
				})
				if doc.DocumentName == config.PrimaryContract {
					primaryDoc = doc.Document
				}
			}
			if primaryDoc == nil && len(docs) > 0 {
				primaryDoc = docs[0].Document
			}
			conflictReport := wiretapSpecs.Analyze(docs, wiretapSpecs.AnalyzeOptions{
				IgnoreClashingOperationID: config.IgnoreClashingOperationID,
			})
			conflictReport.LoadErrors = loadErrors
			conflictReport.SpecCount += len(loadErrors)
			wiretapSpecs.RenderConsole(conflictReport, os.Stdout)

			if dryRun {
				if len(conflictReport.Conflicts)+len(conflictReport.LoadErrors) > 0 {
					return fmt.Errorf("dry run failed: detected %d conflicts and %d load errors",
						len(conflictReport.Conflicts), len(conflictReport.LoadErrors))
				}
				return nil
			}

			if len(docs) != 0 {
				pterm.Info.Printf("OpenAPI Specification(s): '%s' parsed and read\n", config.GetContractList())
				pterm.Info.Printf("Primary OpenAPI Specification: '%s'\n", config.PrimaryContract)
			}

			if len(conflictReport.LoadErrors) > 0 {
				firstErr := conflictReport.LoadErrors[0]
				return fmt.Errorf("failed to load %d OpenAPI specification(s); first error in %s: %w",
					len(conflictReport.LoadErrors), firstErr.Spec, firstErr.Error)
			}

			if !config.HARValidate {

				// ready to boot, let's go!
				_, pErr := runWiretapService(&config, docs, primaryDoc, conflictReport)

				if pErr != nil {
					pterm.Println()
					pterm.Error.Printf("Cannot start wiretap: %s\n", pErr.Error())
					pterm.Println()
					return fmt.Errorf("cannot start wiretap: %w", pErr)
				}
			} else {

				if config.HAR != "" && config.HARValidate {
					result := har.ValidateHARWithResult(config.HAR, docModels, &config)
					count := result.MessageCount
					validationErrors := result.Errors
					if result.Err != nil {
						pterm.Println()
						pterm.Error.Printf("HAR file could not be validated: %s\n", result.Err.Error())
						pterm.Println()
						return fmt.Errorf("har file could not be validated: %w", result.Err)
					}
					if len(validationErrors) > 0 {
						pterm.Println()
						pterm.Error.Printf("HAR file failed validation against OpenAPI specification(s): %s\n", config.GetContractList())
						pterm.Println()

						for _, e := range validationErrors {

							location := pterm.Sprintf("Violation location: %s:%d:%d", pterm.LightCyan(e.SpecName), e.SpecLine, e.SpecCol)
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
											pterm.LightCyan(e.SpecName), e.SchemaValidationErrors[0].Line,
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

						return fmt.Errorf("har file failed validation: detected %d contract violations against %d requests and responses",
							len(validationErrors), count)
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

	registerRootFlags(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func resolvePrimarySpec(primarySpec string, discoveredSpecs []string, ignorePatterns []string) (string, error) {
	if primarySpec == "" {
		if len(discoveredSpecs) == 0 {
			return "", nil
		}
		return discoveredSpecs[0], nil
	}

	resolved, err := wiretapSpecs.DiscoverSpecs([]string{primarySpec}, nil, ignorePatterns)
	if err != nil {
		return "", err
	}
	if len(resolved) == 0 {
		return primarySpec, nil
	}
	return resolved[0], nil
}

func registerRootFlags(cmd *cobra.Command) {
	flags := cmd.Flags()
	if flags.Lookup("url") != nil {
		return
	}

	flags.StringP("url", "u", "", "Set the redirect URL for wiretap to send traffic to")
	flags.IntP("delay", "d", 0, "Set a global delay for all API requests")
	flags.StringP("port", "p", "", "Set port on which to listen for HTTP traffic (default is 9090)")
	flags.StringP("monitor-port", "m", "", "Set port on which to serve the monitor UI (default is 9091)")
	flags.StringP("ws-port", "w", "", "Set port on which to serve the monitor UI websocket (default is 9092)")
	flags.StringP("ws-host", "v", "localhost", "Set the backend hostname for wiretap, for remotely deployed service")
	flags.StringP("spec", "s", "", "List of paths to the OpenAPI specification to use")
	flags.StringSliceP("specs", "S", []string{}, "List of paths to the OpenAPI specification to use")
	flags.StringSlice("spec-dir", []string{}, "Directory roots to recursively scan for OpenAPI specifications")
	flags.StringSlice("ignore", []string{}, "Glob patterns to ignore while discovering OpenAPI specifications")
	flags.Bool("dry-run", false, "Discover and analyze OpenAPI specifications, print the report, then exit")
	flags.Bool("ignore-clashing-operationid", false, "Ignore duplicate operationId conflicts during multi-spec analysis")
	flags.StringP("static", "t", "", "Set the path to a directory of static files to serve")
	flags.StringP("static-index", "i", "index.html", "Set the index filename for static file serving (default is index.html)")
	flags.StringP("cert", "n", "", "Set the path to the TLS certificate to use for TLS/HTTPS")
	flags.StringP("key", "k", "", "Set the path to the TLS certificate key to use for TLS/HTTPS")
	flags.BoolP("hard-validation", "e", false, "Return a HTTP error for non-compliant request/response")
	flags.IntP("hard-validation-code", "q", 400, "Set a custom http error code for non-compliant requests when using the hard-error flag")
	flags.IntP("hard-validation-return-code", "y", 502, "Set a custom http error code for non-compliant responses when using the hard-error flag")
	flags.Bool("hard-error-return-problem", false, "When hard-validation triggers, return an RFC 9457 application/problem+json body describing the validation failures (default is false)")
	flags.StringP("static-mock-dir", "", "", "Directory containing static mock definitions. All requests matching these definitions will return mocked responses.")
	flags.BoolP("mock-mode", "x", false, "Run in mock mode, responses are mocked and no traffic is sent to the target API (requires OpenAPI spec)")
	flags.Bool("mock-bypass-validation", false, "In mock mode, bypass request validation so Preferred / wiretap-status-code examples are returned even for malformed requests (default is false)")
	flags.BoolP("enable-all-mock-response-fields", "o", true, "Enable usage of all property examples in mock responses. When set to false, only required field examples will be used.")
	flags.StringP("config", "c", "", "Location of wiretap configuration file to use (default is .wiretap in current directory)")
	flags.StringP("base", "b", "", "Set a base path to resolve relative file references from, or a overriding base URL to resolve remote references from")
	flags.BoolP("debug", "l", false, "Enable debug logging")
	flags.StringP("har", "z", "", "Load a HAR file instead of sniffing traffic")
	flags.BoolP("har-validate", "g", false, "Load a HAR file instead of sniffing traffic, and validate against the OpenAPI specification (requires -s)")
	flags.StringArrayP("har-allow", "j", nil, "Add a path to the HAR allow list, can use arg multiple times")
	flags.Int("har-replay-delay", 0, "Delay in milliseconds between HAR replayed request and response events (default 10ms)")
	flags.StringP("report-filename", "f", "wiretap-report.jsonl", "Filename for any headless report generation output")
	flags.BoolP("stream-report", "a", false, "Stream violations to report JSON file as they occur (headless mode)")
	flags.BoolP("strict-redirect-location", "r", false, "Rewrite the redirect `Location` header on redirect responses to wiretap's API Gateway Host")
	flags.Bool("strict-mode", false, "Enable strict validation to detect undeclared properties, parameters, headers, and cookies")
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
