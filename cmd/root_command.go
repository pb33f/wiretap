// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package cmd

import (
	"embed"
	"github.com/pb33f/wiretap/shared"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net/url"
	"os"
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

			if configFlag == "" {
				pterm.Info.Println("Attempting to locate wiretap configuration...")
				viper.SetConfigFile(".wiretap")
				viper.SetConfigType("env")
				viper.AddConfigPath("$HOME/.wiretap")
				viper.AddConfigPath(".")
			} else {
				viper.SetConfigFile(configFlag)
			}

			cerr := viper.ReadInConfig()
			if cerr != nil && configFlag != "" {
				pterm.Warning.Printf("No wiretap configuration located. Using defaults: %s\n", cerr.Error())
			}
			if cerr != nil && configFlag == "" {
				pterm.Info.Println("No wiretap configuration located. Using defaults.")
			}
			if cerr == nil {
				pterm.Info.Printf("Located configuration file at: %s\n", viper.ConfigFileUsed())
			}

			var spec string
			var port string
			var monitorPort string
			var redirectHost string
			var redirectPort string
			var redirectScheme string
			var redirectBasePath string
			var redirectURL string
			var globalAPIDelay int

			// extract from wiretap environment variables.
			if viper.IsSet("PORT") {
				port = viper.GetString("PORT")
			}

			if viper.IsSet("SPEC") {
				spec = viper.GetString("SPEC")
			}

			if viper.IsSet("MONITOR_PORT") {
				monitorPort = viper.GetString("MONITOR_PORT")
			}

			if viper.IsSet("REDIRECT_URL") {
				redirectURL = viper.GetString("REDIRECT_URL")
			}

			if viper.IsSet("GLOBAL_API_DELAY") {
				globalAPIDelay = viper.GetInt("GLOBAL_API_DELAY")
			}

			portFlag, _ := cmd.Flags().GetString("port")
			if portFlag != "" {
				port = portFlag
			} else {
				if port == "" {
					port = "9090" // default
				}
			}

			specFlag, _ := cmd.Flags().GetString("spec")
			if specFlag != "" {
				spec = specFlag
			}

			monitorPortFlag, _ := cmd.Flags().GetString("monitor-port")
			if monitorPortFlag != "" {
				monitorPort = monitorPortFlag
			} else {
				if monitorPort == "" {
					monitorPort = "9091" // default
				}
			}

			redirectURLFlag, _ := cmd.Flags().GetString("url")
			if redirectURLFlag != "" {
				redirectURL = redirectURLFlag
			}

			globalAPIDelayFlag, _ := cmd.Flags().GetInt("delay")
			if globalAPIDelayFlag > 0 {
				globalAPIDelay = globalAPIDelayFlag
			}

			if spec == "" {
				pterm.Error.Println("No OpenAPI specification provided. " +
					"Please provide a path to an OpenAPI specification using the --spec or -s flags.")
				pterm.Println()
				return nil
			}

			if redirectURL == "" {
				pterm.Error.Println("No redirect URL provided. " +
					"Please provide a URL to redirect API traffic to using the --url or -u flags.")
				pterm.Println()
				return nil
			}

			if redirectURL != "" {
				parsedURL, e := url.Parse(redirectURL)
				if e != nil {
					pterm.Error.Printf("URL is not valid. "+
						"Please provide a valid URL to redirect to. %s cannot be parsed\n\n", redirectURL)
					return nil
				}
				if parsedURL.Scheme == "" || parsedURL.Host == "" {
					pterm.Error.Printf("URL is not valid. "+
						"Please provide a valid URL to redirect to. %s cannot be parsed\n\n", redirectURL)
					return nil
				}
				redirectHost = parsedURL.Host
				redirectPort = parsedURL.Port()
				redirectScheme = parsedURL.Scheme
				redirectBasePath = parsedURL.Path
			}

			config := shared.WiretapConfiguration{
				Contract:         spec,
				RedirectHost:     redirectHost,
				RedirectBasePath: redirectBasePath,
				RedirectPort:     redirectPort,
				RedirectProtocol: redirectScheme,
				Port:             port,
				MonitorPort:      monitorPort,
				GlobalAPIDelay:   globalAPIDelay,
				FS:               FS,
			}

			// ready to boot, let's go!
			_, _ = runWiretapService(&config)

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
	rootCmd.Flags().StringP("port", "p", "", "Set port on which to listen for API traffic")
	rootCmd.Flags().StringP("monitor-port", "m", "", "Set post on which to serve the monitor UI")
	rootCmd.Flags().StringP("spec", "s", "", "Set the path to the OpenAPI specification to use")
	rootCmd.Flags().StringP("config", "c", "",
		"Location of wiretap configuration file to use (default is .wiretap in current directory)")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
