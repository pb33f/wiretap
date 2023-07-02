// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package config

import (
	"github.com/mitchellh/mapstructure"
	"github.com/pb33f/wiretap/shared"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"strings"
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

	viper.SetConfigType("yaml")
	verr := viper.ReadConfig(strings.NewReader(config))
	assert.NoError(t, verr)

	paths := viper.Get("paths")
	var pc map[string]*shared.WiretapPathConfig

	derr := mapstructure.Decode(paths, &pc)
	assert.NoError(t, derr)

	wcConfig := &shared.WiretapConfiguration{
		PathConfigurations: pc,
	}

	// compile paths
	wcConfig.CompilePaths()

	res := FindPaths("/pb33f/test/123", wcConfig)
	assert.Len(t, res, 1)

	res = FindPaths("/pb33f/test/123/sing/song", wcConfig)
	assert.Len(t, res, 1)

	res = FindPaths("/pb33f/no-match/wrong", wcConfig)
	assert.Len(t, res, 0)

}

func TestRewritePath(t *testing.T) {

	config := `
paths:
  /pb33f/test/**:
    target: http://localhost:9093/
    secure: false
    pathRewrite:
      '^/pb33f/test/': ''`

	viper.SetConfigType("yaml")
	verr := viper.ReadConfig(strings.NewReader(config))
	assert.NoError(t, verr)

	paths := viper.Get("paths")
	var pc map[string]*shared.WiretapPathConfig

	derr := mapstructure.Decode(paths, &pc)
	assert.NoError(t, derr)

	wcConfig := &shared.WiretapConfiguration{
		PathConfigurations: pc,
	}

	// compile paths
	wcConfig.CompilePaths()

	path := RewritePath("/pb33f/test/123/slap/a/chap", wcConfig)
	assert.Equal(t, "http://localhost:9093/123/slap/a/chap", path)

}
