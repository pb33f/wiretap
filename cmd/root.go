// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package cmd

import (
    "github.com/pb33f/wiretap/daemon"
    "github.com/pterm/pterm"
    "github.com/spf13/cobra"
    "os"
)

var (
    Version string
    Commit  string
    Date    string

    rootCmd = &cobra.Command{
        SilenceUsage:  true,
        SilenceErrors: true,
        Use:           "wiretap",
        Short:         "wiretap is a tool for detecting API compliance against an OpenAPI contract, by sniffing network traffic.",
        Long:          `wiretap is a tool for detecting API compliance against an OpenAPI contract, by sniffing network traffic.`,
        RunE: func(cmd *cobra.Command, args []string) error {

            PrintBanner()

            if len(args) == 0 {
                pterm.Error.Println("Supply at least a single argument, pointing to an OpenAPI file to load...")
                return nil
            }

            file := args[0]
            host, _ := cmd.Flags().GetString("server")
            // get port from environment.
            port := os.Getenv("PORT")
            if port == "" {
                port = "9090" // default.
            }

            mport := os.Getenv("MONITOR_PORT")
            if mport == "" {
                mport = "9091" // default.
            }

            config := daemon.WiretapServiceConfiguration{
                Contract:     file,
                RedirectHost: host,
                Port:         port,
                MonitorPort:  mport,
            }

            _, _ = runWiretapService(&config)

            return nil
        },
    }
)

func Execute(version, commit, date string) {
    Version = version
    Commit = commit
    Date = date
    rootCmd.PersistentFlags().StringP("server", "s", "", "override the host in the OpenAPI specification")

    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}
