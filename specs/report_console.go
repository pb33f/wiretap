// Copyright 2026 Princess Beef Heavy Industries LLC
// SPDX-License-Identifier: AGPL

package specs

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func RenderConsole(report *ConflictReport, out io.Writer) {
	if report == nil {
		report = &ConflictReport{}
	}
	formatter := newSpecPathFormatter(report)
	style := newConsoleStyle(out)

	fmt.Fprintln(out)
	fmt.Fprintln(out, style.title("OpenAPI multi-spec analysis"))
	fmt.Fprintln(out, style.separator(strings.Repeat("-", 28)))

	sections := []struct {
		kind  ConflictKind
		title string
	}{
		{KindCrossSpecDuplicate, "Cross-spec duplicate routes"},
		{KindCrossSpecAmbiguous, "Cross-spec ambiguous routes"},
		{KindWithinSpecDuplicate, "Within-spec duplicate routes"},
		{KindWithinSpecAmbiguous, "Within-spec ambiguous routes"},
		{KindDuplicateOperationID, "Duplicate operationIds"},
	}

	for _, section := range sections {
		renderConflictSection(report, section.kind, section.title, formatter, style, out)
	}
	renderLoadErrors(report.LoadErrors, formatter, style, out)

	specCount := report.SpecCount
	if specCount == 0 {
		specCount = len(uniqueSpecs(report))
	}
	fmt.Fprintf(out, "\n%s conflicts across %s specs", style.issueCount(len(report.Conflicts)), style.issueCount(specCount))
	if len(report.LoadErrors) > 0 {
		fmt.Fprintf(out, " (%s load errors)", style.issueCount(len(report.LoadErrors)))
	}
	fmt.Fprintln(out)
}

func renderConflictSection(report *ConflictReport, kind ConflictKind, title string, formatter specPathFormatter, style consoleStyle, out io.Writer) {
	var rows []Conflict
	for _, conflict := range report.Conflicts {
		if conflict.Kind == kind {
			rows = append(rows, conflict)
		}
	}
	if len(rows) == 0 {
		return
	}

	fmt.Fprintf(out, "\n%s %s\n", style.sectionTitle(title), style.sectionCount(len(rows)))
	if kind == KindDuplicateOperationID {
		renderOperationIDIgnoreHint(style, out)
		for _, row := range rows {
			renderOperationIDConflict(row, formatter, style, out)
		}
	} else {
		fmt.Fprintln(out)
		for _, row := range rows {
			renderRouteConflict(row, formatter, style, out)
		}
	}
}

func renderOperationIDIgnoreHint(style consoleStyle, out io.Writer) {
	fmt.Fprintf(out, "%s %s %s\n\n",
		style.label("Use"),
		style.flag("--ignore-clashing-operationid"),
		style.label("to ignore duplicate operation IDs clashing across specs."))
}

func renderRouteConflict(conflict Conflict, formatter specPathFormatter, style consoleStyle, out io.Writer) {
	routeA := valueAt(conflict.RoutePaths, 0)
	routeB := valueAt(conflict.RoutePaths, 1)
	if routeA == "" {
		routeA = valueAt(conflict.Paths, 0)
	}
	if routeB == "" {
		routeB = valueAt(conflict.Paths, 1)
	}

	method := strings.ToUpper(conflict.Method)
	if method == "" {
		method = "METHOD"
	}

	fmt.Fprintf(out, "%s %s\n", style.method(method), style.path(routeA))
	lines := []detailLine{{Label: "spec", Value: formatter.format(valueAt(conflict.Specs, 0)), Kind: detailSpec}}
	if pathA := valueAt(conflict.Paths, 0); pathA != "" && pathA != routeA {
		lines = append(lines, detailLine{Label: "path", Value: pathA, Kind: detailPath})
	}
	lines = append(lines, detailLine{Label: "conflicts with", Value: formatter.format(valueAt(conflict.Specs, 1)), Kind: detailSpec})
	if routeB != "" && routeB != routeA {
		lines = append(lines, detailLine{Label: "conflict route", Value: routeB, Kind: detailPath})
	}
	if pathB := valueAt(conflict.Paths, 1); pathB != "" && pathB != routeB {
		lines = append(lines, detailLine{Label: "conflict path", Value: pathB, Kind: detailPath})
	}
	renderDetailLines(lines, style, out)
	fmt.Fprintln(out)
}

func renderOperationIDConflict(conflict Conflict, formatter specPathFormatter, style consoleStyle, out io.Writer) {
	operationID := conflict.OperationID
	if operationID == "" {
		operationID = "(empty operationId)"
	}

	fmt.Fprintf(out, "%s %s\n", style.label("operationId:"), style.path(operationID))
	max := len(conflict.Paths)
	if len(conflict.Specs) > max {
		max = len(conflict.Specs)
	}
	for i := 0; i < max; i++ {
		prefix := "├─"
		if i == max-1 {
			prefix = "└─"
		}
		pathText := valueAt(conflict.Paths, i)
		specText := formatter.format(valueAt(conflict.Specs, i))
		switch {
		case pathText != "" && specText != "":
			fmt.Fprintf(out, "  %s %s %s%s%s\n", style.tree(prefix), style.operationPath(pathText), style.dim("("), style.spec(specText), style.dim(")"))
		case pathText != "":
			fmt.Fprintf(out, "  %s %s\n", style.tree(prefix), style.operationPath(pathText))
		case specText != "":
			fmt.Fprintf(out, "  %s %s\n", style.tree(prefix), style.spec(specText))
		}
	}
	fmt.Fprintln(out)
}

type detailKind int

const (
	detailText detailKind = iota
	detailSpec
	detailPath
	detailError
)

type detailLine struct {
	Label string
	Value string
	Kind  detailKind
}

func renderDetailLines(lines []detailLine, style consoleStyle, out io.Writer) {
	for i, line := range lines {
		prefix := "├─"
		if i == len(lines)-1 {
			prefix = "└─"
		}
		fmt.Fprintf(out, "  %s %s %s\n", style.tree(prefix), style.label(line.Label+":"), style.detailValue(line))
	}
}

func renderLoadErrors(loadErrors []LoadError, formatter specPathFormatter, style consoleStyle, out io.Writer) {
	if len(loadErrors) == 0 {
		return
	}
	fmt.Fprintf(out, "\n%s %s\n\n", style.sectionTitle("Load errors"), style.sectionCount(len(loadErrors)))
	for _, loadErr := range loadErrors {
		msg := ""
		if loadErr.Error != nil {
			msg = formatter.compactError(loadErr.Spec, loadErr.Error)
		}
		fmt.Fprintf(out, "%s\n", style.spec(formatter.format(loadErr.Spec)))
		renderDetailLines([]detailLine{{Label: "error", Value: msg, Kind: detailError}}, style, out)
		fmt.Fprintln(out)
	}
}

type consoleStyle struct {
	enabled bool
}

const (
	ansiPrimaryCyan   = "\033[1;38;2;98;196;255m"
	ansiSecondaryPink = "\033[22;38;2;248;58;255m"
	ansiHeaderPink    = "\033[1;38;2;248;58;255m"
	ansiDarkGrey      = "\033[38;5;240m"
	ansiTreeGrey      = "\033[38;5;246m"
	ansiMethodGreen   = "\033[38;5;46m"
	ansiMethodYellow  = "\033[38;5;220m"
	ansiMethodBlue    = "\033[38;5;81m"
	ansiMethodRed     = "\033[38;5;196m"
	ansiErrorRed      = "\033[38;5;196m"
	ansiReset         = "\033[0m"
)

func newConsoleStyle(out io.Writer) consoleStyle {
	if os.Getenv("TERM") == "dumb" {
		return consoleStyle{}
	}
	file, ok := out.(*os.File)
	if !ok {
		return consoleStyle{}
	}
	return consoleStyle{enabled: file == os.Stdout || file == os.Stderr}
}

func (s consoleStyle) title(value string) string {
	return s.color(ansiHeaderPink, value)
}

func (s consoleStyle) separator(value string) string {
	return s.color(ansiTreeGrey, value)
}

func (s consoleStyle) sectionTitle(value string) string {
	return s.color(ansiHeaderPink, value)
}

func (s consoleStyle) sectionCount(count int) string {
	return s.color(ansiDarkGrey, fmt.Sprintf("(%d)", count))
}

func (s consoleStyle) issueCount(count int) string {
	return s.color(ansiPrimaryCyan, fmt.Sprintf("%d", count))
}

func (s consoleStyle) label(value string) string {
	return s.color(ansiDarkGrey, value)
}

func (s consoleStyle) tree(value string) string {
	return s.color(ansiTreeGrey, value)
}

func (s consoleStyle) dim(value string) string {
	return s.color(ansiTreeGrey, value)
}

func (s consoleStyle) path(value string) string {
	return s.color(ansiPrimaryCyan, value)
}

func (s consoleStyle) flag(value string) string {
	return s.color(ansiPrimaryCyan, value)
}

func (s consoleStyle) spec(value string) string {
	return s.color(ansiSecondaryPink, value)
}

func (s consoleStyle) error(value string) string {
	return s.color(ansiErrorRed, value)
}

func (s consoleStyle) method(value string) string {
	switch strings.ToUpper(value) {
	case "GET", "QUERY":
		return s.color(ansiMethodGreen, value)
	case "PATCH":
		return s.color(ansiMethodYellow, value)
	case "PUT", "POST":
		return s.color(ansiMethodBlue, value)
	case "DELETE":
		return s.color(ansiMethodRed, value)
	default:
		return value
	}
}

func (s consoleStyle) operationPath(value string) string {
	method, route, ok := strings.Cut(value, " ")
	if !ok {
		return s.path(value)
	}
	return s.method(method) + " " + s.path(route)
}

func (s consoleStyle) detailValue(line detailLine) string {
	switch line.Kind {
	case detailSpec:
		return s.spec(line.Value)
	case detailPath:
		return s.path(line.Value)
	case detailError:
		return s.error(line.Value)
	default:
		return line.Value
	}
}

func (s consoleStyle) color(code, value string) string {
	if !s.enabled || value == "" {
		return value
	}
	return code + value + ansiReset
}

func uniqueSpecs(report *ConflictReport) map[string]struct{} {
	specs := make(map[string]struct{})
	for _, conflict := range report.Conflicts {
		for _, spec := range conflict.Specs {
			if spec != "" {
				specs[spec] = struct{}{}
			}
		}
	}
	for _, loadErr := range report.LoadErrors {
		if loadErr.Spec != "" {
			specs[loadErr.Spec] = struct{}{}
		}
	}
	return specs
}

type specPathFormatter struct {
	base string
}

func newSpecPathFormatter(report *ConflictReport) specPathFormatter {
	var paths []string
	if report != nil {
		for _, conflict := range report.Conflicts {
			paths = append(paths, conflict.Specs...)
		}
		for _, loadErr := range report.LoadErrors {
			paths = append(paths, loadErr.Spec)
		}
	}
	return specPathFormatter{base: commonSpecDir(paths)}
}

func (f specPathFormatter) format(spec string) string {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return "(unknown spec)"
	}
	if isRemotePath(spec) {
		return spec
	}

	abs, err := filepath.Abs(spec)
	if err != nil {
		return filepath.ToSlash(filepath.Clean(spec))
	}
	if f.base != "" {
		if rel, err := filepath.Rel(f.base, abs); err == nil && rel != "." && !isParentRelative(rel) {
			return filepath.ToSlash(rel)
		}
	}
	if rel, err := filepath.Rel(".", abs); err == nil && rel != "." && !isParentRelative(rel) {
		return filepath.ToSlash(rel)
	}
	return filepath.ToSlash(filepath.Clean(spec))
}

func (f specPathFormatter) compactError(spec string, err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	short := f.format(spec)
	if spec != "" && short != "" {
		msg = strings.ReplaceAll(msg, spec, short)
		if abs, absErr := filepath.Abs(spec); absErr == nil {
			msg = strings.ReplaceAll(msg, abs, short)
		}
	}
	return msg
}

func commonSpecDir(paths []string) string {
	var common string
	for _, spec := range paths {
		spec = strings.TrimSpace(spec)
		if spec == "" || isRemotePath(spec) {
			continue
		}
		abs, err := filepath.Abs(spec)
		if err != nil {
			continue
		}
		dir := filepath.Dir(abs)
		if common == "" {
			common = dir
			continue
		}
		for common != "" && !pathWithin(common, dir) {
			parent := filepath.Dir(common)
			if parent == common {
				return ""
			}
			common = parent
		}
	}
	if common == "." || common == string(os.PathSeparator) {
		return ""
	}
	return common
}

func pathWithin(base, target string) bool {
	rel, err := filepath.Rel(base, target)
	return err == nil && !isParentRelative(rel)
}

func isParentRelative(path string) bool {
	return path == ".." || strings.HasPrefix(path, ".."+string(os.PathSeparator))
}

func isRemotePath(path string) bool {
	u, err := url.Parse(path)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

func valueAt(values []string, index int) string {
	if index < 0 || index >= len(values) {
		return ""
	}
	return values[index]
}
