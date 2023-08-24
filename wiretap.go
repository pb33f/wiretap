// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

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
		date = time.Now().Format("Mon, 02 Jan 2006 15:04:05 MST")
	} else {
		parsed, _ := time.Parse(time.RFC3339, date)
		date = parsed.Format("Mon, 02 Jan 2006 15:04:05 MST")
	}

	// run root command.
	cmd.Execute(version, commit, date, staticUI)
}
