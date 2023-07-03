// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package cmd

import (
	"embed"
	"github.com/mitchellh/mapstructure"
	"net/url"
	"os"

	"github.com/pb33f/wiretap/shared"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
				pterm.Error.Printf("No wiretap configuration located. Using defaults: %s\n", cerr.Error())
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
			var wsPort string
			var staticPort string
			var staticDir string
			var pathConfigurations map[string]*shared.WiretapPathConfig
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

			if viper.IsSet("WEBSOCKET_PORT") {
				wsPort = viper.GetString("WEBSOCKET_PORT")
			}

			if viper.IsSet("STATIC_PORT") {
				staticPort = viper.GetString("STATIC_PORT")
			}

			if viper.IsSet("STATIC_DIR") {
				staticDir = viper.GetString("STATIC_DIR")
			}

			if viper.IsSet("PATHS") {
				paths := viper.Get("PATHS")
				var pc map[string]*shared.WiretapPathConfig
				err := mapstructure.Decode(paths, &pc)
				if err != nil {
					pterm.Error.Printf("Unable to decode paths from configuration: %s\n", err.Error())
				} else {
					// print out the path configurations.
					printLoadedPathConfigurations(pc)
					pathConfigurations = pc
				}
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

			staticPortFlag, _ := cmd.Flags().GetString("static-port")
			if staticPortFlag != "" {
				staticPort = staticPortFlag
			} else {
				if staticPort == "" {
					staticPort = "9093" // default
				}
			}

			staticDirFlag, _ := cmd.Flags().GetString("static")
			if staticDirFlag != "" {
				staticDir = staticDirFlag
			}

			wsPortFlag, _ := cmd.Flags().GetString("ws-port")
			if wsPortFlag != "" {
				wsPort = wsPortFlag
			} else {
				if wsPort == "" {
					wsPort = "9092" // default
				}
			}

			redirectURLFlag, _ := cmd.Flags().GetString("url")
			if redirectURLFlag != "" {

				if pathConfigurations != nil {
					// warn the user that the path configurations will trump the switch
					pterm.Warning.Println("Using the --url flag will be *overridden* by the path configuration 'target' setting")
				}
				redirectURL = redirectURLFlag
			}

			globalAPIDelayFlag, _ := cmd.Flags().GetInt("delay")
			if globalAPIDelayFlag > 0 {
				globalAPIDelay = globalAPIDelayFlag
			}

			if spec == "" {
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

			config := shared.WiretapConfiguration{
				Contract:           spec,
				RedirectURL:        redirectURL,
				RedirectHost:       redirectHost,
				RedirectBasePath:   redirectBasePath,
				RedirectPort:       redirectPort,
				RedirectProtocol:   redirectScheme,
				Port:               port,
				MonitorPort:        monitorPort,
				GlobalAPIDelay:     globalAPIDelay,
				WebSocketPort:      wsPort,
				StaticDir:          staticDir,
				StaticPort:         staticPort,
				PathConfigurations: pathConfigurations,
				FS:                 FS,
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
	rootCmd.Flags().StringP("port", "p", "", "Set port on which to listen for API traffic (default is 9090)")
	rootCmd.Flags().StringP("monitor-port", "m", "", "Set port on which to serve the monitor UI (default is 9091)")
	rootCmd.Flags().StringP("ws-port", "w", "", "Set port on which to serve the monitor UI websocket (default is 9092)")
	rootCmd.Flags().StringP("spec", "s", "", "Set the path to the OpenAPI specification to use")
	rootCmd.Flags().StringP("static", "t", "", "Set the path to a directory of static files to serve")
	rootCmd.Flags().StringP("static-port", "r", "", "Set port on which to listen for API traffic (default is 9093)")
	rootCmd.Flags().StringP("config", "c", "",
		"Location of wiretap configuration file to use (default is .wiretap in current directory)")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func printLoadedPathConfigurations(configs map[string]*shared.WiretapPathConfig) {
	plural := func(count int) string {
		if count == 1 {
			return ""
		}
		return "s"
	}

	pterm.Info.Printf("Loaded %d path configuration%s:\n", len(configs), plural(len(configs)))
	pterm.Println()

	for k, v := range configs {
		pterm.Printf("%s\n", pterm.LightMagenta(k))
		for k, p := range v.PathRewrite {
			pterm.Printf("✏️ '%s' re-written to '%s'\n", pterm.LightCyan(k), pterm.LightGreen(p))
		}
		pterm.Println()
	}
}
