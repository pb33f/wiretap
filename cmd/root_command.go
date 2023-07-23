// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package cmd

import (
	"embed"
	"gopkg.in/yaml.v3"
	"net/url"
	"os"
	"path/filepath"

	"github.com/pb33f/wiretap/shared"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var (
	Version string
	Commit  string
	Date    string
	FS      embed.FS

	rootCmd = &cobra.Command{
		SilenceUsage:  true,
		SilenceErrors: true,
		Use:           "wiretap",
		Short:         "wiretap is a tool for detecting API compliance against an OpenAPI contract, by sniffing network traffic.",
		Long:          `wiretap is a tool for detecting API compliance against an OpenAPI contract, by sniffing network traffic.`,
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
			} else {

				pterm.Info.Println("No wiretap configuration located. Using defaults")
				config.StaticIndex = staticIndex
			}

			if spec == "" {
				pterm.Println()
				pterm.Warning.Println("No OpenAPI specification provided. " +
					"Please provide a path to an OpenAPI specification using the --spec or -s flags.")
				pterm.Warning.Println("Without an OpenAPI specification, wiretap will not be able to validate " +
					"requests and responses")
				pterm.Println()
			}

			if redirectURL == "" {
				pterm.Println()
				pterm.Error.Println("No redirect URL provided. " +
					"Please provide a URL to redirect API traffic to using the --url or -u flags.")
				pterm.Println()
				return nil
			}

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
			if config.GlobalAPIDelay == 0 {
				config.GlobalAPIDelay = globalAPIDelay
			}
			if config.StaticDir == "" {
				config.StaticDir = staticDir
			}
			config.FS = FS

			// variables
			if len(config.Variables) > 0 {
				config.CompileVariables()
				printLoadedVariables(config.Variables)
			}

			// paths
			if len(config.PathConfigurations) > 0 {
				config.CompilePaths()
				printLoadedPathConfigurations(config.PathConfigurations)
			}

			if config.Headers != nil && len(config.Headers.DropHeaders) > 0 {
				pterm.Info.Printf("Dropping the following %d %s globally:\n", len(config.Headers.DropHeaders),
					shared.Pluralize(len(config.Headers.DropHeaders), "header", "headers"))
				for _, header := range config.Headers.DropHeaders {
					pterm.Printf("🗑️  %s\n", pterm.LightRed(header))
				}
				pterm.Println()
			}

			if len(config.StaticPaths) > 0 && config.StaticDir != "" {
				staticPath := filepath.Join(config.StaticDir, config.StaticIndex)
				pterm.Info.Printf("Mapping %d static %s to '%s':\n", len(config.StaticPaths),
					shared.Pluralize(len(config.StaticPaths), "path", "paths"), staticPath)
				for _, path := range config.StaticPaths {
					pterm.Printf("⛱️  %s\n", pterm.LightMagenta(path))
				}
				pterm.Println()
			}

			// ready to boot, let's go!
			_, pErr := runWiretapService(&config)

			if pErr != nil {
				pterm.Println()
				pterm.Error.Printf("Cannot start wiretap: %s\n", pErr.Error())
				pterm.Println()
				return nil
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
	rootCmd.Flags().StringP("spec", "s", "", "Set the path to the OpenAPI specification to use")
	rootCmd.Flags().StringP("static", "t", "", "Set the path to a directory of static files to serve")
	rootCmd.Flags().StringP("static-index", "i", "index.html", "Set the index filename for static file serving (default is index.html)")

	rootCmd.Flags().StringP("config", "c", "",
		"Location of wiretap configuration file to use (default is .wiretap in current directory)")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func printLoadedPathConfigurations(configs map[string]*shared.WiretapPathConfig) {
	pterm.Info.Printf("Loaded %d path %s:\n", len(configs),
		shared.Pluralize(len(configs), "configuration", "configurations"))
	pterm.Println()

	for k, v := range configs {
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
		if v.Auth != "" {
			pterm.Printf("🔐 Basic authentication implemented for '%s'\n", pterm.LightMagenta(k))
		}
		pterm.Println()
	}
}

func printLoadedVariables(variables map[string]string) {
	pterm.Info.Printf("Loaded %d %s:\n", len(variables),
		shared.Pluralize(len(variables), "variable", "variables"))
	for k, v := range variables {
		pterm.Printf("📌 Variable '${%s}' points to '%s'\n", pterm.LightCyan(k), pterm.LightGreen(v))
	}
	pterm.Println()
}
