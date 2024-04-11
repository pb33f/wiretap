// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package config

import (
	"fmt"
	"github.com/pb33f/wiretap/shared"
	"strings"
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

func FindPathDelay(path string, configuration *shared.WiretapConfiguration) int {
	var foundMatch int
	for key := range configuration.CompiledPathDelays {
		if configuration.CompiledPathDelays[key].CompiledPathDelay.Match(path) {
			foundMatch = configuration.CompiledPathDelays[key].PathDelayValue
		}
	}
	return foundMatch
}

func IgnoreRedirectOnPath(path string, configuration *shared.WiretapConfiguration) bool {
	for _, redirectPath := range configuration.CompiledIgnoreRedirects {
		if redirectPath.CompiledPath.Match(path) {
			return true
		}
	}
	return false
}

func PathRedirectAllowListed(path string, configuration *shared.WiretapConfiguration) bool {
	for _, redirectPath := range configuration.CompiledRedirectAllowList {
		if redirectPath.CompiledPath.Match(path) {
			return true
		}
	}
	return false
}

func IgnoreValidationOnPath(path string, configuration *shared.WiretapConfiguration) bool {
	for _, redirectPath := range configuration.CompiledIgnoreValidations {
		if redirectPath.CompiledPath.Match(path) {
			return true
		}
	}
	return false
}

func PathValidationAllowListed(path string, configuration *shared.WiretapConfiguration) bool {
	for _, redirectPath := range configuration.CompiledIgnoreValidations {
		if redirectPath.CompiledPath.Match(path) {
			return true
		}
	}
	return false
}

func RewritePath(path string, configuration *shared.WiretapConfiguration) string {
	paths := FindPaths(path, configuration)
	var replaced string = path
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
				target := strings.ReplaceAll(strings.ReplaceAll(configuration.ReplaceWithVariables(pathConfig.Target),
					"http://", ""), "https://", "")

				replaced = fmt.Sprintf("%s%s%s", scheme, target, replacedPath)
				break
			}
		}
		// no rewriting, just replace target.
		if replaced == "" {
			scheme := "http://"
			if pathConfig.Secure {
				scheme = "https://"
			}
			target := strings.ReplaceAll(strings.ReplaceAll(configuration.ReplaceWithVariables(pathConfig.Target),
				"http://", ""), "https://", "")

			if path[0] != '/' && pathConfig.Target[len(pathConfig.Target)-1] != '/' {
				path = fmt.Sprintf("/%s", path)
			}
			replaced = fmt.Sprintf("%s%s%s", scheme, target, path)
		}
	}

	return replaced
}
