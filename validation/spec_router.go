// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package validation

import (
	"net/http"
	"net/url"
	"strings"

	lru "github.com/hashicorp/golang-lru/v2"
	validatorHelpers "github.com/pb33f/libopenapi-validator/helpers"
	"github.com/pb33f/libopenapi-validator/paths"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/wiretap/specs"
)

const specRouterCacheSize = 256

type DocumentValidator struct {
	DocumentName string
	DocModel     *v3.Document
	Validator    HttpValidator
}

type SpecRouter struct {
	validators []DocumentValidator
	cache      *lru.Cache[string, RouteMatch]
}

type RouteMatch struct {
	Index              int
	Document           *DocumentValidator
	MatchedPath        string
	EffectiveRoutePath string
	BasePath           string
	MethodMatched      bool
}

type routeCandidate struct {
	matchedPath        string
	effectiveRoutePath string
	basePath           string
	score              int
	methodMatched      bool
	operationScoped    bool
}

func NewSpecRouter(validators []DocumentValidator) *SpecRouter {
	docs := make([]DocumentValidator, len(validators))
	copy(docs, validators)

	router := &SpecRouter{validators: docs}
	if len(docs) > 1 {
		cache, _ := lru.New[string, RouteMatch](specRouterCacheSize)
		router.cache = cache
	}
	return router
}

func (r *SpecRouter) Resolve(request *http.Request) *DocumentValidator {
	match := r.ResolveMatch(request)
	if match == nil {
		return nil
	}
	return match.Document
}

func (r *SpecRouter) ResolveIndex(request *http.Request) (int, *DocumentValidator) {
	match := r.ResolveMatch(request)
	if match == nil {
		return -1, nil
	}
	return match.Index, match.Document
}

func (r *SpecRouter) ResolveMatch(request *http.Request) *RouteMatch {
	if r == nil || len(r.validators) == 0 || request == nil {
		return nil
	}
	if len(r.validators) == 1 {
		return r.matchValidator(0, request, false)
	}
	if request.URL == nil {
		return &RouteMatch{Index: 0, Document: &r.validators[0]}
	}

	key := routeCacheKey(request)
	if r.cache != nil {
		if match, ok := r.cache.Get(key); ok && match.Index >= 0 && match.Index < len(r.validators) {
			match.Document = &r.validators[match.Index]
			return &match
		}
	}

	var fallback *RouteMatch
	for i := range r.validators {
		match := r.matchValidator(i, request, true)
		if match == nil || match.MatchedPath == "" {
			continue
		}
		if match.MethodMatched {
			r.cacheResult(key, *match)
			return match
		}
		if fallback == nil {
			fallbackMatch := *match
			fallback = &fallbackMatch
		}
	}
	if fallback != nil {
		r.cacheResult(key, *fallback)
		return fallback
	}

	match := RouteMatch{Index: 0, Document: &r.validators[0]}
	r.cacheResult(key, match)
	return &match
}

func (r *SpecRouter) matchValidator(index int, request *http.Request, requirePathMatch bool) *RouteMatch {
	if index < 0 || index >= len(r.validators) {
		return nil
	}
	validator := &r.validators[index]
	match := &RouteMatch{Index: index, Document: validator}
	if request == nil || request.URL == nil || validator.DocModel == nil {
		if requirePathMatch {
			return nil
		}
		return match
	}

	if routeMatch := matchEffectiveRoute(request, validator.DocModel); routeMatch != nil {
		routeMatch.Index = index
		routeMatch.Document = validator
		return routeMatch
	}

	pathItem, _, matchedPath := paths.FindPath(request, validator.DocModel, nil)
	if pathItem == nil && requirePathMatch {
		return nil
	}
	match.MatchedPath = matchedPath
	match.EffectiveRoutePath = effectiveRoutePath(request, validator.DocModel, matchedPath)
	match.MethodMatched = pathHasMethod(pathItem, request.Method)
	return match
}

func effectiveRoutePath(request *http.Request, doc *v3.Document, matchedPath string) string {
	if matchedPath == "" {
		return ""
	}
	if request == nil || request.URL == nil {
		return specs.JoinRoutePath("", matchedPath)
	}
	for _, basePath := range specs.ServerBasePaths(doc) {
		if basePath == "" || basePath == "/" {
			continue
		}
		if request.URL.Path == basePath || strings.HasPrefix(request.URL.Path, basePath+"/") {
			return specs.JoinRoutePath(basePath, matchedPath)
		}
	}
	return specs.JoinRoutePath("", matchedPath)
}

func matchEffectiveRoute(request *http.Request, doc *v3.Document) *RouteMatch {
	if request == nil || request.URL == nil || doc == nil || doc.Paths == nil || doc.Paths.PathItems == nil {
		return nil
	}

	var bestWithMethod *routeCandidate
	var bestOverall *routeCandidate
	requestPath := request.URL.EscapedPath()

	for pathName, pathItem := range doc.Paths.PathItems.FromOldest() {
		if pathItem == nil {
			continue
		}
		for methodName, operation := range pathItem.GetOperations().FromOldest() {
			if operation == nil {
				continue
			}
			methodMatched := operationMatchesMethod(pathItem, methodName, request.Method)
			operationScoped := len(operation.Servers) > 0
			for _, basePath := range specs.EffectiveOperationServerBasePaths(doc, pathItem, operation) {
				routePath := specs.JoinRoutePath(basePath, pathName)
				if !routePathMatches(routePath, requestPath) {
					continue
				}
				candidate := routeCandidate{
					matchedPath:        pathName,
					effectiveRoutePath: routePath,
					basePath:           basePath,
					score:              routeSpecificityScore(routePath),
					methodMatched:      methodMatched,
					operationScoped:    operationScoped,
				}
				if candidate.methodMatched && (bestWithMethod == nil || candidate.score > bestWithMethod.score) {
					bestWithMethod = &candidate
				}
				if candidate.operationScoped {
					// Operation-level servers only apply to that operation; using them for a
					// missing-method fallback would strip another method's base path.
					continue
				}
				if bestOverall == nil || candidate.score > bestOverall.score {
					bestOverall = &candidate
				}
			}
		}
	}

	selected := bestWithMethod
	if selected == nil {
		selected = bestOverall
	}
	if selected == nil {
		return nil
	}
	return &RouteMatch{
		MatchedPath:        selected.matchedPath,
		EffectiveRoutePath: selected.effectiveRoutePath,
		BasePath:           selected.basePath,
		MethodMatched:      selected.methodMatched,
	}
}

func routePathMatches(routePath, requestPath string) bool {
	return RoutePathMatches(routePath, requestPath)
}

func RoutePathMatches(routePath, requestPath string) bool {
	if routePath == requestPath {
		return true
	}
	regex, err := validatorHelpers.GetRegexForPath(routePath)
	if err != nil {
		return false
	}
	return regex.MatchString(requestPath)
}

func routeCacheKey(request *http.Request) string {
	if request == nil || request.URL == nil {
		return ""
	}
	return request.Method + " " + request.URL.EscapedPath()
}

func pathHasMethod(pathItem *v3.PathItem, method string) bool {
	if pathItem == nil {
		return false
	}
	if method == http.MethodHead && pathItem.Head == nil && pathItem.Get != nil {
		return true
	}
	for methodName := range pathItem.GetOperations().FromOldest() {
		if strings.EqualFold(methodName, method) {
			return true
		}
	}
	return false
}

func operationMatchesMethod(pathItem *v3.PathItem, operationMethod, requestMethod string) bool {
	if strings.EqualFold(operationMethod, requestMethod) {
		return true
	}
	return requestMethod == http.MethodHead && strings.EqualFold(operationMethod, http.MethodGet) && pathItem != nil && pathItem.Head == nil
}

func routeSpecificityScore(pathTemplate string) int {
	score := 0
	for _, segment := range strings.Split(pathTemplate, "/") {
		if segment == "" {
			continue
		}
		if strings.Contains(segment, "{") && strings.Contains(segment, "}") {
			score++
			continue
		}
		score += 1000
	}
	return score
}

func ValidationRequestForRouteMatch(request *http.Request, match *RouteMatch) *http.Request {
	if request == nil || request.URL == nil || match == nil || match.BasePath == "" || match.BasePath == "/" {
		return request
	}
	if request.URL.Path != match.BasePath && !strings.HasPrefix(request.URL.Path, match.BasePath+"/") {
		return request
	}

	cloned := new(http.Request)
	*cloned = *request
	u := *request.URL
	u.Path = trimRouteBasePath(request.URL.Path, match.BasePath)
	if u.Path == "" {
		u.Path = "/"
	}
	if !strings.HasPrefix(u.Path, "/") {
		u.Path = "/" + u.Path
	}
	u.RawPath = trimRouteBaseRawPath(request.URL.EscapedPath(), match.BasePath, u.Path)
	cloned.URL = &u
	cloneRequestBody(request, cloned)
	return cloned
}

func cloneRequestBody(request, cloned *http.Request) {
	if request == nil || cloned == nil || request.GetBody == nil {
		return
	}
	body, err := request.GetBody()
	if err != nil {
		return
	}
	cloned.Body = body
	cloned.GetBody = request.GetBody
}

func trimRouteBasePath(pathValue, basePath string) string {
	if pathValue == basePath {
		return "/"
	}
	return strings.TrimPrefix(pathValue, basePath)
}

func trimRouteBaseRawPath(escapedPath, basePath, decodedPath string) string {
	escapedBasePath := escapePath(basePath)
	if escapedPath != escapedBasePath && !strings.HasPrefix(escapedPath, escapedBasePath+"/") {
		return ""
	}
	rawPath := strings.TrimPrefix(escapedPath, escapedBasePath)
	if rawPath == "" {
		rawPath = "/"
	}
	if !strings.HasPrefix(rawPath, "/") {
		rawPath = "/" + rawPath
	}
	if rawPath == escapePath(decodedPath) {
		return ""
	}
	return rawPath
}

func escapePath(pathValue string) string {
	if pathValue == "" {
		return ""
	}
	segments := strings.Split(pathValue, "/")
	for i, segment := range segments {
		segments[i] = url.PathEscape(segment)
	}
	return strings.Join(segments, "/")
}

func (r *SpecRouter) cacheResult(key string, match RouteMatch) {
	if r.cache != nil {
		r.cache.Add(key, match)
	}
}
