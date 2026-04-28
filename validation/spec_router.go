// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package validation

import (
	"net/http"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/pb33f/libopenapi-validator/paths"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

const specRouterCacheSize = 256

type DocumentValidator struct {
	DocumentName string
	DocModel     *v3.Document
	Validator    HttpValidator
}

type SpecRouter struct {
	validators []DocumentValidator
	cache      *lru.Cache[string, int]
}

func NewSpecRouter(validators []DocumentValidator) *SpecRouter {
	docs := make([]DocumentValidator, len(validators))
	copy(docs, validators)

	router := &SpecRouter{validators: docs}
	if len(docs) > 1 {
		cache, _ := lru.New[string, int](specRouterCacheSize)
		router.cache = cache
	}
	return router
}

func (r *SpecRouter) Resolve(request *http.Request) *DocumentValidator {
	_, doc := r.ResolveIndex(request)
	return doc
}

func (r *SpecRouter) ResolveIndex(request *http.Request) (int, *DocumentValidator) {
	if r == nil || len(r.validators) == 0 || request == nil {
		return -1, nil
	}
	if len(r.validators) == 1 {
		return 0, &r.validators[0]
	}
	if request.URL == nil {
		return 0, &r.validators[0]
	}

	key := request.Method + " " + request.URL.Path
	if r.cache != nil {
		if index, ok := r.cache.Get(key); ok && index >= 0 && index < len(r.validators) {
			return index, &r.validators[index]
		}
	}

	for i := range r.validators {
		pathItem, _, _ := paths.FindPath(request, r.validators[i].DocModel, nil)
		if pathItem != nil {
			r.cacheResult(key, i)
			return i, &r.validators[i]
		}
	}

	r.cacheResult(key, 0)
	return 0, &r.validators[0]
}

func (r *SpecRouter) cacheResult(key string, index int) {
	if r.cache != nil {
		r.cache.Add(key, index)
	}
}
