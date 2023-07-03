// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package config

import (
	"fmt"
	"github.com/pb33f/wiretap/shared"
)

func FindPaths(path string, configuration *shared.WiretapConfiguration) []*shared.WiretapPathConfig {
	var foundConfigurations []*shared.WiretapPathConfig
	for key := range configuration.CompiledPaths {
		if configuration.CompiledPaths[key].CompiledKey.Match(path) {
			foundConfigurations = append(foundConfigurations, configuration.CompiledPaths[key].PathConfig)
		}
	}
	return foundConfigurations
}

func RewritePath(path string, configuration *shared.WiretapConfiguration) string {
	paths := FindPaths(path, configuration)
	var replaced string
	if len(paths) > 0 {
		// extract first path
		pathConfig := paths[0]
		replaced = ""
		for key := range pathConfig.CompiledPath.CompiledPathRewrite {
			if pathConfig.CompiledPath.CompiledPathRewrite[key].MatchString(path) {
				replace := pathConfig.PathRewrite[key]
				rex := pathConfig.CompiledPath.CompiledPathRewrite[key]
				replacedPath := rex.ReplaceAllString(path, replace)

				scheme := "http://"
				if pathConfig.Secure {
					scheme = "https://"
				}
				if replacedPath[0] != '/' && pathConfig.Target[len(pathConfig.Target)-1] != '/' {
					replacedPath = fmt.Sprintf("/%s", replacedPath)
				}
				replaced = fmt.Sprintf("%s%s%s", scheme, pathConfig.Target, replacedPath)
				break
			}
		}
	}
	return replaced
}
