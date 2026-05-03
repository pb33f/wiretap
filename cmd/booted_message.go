// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: BUSL-1.1

package cmd

import (
	"fmt"

	"github.com/pb33f/doctor/terminal"
	"github.com/pb33f/ranch/bus"
	"github.com/pb33f/ranch/model"
	"github.com/pb33f/ranch/plank/pkg/server"
	"github.com/pb33f/wiretap/shared"
)

func bootedMessage(wiretapConfig *shared.WiretapConfiguration, eventBus bus.EventBus) {
	// print a nice message to the user when ranch is online.
	go func() {
		handler, _ := eventBus.ListenStream(server.RANCH_SERVER_ONLINE_CHANNEL)
		seen := false
		handler.Handle(func(message *model.Message) {
			if !seen {
				seen = true
				fmt.Println()

				b1 := terminal.RenderBox(wiretapConfig.GetApiGateway(), terminal.BoxOptions{Title: style.Secondary("API Gateway")})
				b2 := terminal.RenderBox(wiretapConfig.GetMonitorUI(), terminal.BoxOptions{Title: style.Secondary("Monitor UI")})
				b3 := terminal.RenderBox(wiretapConfig.StaticDir, terminal.BoxOptions{Title: style.Secondary("Static files served from")})

				var panels string
				if wiretapConfig.StaticDir != "" {
					panels = terminal.RenderPanelGrid([][]string{{b1, b2, b3}}, terminal.PanelOptions{Gap: 2})
				} else {
					panels = terminal.RenderPanelGrid([][]string{{b1, b2}}, terminal.PanelOptions{Gap: 2})
				}

				fmt.Println(terminal.RenderBox(panels, terminal.BoxOptions{
					Title:        style.Primary("wiretap is online!"),
					PaddingLeft:  3,
					PaddingRight: 3,
					PaddingTop:   1,
				}))

				fmt.Println()
				if wiretapConfig.MockMode {
					cliLog.Info(fmt.Sprintf("Ⓜ️ Mock mode: wiretap is not proxying any traffic, all responses are %s.",
						style.Secondary("generated mocks/simulations")))

				} else {
					if wiretapConfig.RedirectURL != "" {
						cliLog.Info(fmt.Sprintf("wiretap is proxying all traffic to '%s'",
							style.Secondary(wiretapConfig.RedirectURL)))
					} else {
						cliLog.Info("no redirect URL configured, wiretap is not operating as a proxy")
					}
				}

				fmt.Println()
			}
		}, nil)
	}()
}
