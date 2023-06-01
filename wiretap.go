// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package main

import (
	"embed"
	"github.com/pb33f/wiretap/cmd"
	"time"
)

//go:embed ui/dist
var staticUI embed.FS

// defined at build time
var version string
var commit string
var date string

func main() {
	if version == "" {
		version = "latest"
	}
	if commit == "" {
		commit = "latest"
	}
	if date == "" {
		date = time.Now().Format("2006-01-02 15:04:05 MST")
	}

	// run root command.
	cmd.Execute(version, commit, date, staticUI)
}
