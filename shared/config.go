// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package shared

import "embed"

type WiretapConfiguration struct {
	Contract       string   `json:"-"`
	RedirectHost   string   `json:"redirectHost,omitempty"`
	Port           string   `json:"port,omitempty"`
	MonitorPort    string   `json:"monitorPort,omitempty"`
	GlobalAPIDelay int      `json:"globalAPIDelay,omitempty"`
	FS             embed.FS `json:"-"`
}

const ConfigKey = "config"
