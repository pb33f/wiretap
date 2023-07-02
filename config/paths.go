// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package config

import (
	"github.com/pb33f/wiretap/shared"
)

func FindPath(path string, configuration *shared.WiretapConfiguration) []*shared.WiretapPathConfig {
	var foundConfigurations []*shared.WiretapPathConfig
	for key := range configuration.CompiledPaths {
		if configuration.CompiledPaths[key].CompiledKey.Match(path) {
			foundConfigurations = append(foundConfigurations, configuration.CompiledPaths[key].PathConfig)
		}
	}
	return foundConfigurations
}
