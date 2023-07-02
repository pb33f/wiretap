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

	res := FindPath("/pb33f/test/123", wcConfig)
	assert.Len(t, res, 1)

	res = FindPath("/pb33f/test/123/sing/song", wcConfig)
	assert.Len(t, res, 1)

	res = FindPath("/pb33f/no-match/wrong", wcConfig)
	assert.Len(t, res, 0)

}
