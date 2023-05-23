// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package daemon

type WiretapServiceConfiguration struct {
	Contract       string
	RedirectHost   string
	Port           string
	MonitorPort    string
	GlobalAPIDelay int // how long to delay all API responses by.
}
