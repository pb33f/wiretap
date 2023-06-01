// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package cmd

import (
	"github.com/pb33f/ranch/bus"
	"github.com/pb33f/ranch/model"
	"github.com/pb33f/ranch/plank/pkg/server"
	"github.com/pb33f/wiretap/shared"
	"github.com/pterm/pterm"
)

func bootedMessage(wiretapConfig *shared.WiretapConfiguration) {
	// print a nice message to the user when ranch is online.
	go func() {
		handler, _ := bus.GetBus().ListenStream(server.RANCH_SERVER_ONLINE_CHANNEL)
		seen := false
		handler.Handle(func(message *model.Message) {
			if !seen {
				seen = true
				pterm.Println()
				pterm.Info.Println("Wiretap Service is ready.")
				pterm.Println()
				pterm.Info.Printf("API Gateway: http://localhost:%s\n", wiretapConfig.Port)
				pterm.Info.Printf("Monitor: http://localhost:%s\n", wiretapConfig.MonitorPort)
				pterm.Println()

			}
		}, nil)
	}()
}
