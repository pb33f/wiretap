// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"
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
				b1 := pterm.DefaultBox.WithTitle(pterm.LightMagenta("API Gateway")).Sprint(fmt.Sprintf("http://localhost:%s", wiretapConfig.Port))
				b2 := pterm.DefaultBox.WithTitle(pterm.LightMagenta("Monitor UI")).Sprint(fmt.Sprintf("http://localhost:%s", wiretapConfig.MonitorPort))
				panels, _ := pterm.DefaultPanel.WithPanels(pterm.Panels{
					{{Data: b1}, {Data: b2}},
				}).Srender()

				pterm.DefaultBox.WithTitle(pterm.LightCyan("wiretap is online!")).
					WithTitleTopLeft().
					WithRightPadding(3).
					WithTopPadding(1).
					WithLeftPadding(3).
					WithBottomPadding(0).
					Println(panels)

				pterm.Println()
				pterm.Info.Printf("API Gateway is now proxying all traffic to '%s'\n",
					pterm.LightMagenta(wiretapConfig.RedirectURL))
				pterm.Println()
			}
		}, nil)
	}()
}
