// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: BUSL-1.1

package cmd

import (
	"os"

	"github.com/pb33f/doctor/terminal"
)

func PrintBanner() {
	terminal.PrintBanner(terminal.BannerOptions{
		Writer:      os.Stdout,
		Palette:     terminal.PaletteForTheme(terminal.ThemeDark),
		ProductName: "wiretap",
		ProductURL:  "Designed and built by Princess Beef Heavy Industries, LLC (pb33f): https://pb33f.io/wiretap",
		Version:     Version,
		Date:        Date,
	})
}
