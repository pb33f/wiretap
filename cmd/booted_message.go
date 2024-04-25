// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

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

				b1 := pterm.DefaultBox.WithTitle(pterm.LightMagenta("API Gateway")).Sprint(wiretapConfig.GetApiGateway())
				b2 := pterm.DefaultBox.WithTitle(pterm.LightMagenta("Monitor UI")).Sprint(wiretapConfig.GetMonitorUI())
				b3 := pterm.DefaultBox.WithTitle(pterm.LightMagenta("Static files served from")).Sprint(wiretapConfig.StaticDir)

				var pp *pterm.PanelPrinter
				if wiretapConfig.StaticDir != "" {
					pp = pterm.DefaultPanel.WithPanels(pterm.Panels{
						{{Data: b1}, {Data: b2}, {Data: b3}},
					})
				} else {
					pp = pterm.DefaultPanel.WithPanels(pterm.Panels{
						{{Data: b1}, {Data: b2}},
					})
				}
				panels, _ := pp.Srender()

				pterm.DefaultBox.WithTitle(pterm.LightCyan("wiretap is online!")).
					WithTitleTopLeft().
					WithRightPadding(3).
					WithTopPadding(1).
					WithLeftPadding(3).
					WithBottomPadding(0).
					Println(panels)

				pterm.Println()
				if wiretapConfig.MockMode {
					pterm.Info.Printf("Ⓜ️ Mock mode: wiretap is not proxying any traffic, all responses are %s.\n",
						pterm.LightMagenta("generated mocks/simulations"))

				} else {
					if wiretapConfig.RedirectURL != "" {
						pterm.Info.Printf("wiretap is proxying all traffic to '%s'\n",
							pterm.LightMagenta(wiretapConfig.RedirectURL))
					} else {
						pterm.Info.Printf("no redirect URL configured, wiretap is not operating as a proxy\n")
					}
				}

				pterm.Println()
			}
		}, nil)
	}()
}
