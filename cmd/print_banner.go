// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"
	"github.com/pterm/pterm"
)

func PrintBanner() {
	text := `
@@@@@@@   @@@@@@@   @@@@@@   @@@@@@   @@@@@@@@
@@@@@@@@  @@@@@@@@  @@@@@@@  @@@@@@@  @@@@@@@@
@@!  @@@  @@!  @@@      @@@      @@@  @@!
!@!  @!@  !@   @!@      @!@      @!@  !@!
@!@@!@!   @!@!@!@   @!@!!@   @!@!!@   @!!!:!
!!@!!!    !!!@!!!!  !!@!@!   !!@!@!   !!!!!:
!!:       !!:  !!!      !!:      !!:  !!:
:!:       :!:  !:!      :!:      :!:  :!:
 ::        :: ::::  :: ::::  :: ::::   ::
 :        :: : ::    : : :    : : :    :`
	pterm.DefaultBasicText.Println(pterm.LightMagenta(text))
	pterm.Print(pterm.LightCyan(fmt.Sprintf("wiretap version: %s", Version)))
	pterm.Println(pterm.LightMagenta(fmt.Sprintf(" | compiled: %s", Date)))
	pterm.Println(pterm.LightCyan("Designed and built by Princess Beef Heavy Industries: https://pb33f.io/wiretap"))
	pterm.Println()
}
