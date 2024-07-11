// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package config

import (
	"encoding/json"
	"github.com/pb33f/wiretap/shared"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"net/http"
	"testing"
)

func TestFindPath(t *testing.T) {

	config := `
paths:
  /pb33f/test/**:
    target: /
    secure: false
    pathRewrite:
      '^/pb33f/test/': ''`

	var wcConfig shared.WiretapConfiguration
	_ = yaml.Unmarshal([]byte(config), &wcConfig)

	wcConfig.CompilePaths()

	res := FindPaths("/pb33f/test/123", &wcConfig)
	assert.Len(t, res, 1)

	res = FindPaths("/pb33f/test/123/sing/song", &wcConfig)
	assert.Len(t, res, 1)

	res = FindPaths("/pb33f/no-match/wrong", &wcConfig)
	assert.Len(t, res, 0)

}

func TestFindPath_JSON(t *testing.T) {

	config := `
{
    "paths": {
		"/pb33f/test/**": {
			"target": "/",
			"secure": false,
			"pathRewrite": {
				"^/pb33f/test/": ""
			}
		}
	}
}`

	var wcConfig shared.WiretapConfiguration
	_ = json.Unmarshal([]byte(config), &wcConfig)

	wcConfig.CompilePaths()

	res := FindPaths("/pb33f/test/123", &wcConfig)
	assert.Len(t, res, 1)

	res = FindPaths("/pb33f/test/123/sing/song", &wcConfig)
	assert.Len(t, res, 1)

	res = FindPaths("/pb33f/no-match/wrong", &wcConfig)
	assert.Len(t, res, 0)

}

func TestRewritePath(t *testing.T) {

	config := `
paths:
  /pb33f/test/**:
    target: localhost:9093/
    secure: false
    pathRewrite:
      '^/pb33f/test/': ''`

	var wcConfig shared.WiretapConfiguration
	_ = yaml.Unmarshal([]byte(config), &wcConfig)

	wcConfig.CompilePaths()

	path := RewritePath("/pb33f/test/123/slap/a/chap", nil, &wcConfig)
	assert.Equal(t, "http://localhost:9093/123/slap/a/chap", path.RewrittenPath)

}

func TestRewritePath_Secure(t *testing.T) {

	config := `
paths:
  /pb33f/*/test/**:
    target: localhost:9093
    secure: true
    pathRewrite:
      '^/pb33f/(\w+)/test/': '/flat/jam/'`

	var wcConfig shared.WiretapConfiguration
	_ = yaml.Unmarshal([]byte(config), &wcConfig)

	wcConfig.CompilePaths()

	path := RewritePath("/pb33f/cakes/test/123/smelly/jelly", nil, &wcConfig)
	assert.Equal(t, "https://localhost:9093/flat/jam/123/smelly/jelly", path.RewrittenPath)

}

func TestRewritePath_Secure_With_Variables(t *testing.T) {

	config := `
paths:
  /pb33f/*/test/*/321/**:
    target: localhost:9093
    secure: true
    pathRewrite:
      '^/pb33f/(\w+)/test/(\w+)/(\d+)/': '/slippy/$1/whip/$3/$2/'`

	var wcConfig shared.WiretapConfiguration
	_ = yaml.Unmarshal([]byte(config), &wcConfig)

	wcConfig.CompilePaths()

	path := RewritePath("/pb33f/cakes/test/lemons/321/smelly/jelly", nil, &wcConfig)
	assert.Equal(t, "https://localhost:9093/slippy/cakes/whip/321/lemons/smelly/jelly", path.RewrittenPath)

}

func TestRewritePath_Secure_With_Variables_CaseSensitive(t *testing.T) {

	config := `
paths:
  /en-US/burgerd/__raw/*:
    target: localhost:80
    pathRewrite:
      '^/en-US/burgerd/__raw/(\w+)/nobody/': '$1/-/'
  /en-US/burgerd/services/*:
    target: locahost:80
    pathRewrite:
      '^/en-US/burgerd/services': '/services'`

	var c shared.WiretapConfiguration
	_ = yaml.Unmarshal([]byte(config), &c)

	c.CompilePaths()

	path := RewritePath("/en-US/burgerd/__raw/noKetchupPlease/nobody/", nil, &c)
	assert.Equal(t, "http://localhost:80/noKetchupPlease/-/", path.RewrittenPath)

}

func TestRewritePath_Secure_With_Variables_CaseSensitive_AndQuery(t *testing.T) {

	config := `
paths:
  /en-US/burgerd/__raw/*:
    target: localhost:80
    pathRewrite:
      '^/en-US/burgerd/__raw/(\w+)/nobody/': '$1/-/'
  /en-US/burgerd/services/*:
    target: locahost:80
    pathRewrite:
      '^/en-US/burgerd/services': '/services'`

	var c shared.WiretapConfiguration
	_ = yaml.Unmarshal([]byte(config), &c)

	c.CompilePaths()

	path := RewritePath("/en-US/burgerd/__raw/noKetchupPlease/nobody/yummy/yum?onions=true", nil, &c)
	assert.Equal(t, "http://localhost:80/noKetchupPlease/-/yummy/yum?onions=true", path.RewrittenPath)

}

func TestRewritePath_With_RewriteId(t *testing.T) {

	config := `
paths:
  /pb33f/test/**:
    target: localhost:9093
    secure: true
    pathRewrite:
      'test': 'invalid'
  /pb33f/test/id:
    target: localhost:9093
    secure: true
    rewriteId: test_id
    pathRewrite:
      'test': 'correct'`

	var c shared.WiretapConfiguration
	_ = yaml.Unmarshal([]byte(config), &c)

	c.CompilePaths()

	req := &http.Request{
		Header: http.Header{
			"RewriteId":    []string{"garbage", "garbage again", "test_id", "gargabe please"},
			"Other-Header": []string{"another header"},
		},
	}

	pathConfigs := make([]*shared.WiretapPathConfig, 0, c.PathConfigurations.Len())

	for x := c.PathConfigurations.First(); x != nil; x = x.Next() {
		pathConfigs = append(pathConfigs, x.Value())
	}

	path := RewritePath("/pb33f/test/id", req, &c)
	assert.Equal(t, "https://localhost:9093/pb33f/correct/id", path.RewrittenPath)

	actualConfig := FindPathWithRewriteId(pathConfigs, req)
	expectedConfig := pathConfigs[1] // second config is the valid one
	assert.Equal(t, expectedConfig, actualConfig)

}

func TestRewritePath_With_RewriteId_No_Header(t *testing.T) {

	config := `
paths:
  /pb33f/test/**:
    target: localhost:9093
    secure: true
    pathRewrite:
      'test': 'correct'
  /pb33f/test/id:
    target: localhost:9093
    secure: true
    rewriteId: test_id
    pathRewrite:
      'test': 'invalid'`

	var c shared.WiretapConfiguration
	_ = yaml.Unmarshal([]byte(config), &c)

	c.CompilePaths()

	req := &http.Request{
		Header: http.Header{},
	}

	path := RewritePath("/pb33f/test/id", req, &c)
	assert.Equal(t, "https://localhost:9093/pb33f/correct/id", path.RewrittenPath)

}

func TestRewritePath_With_RewriteId_No_Valid_Rewrite_Id(t *testing.T) {

	config := `
paths:
  /pb33f/test/**:
    target: localhost:9093
    secure: true
    pathRewrite:
      'test': 'correct'
  /pb33f/test/id:
    target: localhost:9093
    secure: true
    rewriteId: test_id
    pathRewrite:
      'test': 'invalid'`

	var c shared.WiretapConfiguration
	_ = yaml.Unmarshal([]byte(config), &c)

	c.CompilePaths()

	req := &http.Request{
		Header: http.Header{
			"RewriteId":    []string{"garbage", "garbage again", "gargabe please"},
			"Other-Header": []string{"another header"},
		},
	}

	path := RewritePath("/pb33f/test/id", req, &c)
	assert.Equal(t, "https://localhost:9093/pb33f/correct/id", path.RewrittenPath)

}

func TestLocatePathDelay(t *testing.T) {

	config := `pathDelays:
  /pb33f/test/**: 1000
  /pb33f/cakes/123: 2000
  /*/test/123: 3000`

	var c shared.WiretapConfiguration
	_ = yaml.Unmarshal([]byte(config), &c)

	c.CompilePathDelays()

	delay := FindPathDelay("/pb33f/test/burgers/fries?1234=no", &c)
	assert.Equal(t, 1000, delay)

	delay = FindPathDelay("/pb33f/cakes/123", &c)
	assert.Equal(t, 2000, delay)

	delay = FindPathDelay("/roastbeef/test/123", &c)
	assert.Equal(t, 3000, delay)

	delay = FindPathDelay("/not-registered", &c)
	assert.Equal(t, 0, delay)

}

func TestIgnoreRedirect(t *testing.T) {

	config := `ignoreRedirects:
  - /pb33f/test/**
  - /pb33f/cakes/123
  - /*/test/123`

	var c shared.WiretapConfiguration
	_ = yaml.Unmarshal([]byte(config), &c)

	c.CompileIgnoreRedirects()

	ignore := IgnoreRedirectOnPath("/pb33f/test/burgers/fries?1234=no", &c)
	assert.True(t, ignore)

	ignore = IgnoreRedirectOnPath("/pb33f/cakes/123", &c)
	assert.True(t, ignore)

	ignore = IgnoreRedirectOnPath("/roastbeef/test/123", &c)
	assert.True(t, ignore)

	ignore = IgnoreRedirectOnPath("/not-registered", &c)
	assert.False(t, ignore)

}

func TestIgnoreRedirect_NoPathsRegistered(t *testing.T) {

	var c shared.WiretapConfiguration
	_ = yaml.Unmarshal([]byte(""), &c)

	c.CompileIgnoreRedirects()

	ignore := IgnoreRedirectOnPath("/pb33f/test/burgers/fries?1234=no", &c)
	assert.False(t, ignore)

	ignore = IgnoreRedirectOnPath("/pb33f/cakes/123", &c)
	assert.False(t, ignore)

	ignore = IgnoreRedirectOnPath("/roastbeef/test/123", &c)
	assert.False(t, ignore)

	ignore = IgnoreRedirectOnPath("/not-registered", &c)
	assert.False(t, ignore)

}

func TestRedirectAllowList(t *testing.T) {

	config := `redirectAllowList:
  - /pb33f/test/**
  - /pb33f/cakes/123
  - /*/test/123`

	var c shared.WiretapConfiguration
	_ = yaml.Unmarshal([]byte(config), &c)

	c.CompileRedirectAllowList()

	ignore := PathRedirectAllowListed("/pb33f/test/burgers/fries?1234=no", &c)
	assert.True(t, ignore)

	ignore = PathRedirectAllowListed("/pb33f/cakes/123", &c)
	assert.True(t, ignore)

	ignore = PathRedirectAllowListed("/roastbeef/test/123", &c)
	assert.True(t, ignore)

	ignore = PathRedirectAllowListed("/not-registered", &c)
	assert.False(t, ignore)

}

func TestRedirectAllowList_NoPathsRegistered(t *testing.T) {

	var c shared.WiretapConfiguration
	_ = yaml.Unmarshal([]byte(""), &c)

	c.CompileRedirectAllowList()

	ignore := PathRedirectAllowListed("/pb33f/test/burgers/fries?1234=no", &c)
	assert.False(t, ignore)

	ignore = PathRedirectAllowListed("/pb33f/cakes/123", &c)
	assert.False(t, ignore)

	ignore = PathRedirectAllowListed("/roastbeef/test/123", &c)
	assert.False(t, ignore)

	ignore = PathRedirectAllowListed("/not-registered", &c)
	assert.False(t, ignore)

}

func TestIgnoreValidation(t *testing.T) {

	config := `ignoreValidation:
  - /pb33f/test/**
  - /pb33f/cakes/123
  - /*/test/123`

	var c shared.WiretapConfiguration
	_ = yaml.Unmarshal([]byte(config), &c)

	c.CompileIgnoreValidations()

	ignore := IgnoreValidationOnPath("/pb33f/test/burgers/fries?1234=no", &c)
	assert.True(t, ignore)

	ignore = IgnoreValidationOnPath("/pb33f/cakes/123", &c)
	assert.True(t, ignore)

	ignore = IgnoreValidationOnPath("/roastbeef/test/123", &c)
	assert.True(t, ignore)

	ignore = IgnoreValidationOnPath("/not-registered", &c)
	assert.False(t, ignore)

}

func TestIgnoreValidation_NoPathsRegistered(t *testing.T) {

	var c shared.WiretapConfiguration
	_ = yaml.Unmarshal([]byte(""), &c)

	c.CompileIgnoreValidations()

	ignore := IgnoreValidationOnPath("/pb33f/test/burgers/fries?1234=no", &c)
	assert.False(t, ignore)

	ignore = IgnoreValidationOnPath("/pb33f/cakes/123", &c)
	assert.False(t, ignore)

	ignore = IgnoreValidationOnPath("/roastbeef/test/123", &c)
	assert.False(t, ignore)

	ignore = IgnoreValidationOnPath("/not-registered", &c)
	assert.False(t, ignore)

}

func TestValidationAllowList(t *testing.T) {

	config := `validationAllowList:
  - /pb33f/test/**
  - /pb33f/cakes/123
  - /*/test/123`

	var c shared.WiretapConfiguration
	_ = yaml.Unmarshal([]byte(config), &c)

	c.CompileValidationAllowList()

	ignore := PathValidationAllowListed("/pb33f/test/burgers/fries?1234=no", &c)
	assert.True(t, ignore)

	ignore = PathValidationAllowListed("/pb33f/cakes/123", &c)
	assert.True(t, ignore)

	ignore = PathValidationAllowListed("/roastbeef/test/123", &c)
	assert.True(t, ignore)

	ignore = PathValidationAllowListed("/not-registered", &c)
	assert.False(t, ignore)

}

func TestValidationAllowList_NoPathsRegistered(t *testing.T) {

	var c shared.WiretapConfiguration
	_ = yaml.Unmarshal([]byte(""), &c)

	c.CompileValidationAllowList()

	ignore := PathValidationAllowListed("/pb33f/test/burgers/fries?1234=no", &c)
	assert.False(t, ignore)

	ignore = PathValidationAllowListed("/pb33f/cakes/123", &c)
	assert.False(t, ignore)

	ignore = PathValidationAllowListed("/roastbeef/test/123", &c)
	assert.False(t, ignore)

	ignore = PathValidationAllowListed("/not-registered", &c)
	assert.False(t, ignore)

}

func TestGetRewriteHeaderValues(t *testing.T) {

	expectedValue := []string{"ExpectedValue"}

	requestList := []*http.Request{
		{
			Header: http.Header{
				"Rewriteid":    expectedValue,
				"Other-Header": []string{"another header"},
				"other-header": []string{"another another header"},
			},
		},
		{
			Header: http.Header{
				"Rewrite-Id":   expectedValue,
				"Other-Header": []string{"another header"},
				"other-header": []string{"another another header"},
			},
		},
		{
			Header: http.Header{
				"Rewrite_id":   expectedValue,
				"Other-Header": []string{"another header"},
				"other-header": []string{"another another header"},
			},
		},
		{
			Header: http.Header{
				"RewriteId":    expectedValue,
				"Other-Header": []string{"another header"},
				"other-header": []string{"another another header"},
			},
		},
		{
			Header: http.Header{
				"RewrIte-Id":   expectedValue,
				"Other-Header": []string{"another header"},
				"other-header": []string{"another another header"},
			},
		},
		{
			Header: http.Header{
				"rewriteid":    expectedValue,
				"Other-Header": []string{"another header"},
				"other-header": []string{"another another header"},
			},
		},
		{
			Header: http.Header{
				"rewrite-id":   expectedValue,
				"Other-Header": []string{"another header"},
				"other-header": []string{"another another header"},
			},
		},
		{
			Header: http.Header{
				"rewrite_id":   expectedValue,
				"Other-Header": []string{"another header"},
				"other-header": []string{"another another header"},
			},
		},
	}

	for _, request := range requestList {
		actualValue, found := getRewriteIdHeaderValues(request)
		assert.Equal(t, expectedValue, actualValue)
		assert.True(t, found)
	}

}

func TestGetRewriteHeaderValues_MissingHeader(t *testing.T) {

	requestList := []*http.Request{
		{
			Header: http.Header{
				"Other-Header": []string{"another header"},
				"other-header": []string{"another another header"},
			},
		},
		{
			Header: http.Header{},
		},
	}

	for _, request := range requestList {
		actualValue, found := getRewriteIdHeaderValues(request)
		assert.Equal(t, []string{}, actualValue)
		assert.False(t, found)
	}

}
