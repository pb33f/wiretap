// Copyright 2026 Princess Beef Heavy Industries LLC
// SPDX-License-Identifier: AGPL

package specs

import (
	"bytes"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
)

var specExtensions = map[string]struct{}{
	".yaml": {},
	".yml":  {},
	".json": {},
}

// DiscoverSpecs returns a deduplicated, ordered list of local OpenAPI spec
// files and remote spec URLs. Explicit roots are returned before directory
// discoveries; that ordering is the routing tiebreak.
func DiscoverSpecs(roots []string, dirs []string, ignorePatterns []string) ([]string, error) {
	ignoreGlobs, err := compileIgnorePatterns(ignorePatterns)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})
	var discovered []string
	appendSpec := func(path string) {
		if path == "" {
			return
		}
		if _, ok := seen[path]; ok {
			return
		}
		seen[path] = struct{}{}
		discovered = append(discovered, path)
	}

	for _, root := range roots {
		root = strings.TrimSpace(root)
		if root == "" {
			continue
		}
		if isRemoteSpec(root) {
			appendSpec(root)
			continue
		}
		if hasGlobMeta(root) {
			matches, err := expandGlob(root, ignoreGlobs)
			if err != nil {
				return nil, err
			}
			for _, match := range matches {
				appendSpec(match)
			}
			continue
		}
		abs, err := filepath.Abs(root)
		if err != nil {
			return nil, err
		}
		appendSpec(abs)
	}

	for _, dir := range dirs {
		dir = strings.TrimSpace(dir)
		if dir == "" {
			continue
		}
		matches, err := walkDir(dir, ignoreGlobs)
		if err != nil {
			return nil, err
		}
		for _, match := range matches {
			appendSpec(match)
		}
	}

	return discovered, nil
}

func isRemoteSpec(path string) bool {
	u, err := url.Parse(path)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

func hasGlobMeta(path string) bool {
	return strings.ContainsAny(path, "*?[{")
}

func compileIgnorePatterns(patterns []string) ([]glob.Glob, error) {
	compiled := make([]glob.Glob, 0, len(patterns)*2)
	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}
		g, err := glob.Compile(filepath.ToSlash(pattern), '/')
		if err != nil {
			return nil, fmt.Errorf("compile ignore pattern %q: %w", pattern, err)
		}
		compiled = append(compiled, g)
		if !filepath.IsAbs(pattern) {
			abs, err := filepath.Abs(pattern)
			if err != nil {
				return nil, err
			}
			g, err = glob.Compile(filepath.ToSlash(abs), '/')
			if err != nil {
				return nil, fmt.Errorf("compile ignore pattern %q: %w", pattern, err)
			}
			compiled = append(compiled, g)
		}
	}
	return compiled, nil
}

func walkDir(root string, ignoreGlobs []glob.Glob) ([]string, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	var matches []string
	err = filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path != absRoot && isIgnored(path, absRoot, ignoreGlobs) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if !isSpecExtension(path) || !containsOpenAPIMarkers(path) {
			return nil
		}
		abs, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		matches = append(matches, abs)
		return nil
	})
	return matches, err
}

func expandGlob(pattern string, ignoreGlobs []glob.Glob) ([]string, error) {
	absPattern, err := filepath.Abs(pattern)
	if err != nil {
		return nil, err
	}
	base := globBase(pattern)
	absBase, err := filepath.Abs(base)
	if err != nil {
		return nil, err
	}
	matcher, err := glob.Compile(filepath.ToSlash(absPattern), '/')
	if err != nil {
		return nil, fmt.Errorf("compile spec glob %q: %w", pattern, err)
	}
	matchers := []glob.Glob{matcher}
	if strings.Contains(filepath.ToSlash(absPattern), "**/") {
		zeroDepthPattern := strings.Replace(filepath.ToSlash(absPattern), "**/", "", 1)
		zeroDepthMatcher, err := glob.Compile(zeroDepthPattern, '/')
		if err != nil {
			return nil, fmt.Errorf("compile spec glob %q: %w", pattern, err)
		}
		matchers = append(matchers, zeroDepthMatcher)
	}

	var matches []string
	err = filepath.WalkDir(absBase, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path != absBase && isIgnored(path, absBase, ignoreGlobs) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		abs, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		if !matchesAny(matchers, filepath.ToSlash(abs)) || !isSpecExtension(abs) || !containsOpenAPIMarkers(abs) {
			return nil
		}
		matches = append(matches, abs)
		return nil
	})
	return matches, err
}

func matchesAny(matchers []glob.Glob, value string) bool {
	for _, matcher := range matchers {
		if matcher.Match(value) {
			return true
		}
	}
	return false
}

func globBase(pattern string) string {
	firstMeta := strings.IndexAny(pattern, "*?[{")
	if firstMeta < 0 {
		return filepath.Dir(pattern)
	}
	prefix := pattern[:firstMeta]
	sep := strings.LastIndexAny(prefix, `/\`)
	if sep < 0 {
		return "."
	}
	base := prefix[:sep]
	if base == "" {
		return string(filepath.Separator)
	}
	return base
}

func isSpecExtension(path string) bool {
	_, ok := specExtensions[strings.ToLower(filepath.Ext(path))]
	return ok
}

func containsOpenAPIMarkers(path string) bool {
	content, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	lower := bytes.ToLower(content)
	return bytes.Contains(lower, []byte("openapi:")) ||
		bytes.Contains(lower, []byte(`"openapi"`)) ||
		bytes.Contains(lower, []byte("swagger:")) ||
		bytes.Contains(lower, []byte(`"swagger"`))
}

func isIgnored(path, root string, ignoreGlobs []glob.Glob) bool {
	if len(ignoreGlobs) == 0 {
		return false
	}
	abs := filepath.ToSlash(path)
	rel, err := filepath.Rel(root, path)
	if err != nil {
		rel = path
	}
	rel = filepath.ToSlash(rel)
	for _, g := range ignoreGlobs {
		if g.Match(abs) || g.Match(rel) {
			return true
		}
	}
	return false
}
