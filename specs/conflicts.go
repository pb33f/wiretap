// Copyright 2026 Princess Beef Heavy Industries LLC
// SPDX-License-Identifier: AGPL

package specs

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	doctorHelpers "github.com/pb33f/doctor/helpers"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/wiretap/shared"
)

type ConflictKind string

const (
	KindCrossSpecDuplicate   ConflictKind = "cross-spec-duplicate"
	KindCrossSpecAmbiguous   ConflictKind = "cross-spec-ambiguous"
	KindWithinSpecDuplicate  ConflictKind = "within-spec-duplicate"
	KindWithinSpecAmbiguous  ConflictKind = "within-spec-ambiguous"
	KindDuplicateOperationID ConflictKind = "duplicate-operation-id"
)

type LoadError struct {
	Spec  string
	Error error
}

type Conflict struct {
	Kind        ConflictKind
	Method      string
	Paths       []string
	RoutePaths  []string
	Specs       []string
	OperationID string
}

type ConflictReport struct {
	Conflicts  []Conflict
	LoadErrors []LoadError
	RouteIndex *RouteConflictIndex
	SpecCount  int
}

type AnalyzeOptions struct {
	IgnoreClashingOperationID bool
}

type RouteConflictIndex struct {
	Entries map[string][]RouteConflict
}

type RouteConflict struct {
	Kind              ConflictKind
	MatchedSpec       string
	MatchedPath       string
	MatchedRoutePath  string
	ConflictSpec      string
	ConflictPath      string
	ConflictRoutePath string
	Method            string
}

type operationEntry struct {
	specIndex   int
	specName    string
	path        string
	routePath   string
	method      string
	paramTypes  map[string]string
	operationID string
}

func Analyze(docs []shared.ApiDocument, options ...AnalyzeOptions) *ConflictReport {
	opts := AnalyzeOptions{}
	if len(options) > 0 {
		opts = options[0]
	}
	report := &ConflictReport{
		RouteIndex: &RouteConflictIndex{Entries: map[string][]RouteConflict{}},
		SpecCount:  len(docs),
	}

	entries := collectOperations(docs)
	byMethod := make(map[string][]operationEntry)
	for _, entry := range entries {
		byMethod[entry.method] = append(byMethod[entry.method], entry)
	}

	for _, methodEntries := range byMethod {
		for i := 0; i < len(methodEntries); i++ {
			for j := i + 1; j < len(methodEntries); j++ {
				entryA := methodEntries[i]
				entryB := methodEntries[j]
				conflict, ok := pathConflict(entryA, entryB)
				if !ok {
					continue
				}
				report.Conflicts = append(report.Conflicts, conflict)
				report.RouteIndex.Add(entryA, entryB, conflict.Kind)
				report.RouteIndex.Add(entryB, entryA, conflict.Kind)
			}
		}
	}

	if !opts.IgnoreClashingOperationID {
		for _, conflict := range duplicateOperationIDConflicts(entries) {
			report.Conflicts = append(report.Conflicts, conflict)
		}
	}

	sortConflicts(report.Conflicts)
	return report
}

func collectOperations(docs []shared.ApiDocument) []operationEntry {
	var entries []operationEntry
	for specIndex, doc := range docs {
		model := doc.DocumentModel
		if model == nil && doc.Document != nil {
			built, err := doc.Document.BuildV3Model()
			if err == nil && built != nil {
				model = built
			}
		}
		if model == nil || model.Model.Paths == nil || model.Model.Paths.PathItems == nil {
			continue
		}

		for pathName, pathItem := range model.Model.Paths.PathItems.FromOldest() {
			if pathItem == nil {
				continue
			}
			for methodName, operation := range pathItem.GetOperations().FromOldest() {
				if operation == nil {
					continue
				}
				routePaths := EffectiveOperationRoutePaths(&model.Model, pathItem, operation, pathName)
				method := strings.ToUpper(methodName)
				paramTypes := parameterTypes(pathItem.Parameters, operation.Parameters)
				for _, routePath := range routePaths {
					entries = append(entries, operationEntry{
						specIndex:   specIndex,
						specName:    doc.DocumentName,
						path:        pathName,
						routePath:   routePath,
						method:      method,
						paramTypes:  paramTypes,
						operationID: operation.OperationId,
					})
				}
			}
		}
	}
	return entries
}

func pathConflict(a, b operationEntry) (Conflict, bool) {
	if sameOperation(a, b) {
		return Conflict{}, false
	}

	if a.routePath == b.routePath {
		kind := KindCrossSpecDuplicate
		if a.specIndex == b.specIndex {
			kind = KindWithinSpecDuplicate
		}
		return Conflict{
			Kind:       kind,
			Method:     a.method,
			Paths:      []string{a.path, b.path},
			RoutePaths: []string{a.routePath, b.routePath},
			Specs:      []string{a.specName, b.specName},
		}, true
	}

	if !routePathsOverlap(a, b) {
		return Conflict{}, false
	}

	kind := KindCrossSpecAmbiguous
	if a.specIndex == b.specIndex {
		kind = KindWithinSpecAmbiguous
	}
	return Conflict{
		Kind:       kind,
		Method:     a.method,
		Paths:      []string{a.path, b.path},
		RoutePaths: []string{a.routePath, b.routePath},
		Specs:      []string{a.specName, b.specName},
	}, true
}

func sameOperation(a, b operationEntry) bool {
	return a.specIndex == b.specIndex && a.path == b.path && a.method == b.method
}

func routePathsOverlap(a, b operationEntry) bool {
	if hasTrailingSlash(a.routePath) != hasTrailingSlash(b.routePath) {
		return false
	}
	if doctorHelpers.CheckPaths(a.routePath, b.routePath, a.paramTypes, b.paramTypes) {
		return true
	}
	if a.specIndex == b.specIndex {
		return false
	}
	return routeTemplatesCanMatchSameRequest(a.routePath, b.routePath, a.paramTypes, b.paramTypes)
}

func routeTemplatesCanMatchSameRequest(pathA, pathB string, paramsA, paramsB map[string]string) bool {
	segsA := doctorHelpers.ParseSegments(pathA, paramsA)
	segsB := doctorHelpers.ParseSegments(pathB, paramsB)
	if len(segsA) != len(segsB) {
		return false
	}
	for i := range segsA {
		if !segmentsCanMatchSameRequest(segsA[i], segsB[i]) {
			return false
		}
	}
	return true
}

func segmentsCanMatchSameRequest(a, b doctorHelpers.Segment) bool {
	switch {
	case a.IsVar && b.IsVar:
		if a.Operator != b.Operator {
			return false
		}
		if !pathParamTypesOverlap(a.ParamType, b.ParamType) {
			return false
		}
		return true
	case !a.IsVar && !b.IsVar:
		return a.Value == b.Value
	case a.IsVar:
		return doctorHelpers.CanLiteralMatchType(b.Value, a.ParamType)
	default:
		return doctorHelpers.CanLiteralMatchType(a.Value, b.ParamType)
	}
}

func pathParamTypesOverlap(typeA, typeB string) bool {
	if typeA == "" || typeB == "" || typeA == "string" || typeB == "string" {
		return true
	}
	if typeA == typeB {
		return true
	}
	if typeA == "integer" && typeB == "number" {
		return true
	}
	return typeA == "number" && typeB == "integer"
}

func duplicateOperationIDConflicts(entries []operationEntry) []Conflict {
	byID := make(map[string][]operationEntry)
	for _, entry := range entries {
		if entry.operationID == "" {
			continue
		}
		byID[entry.operationID] = append(byID[entry.operationID], entry)
	}

	var conflicts []Conflict
	for operationID, bucket := range byID {
		if len(bucket) < 2 {
			continue
		}
		specs := make([]string, 0, len(bucket))
		paths := make([]string, 0, len(bucket))
		seen := make(map[string]struct{})
		for _, entry := range bucket {
			key := entry.specName + " " + entry.path + " " + entry.method
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			specs = append(specs, entry.specName)
			paths = append(paths, fmt.Sprintf("%s %s", entry.method, entry.path))
		}
		if len(paths) < 2 {
			continue
		}
		conflicts = append(conflicts, Conflict{
			Kind:        KindDuplicateOperationID,
			Paths:       paths,
			Specs:       specs,
			OperationID: operationID,
		})
	}
	return conflicts
}

func (idx *RouteConflictIndex) Add(matched, conflict operationEntry, kind ConflictKind) {
	if idx == nil {
		return
	}
	if idx.Entries == nil {
		idx.Entries = map[string][]RouteConflict{}
	}
	key := RouteKey(matched.method, matched.routePath)
	idx.Entries[key] = append(idx.Entries[key], RouteConflict{
		Kind:              kind,
		MatchedSpec:       matched.specName,
		MatchedPath:       matched.path,
		MatchedRoutePath:  matched.routePath,
		ConflictSpec:      conflict.specName,
		ConflictPath:      conflict.path,
		ConflictRoutePath: conflict.routePath,
		Method:            matched.method,
	})
}

func (idx *RouteConflictIndex) Lookup(method, routePath string) []RouteConflict {
	if idx == nil || len(idx.Entries) == 0 {
		return nil
	}
	return idx.Entries[RouteKey(method, routePath)]
}

func RouteKey(method, routePath string) string {
	return strings.ToUpper(method) + " " + routePath
}

func parameterTypes(pathParams, operationParams []*v3.Parameter) map[string]string {
	types := make(map[string]string)
	add := func(params []*v3.Parameter) {
		for _, param := range params {
			if param == nil || param.In != "path" || param.Name == "" || param.Schema == nil {
				continue
			}
			schema := param.Schema.Schema()
			if schema == nil || len(schema.Type) == 0 {
				continue
			}
			types[param.Name] = schema.Type[0]
		}
	}
	add(pathParams)
	add(operationParams)
	return types
}

func EffectiveRoutePaths(doc *v3.Document, contractPath string) []string {
	basePaths := ServerBasePaths(doc)
	return routePathsFromBasePaths(basePaths, contractPath)
}

func EffectiveOperationRoutePaths(doc *v3.Document, pathItem *v3.PathItem, operation *v3.Operation, contractPath string) []string {
	basePaths := EffectiveOperationServerBasePaths(doc, pathItem, operation)
	return routePathsFromBasePaths(basePaths, contractPath)
}

func routePathsFromBasePaths(basePaths []string, contractPath string) []string {
	paths := make([]string, 0, len(basePaths))
	for _, basePath := range basePaths {
		paths = append(paths, JoinRoutePath(basePath, contractPath))
	}
	return paths
}

func ServerBasePaths(doc *v3.Document) []string {
	if doc == nil {
		return []string{""}
	}
	return serverBasePathsFromServers(doc.Servers)
}

func EffectiveOperationServerBasePaths(doc *v3.Document, pathItem *v3.PathItem, operation *v3.Operation) []string {
	if operation != nil && len(operation.Servers) > 0 {
		return serverBasePathsFromServers(operation.Servers)
	}
	if pathItem != nil && len(pathItem.Servers) > 0 {
		return serverBasePathsFromServers(pathItem.Servers)
	}
	return ServerBasePaths(doc)
}

func serverBasePathsFromServers(servers []*v3.Server) []string {
	if len(servers) == 0 {
		return []string{""}
	}
	seen := make(map[string]struct{})
	var basePaths []string
	for _, server := range servers {
		if server == nil || strings.TrimSpace(server.URL) == "" {
			continue
		}
		basePath := serverPath(resolveServerURL(server))
		if _, ok := seen[basePath]; ok {
			continue
		}
		seen[basePath] = struct{}{}
		basePaths = append(basePaths, basePath)
	}
	if len(basePaths) == 0 {
		return []string{""}
	}
	return basePaths
}

func resolveServerURL(server *v3.Server) string {
	if server == nil {
		return ""
	}
	rawURL := server.URL
	if server.Variables == nil {
		return rawURL
	}
	for name, variable := range server.Variables.FromOldest() {
		value := serverVariableDefault(variable)
		if value == "" {
			continue
		}
		rawURL = strings.ReplaceAll(rawURL, "{"+name+"}", value)
	}
	return rawURL
}

func serverVariableDefault(variable *v3.ServerVariable) string {
	if variable == nil {
		return ""
	}
	if value := strings.TrimSpace(variable.Default); value != "" {
		return value
	}
	if len(variable.Enum) == 1 {
		return strings.TrimSpace(variable.Enum[0])
	}
	return ""
}

func serverPath(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err == nil && u != nil && u.Path != "" {
		return cleanBasePath(u.Path)
	}

	_, serverPath, ok := strings.Cut(strings.Replace(rawURL, "//", "", 1), "/")
	if !ok || serverPath == "" {
		return ""
	}
	return cleanBasePath("/" + serverPath)
}

func JoinRoutePath(basePath, contractPath string) string {
	contractPath = cleanRoutePath(contractPath)
	if basePath == "" || basePath == "/" {
		return contractPath
	}
	basePath = cleanBasePath(basePath)
	if contractPath == "" {
		return basePath
	}
	if contractPath == "/" {
		return basePath + "/"
	}
	return basePath + "/" + strings.TrimPrefix(contractPath, "/")
}

func cleanRoutePath(routePath string) string {
	if routePath == "" {
		return ""
	}
	if strings.HasPrefix(routePath, "/") {
		return routePath
	}
	return "/" + routePath
}

func cleanBasePath(routePath string) string {
	routePath = cleanRoutePath(routePath)
	if routePath == "/" {
		return routePath
	}
	return strings.TrimRight(routePath, "/")
}

func hasTrailingSlash(routePath string) bool {
	return len(routePath) > 1 && strings.HasSuffix(routePath, "/")
}

func sortConflicts(conflicts []Conflict) {
	sort.SliceStable(conflicts, func(i, j int) bool {
		a, b := conflicts[i], conflicts[j]
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		aSpec, bSpec := first(a.Specs), first(b.Specs)
		if aSpec != bSpec {
			return aSpec < bSpec
		}
		if a.Method != b.Method {
			return a.Method < b.Method
		}
		return first(a.Paths) < first(b.Paths)
	})
}

func first(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
