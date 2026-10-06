// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package shared

import (
	"testing"

	"github.com/pb33f/go-yaml"
	"github.com/pb33f/testify/assert"
	"github.com/pb33f/testify/require"
)

func TestWiretapConfigurationYAMLPreservesPathOrderAndAliases(t *testing.T) {
	const source = `paths:
  /products/special: &upstream
    target: api.example.com
    secure: true
  /products/**: *upstream
mockMode: true
`
	var config WiretapConfiguration
	require.NoError(t, yaml.Unmarshal([]byte(source), &config))
	require.NotNil(t, config.PathConfigurations)
	require.NotNil(t, config.CompiledPaths)
	assert.True(t, config.MockMode)

	var paths []string
	for pair := config.PathConfigurations.First(); pair != nil; pair = pair.Next() {
		paths = append(paths, pair.Key())
		require.NotNil(t, pair.Value())
		assert.Equal(t, "api.example.com", pair.Value().Target)
		assert.True(t, pair.Value().Secure)
	}
	assert.Equal(t, []string{"/products/special", "/products/**"}, paths)

	config.CompilePaths()
	var compiledPaths []string
	for pair := config.CompiledPaths.First(); pair != nil; pair = pair.Next() {
		compiledPaths = append(compiledPaths, pair.Key())
		assert.True(t, pair.Value().CompiledKey.Match("/products/special"))
	}
	assert.Equal(t, paths, compiledPaths)
}

func TestWiretapConfigurationYAMLRejectsInvalidFieldTypes(t *testing.T) {
	for _, source := range []string{
		"mockMode: invalid\n",
		"paths:\n  /products:\n    secure: invalid\n",
	} {
		t.Run(source, func(t *testing.T) {
			var config WiretapConfiguration
			require.Error(t, yaml.Unmarshal([]byte(source), &config))
		})
	}
}
