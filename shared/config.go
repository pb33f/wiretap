// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package shared

type WiretapConfiguration struct {
	Contract       string `json:"-"`
	RedirectHost   string `json:"redirectHost,omitempty"`
	Port           string `json:"port,omitempty"`
	MonitorPort    string `json:"monitorPort,omitempty"`
	GlobalAPIDelay int    `json:"globalAPIDelay,omitempty"`
}

const ConfigKey = "config"
