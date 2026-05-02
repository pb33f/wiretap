// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package cmd

import (
	"fmt"
	"os"
)

const (
	bannerPink  = "\033[1;38;2;248;58;255m"
	bannerCyan  = "\033[1;38;2;98;196;255m"
	bannerReset = "\033[0m"
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
	fmt.Println(bannerColor(bannerPink, text))
	fmt.Print(bannerColor(bannerCyan, fmt.Sprintf("wiretap version: %s", Version)))
	fmt.Println(bannerColor(bannerPink, fmt.Sprintf(" | compiled: %s", Date)))
	fmt.Println(bannerColor(bannerCyan, "Designed and built by Princess Beef Heavy Industries: https://pb33f.io/wiretap"))
	fmt.Println()
}

func bannerColor(code, value string) string {
	if os.Getenv("TERM") == "dumb" || value == "" {
		return value
	}
	return code + value + bannerReset
}
