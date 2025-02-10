// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package config

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/pb33f/wiretap/shared"
	"github.com/pterm/pterm"
)

const (
	PascalCaseRewriteIdHeader = "Rewriteid"
	SnakeCaseRewriteIdHeader  = "Rewrite_id"
	KebabCaseRewriteIdHeader  = "Rewrite-Id"
)

var rewriteIdHeaders = []string{
	PascalCaseRewriteIdHeader,
	SnakeCaseRewriteIdHeader,
	KebabCaseRewriteIdHeader,
}

type PathRewrite struct {
	RewrittenPath     string
	PathConfiguration *shared.WiretapPathConfig
}

func FindPaths(path string, configuration *shared.WiretapConfiguration) []*shared.WiretapPathConfig {
	var foundConfigurations []*shared.WiretapPathConfig
	for x := configuration.CompiledPaths.First(); x != nil; x = x.Next() {
		compiledPath := x.Value()
		if compiledPath.CompiledKey.Match(path) {
			foundConfigurations = append(foundConfigurations, compiledPath.PathConfig)
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

func IncludePathOnMockMode(path string, configuration *shared.WiretapConfiguration) bool {
	for _, mockModePath := range configuration.CompiledMockModeList {
		if mockModePath.Match(path) {
			return true
		}
	}
	return false
}

func hardErrorOnPath(path string, configuration *shared.WiretapConfiguration) bool {
	for _, mockModePath := range configuration.CompiledHardErrorList {
		if mockModePath.Match(path) {
			return true
		}
	}
	return false
}

func IsHardErrorsSet(path string, configuration *shared.WiretapConfiguration) bool {
	if configuration.HardErrors || hardErrorOnPath(path, configuration) {
		return true
	}

	return false
}

func IgnoreValidationOnPath(path string, configuration *shared.WiretapConfiguration) bool {
	for _, validationPath := range configuration.CompiledIgnoreValidations {
		if validationPath.CompiledPath.Match(path) {
			return true
		}
	}
	return false
}

func PathValidationAllowListed(path string, configuration *shared.WiretapConfiguration) bool {
	for _, validationPath := range configuration.CompiledValidationAllowList {
		if validationPath.CompiledPath.Match(path) {
			return true
		}
	}
	return false
}

func rewriteTaget(path string, pathConfig *shared.WiretapPathConfig, configuration *shared.WiretapConfiguration) *PathRewrite {
	scheme := "http://"
	if pathConfig.Secure {
		scheme = "https://"
	}
	target := strings.ReplaceAll(strings.ReplaceAll(configuration.ReplaceWithVariables(pathConfig.Target),
		"http://", ""), "https://", "")

	if path[0] != '/' && pathConfig.Target[len(pathConfig.Target)-1] != '/' {
		path = fmt.Sprintf("/%s", path)
	}
	return &PathRewrite{
		RewrittenPath:     fmt.Sprintf("%s%s%s", scheme, target, path),
		PathConfiguration: pathConfig,
	}
}

func getRewriteIdHeaderValues(req *http.Request) ([]string, bool) {

	// Let's first try to get the header with expected key names
	for _, possibleHeaderKey := range rewriteIdHeaders {

		if rewriteIdHeaderValues, ok := req.Header[possibleHeaderKey]; ok {
			return rewriteIdHeaderValues, true
		}

		if rewriteIdHeaderValues, ok := req.Header[strings.ToLower(possibleHeaderKey)]; ok {
			return rewriteIdHeaderValues, true
		}

	}

	// Let's now try to ignore case ; this may produce collisions if a user has two headers with similar keys,
	// but different capitalization. This is okay, as this is a last ditch effort to find any possible match
	loweredHeaders := map[string][]string{}

	for headerKey, headerValues := range req.Header {
		loweredKey := strings.ToLower(headerKey)

		if _, ok := loweredHeaders[loweredKey]; ok {
			loweredHeaders[loweredKey] = append(loweredHeaders[loweredKey], headerValues...)
		} else {
			loweredHeaders[loweredKey] = headerValues
		}

	}

	for _, possibleHeaderKey := range rewriteIdHeaders {

		if rewriteIdHeaderValues, ok := loweredHeaders[strings.ToLower(possibleHeaderKey)]; ok {
			return rewriteIdHeaderValues, true
		}

	}

	return []string{}, false
}

func FindPathWithRewriteId(paths []*shared.WiretapPathConfig, req *http.Request) *shared.WiretapPathConfig {

	if req == nil {
		return nil
	}

	if rewriteIdHeaderValues, ok := getRewriteIdHeaderValues(req); ok {
		for _, pathRewriteConfig := range paths {

			// Iterate through header values - since it's a multi-value field
			for _, rewriteId := range rewriteIdHeaderValues {
				if pathRewriteConfig.RewriteId == rewriteId {
					return pathRewriteConfig
				}
			}

		}
	}

	return nil
}

func RewritePath(path string, req *http.Request, configuration *shared.WiretapConfiguration) *PathRewrite {
	paths := FindPaths(path, configuration)

	// If there are no configurations that match the request path, we should crash out early
	if len(paths) == 0 {
		return &PathRewrite{
			RewrittenPath:     path,
			PathConfiguration: nil,
		}
	}

	var pathConfig *shared.WiretapPathConfig

	// Check if request headers have rewrite id; if so, we should try to find a matching rewrite config
	pathConfig = FindPathWithRewriteId(paths, req)

	// if rewriteId not specified in request or not found, extract first path
	if pathConfig == nil {
		pathConfig = paths[0]
	}

	for _, globalIgnoreRewrite := range configuration.CompiledIgnorePathRewrite {
		// If the current path matches the ignore rewrite, we should skip rewriting,
		// and instead check if we even want to rewrite the target
		if globalIgnoreRewrite.CompiledIgnoreRewrite.Match(path) {
			if globalIgnoreRewrite.RewriteTarget {
				return rewriteTaget(path, pathConfig, configuration)
			} else {
				pterm.Info.Printf("[wiretap] Not re-writing path '%s' due to global ignore rewrite configuration\n", path)
				return &PathRewrite{
					RewrittenPath:     path,
					PathConfiguration: pathConfig,
				}
			}
		}
	}

	var replaced = ""

	for key := range pathConfig.CompiledPath.CompiledPathRewrite {
		if pathConfig.CompiledPath.CompiledPathRewrite[key].MatchString(path) {

			// Check if this path matches a local ignore rewrite. If so, then check if we need to rewrite the target
			for _, ignoreRewrite := range pathConfig.CompiledIgnoreRewrite {
				if ignoreRewrite.CompiledIgnoreRewrite.Match(path) {
					if ignoreRewrite.RewriteTarget {
						return rewriteTaget(path, pathConfig, configuration)
					} else {
						pterm.Info.Printf("[wiretap] Not re-writing path '%s' due to local ignore rewrite configuration\n", path)
						return &PathRewrite{
							RewrittenPath:     path,
							PathConfiguration: pathConfig,
						}
					}
				}
			}

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

	// If we already replaced the path, then we should just return that
	if replaced != "" {
		return &PathRewrite{
			RewrittenPath:     replaced,
			PathConfiguration: pathConfig,
		}
	}

	// Otherwise, there's no rewriting,  and we just try to replace the target.
	return rewriteTaget(path, pathConfig, configuration)
}
