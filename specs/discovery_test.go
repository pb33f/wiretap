// Copyright 2026 Princess Beef Heavy Industries LLC
// SPDX-License-Identifier: AGPL

package specs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoverSpecsFromDirectoryAndGlob(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "users.yaml"), "openapi: 3.1.0\npaths: {}\n")
	writeFile(t, filepath.Join(root, "nested", "accounts.yaml"), "openapi: 3.1.0\npaths: {}\n")
	writeFile(t, filepath.Join(root, "notes.yaml"), "not: openapi\n")

	discovered, err := DiscoverSpecs(nil, []string{root}, nil)
	require.NoError(t, err)
	require.Len(t, discovered, 2)
	assert.Equal(t, filepath.Join(root, "nested", "accounts.yaml"), discovered[0])
	assert.Equal(t, filepath.Join(root, "users.yaml"), discovered[1])

	globbed, err := DiscoverSpecs([]string{filepath.Join(root, "**", "*.yaml")}, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, discovered, globbed)
}

func TestDiscoverSpecsHonorsIgnore(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "users.yaml"), "openapi: 3.1.0\npaths: {}\n")
	writeFile(t, filepath.Join(root, "ignored", "accounts.yaml"), "openapi: 3.1.0\npaths: {}\n")

	discovered, err := DiscoverSpecs(nil, []string{root}, []string{"ignored/**"})
	require.NoError(t, err)
	require.Len(t, discovered, 1)
	assert.Equal(t, filepath.Join(root, "users.yaml"), discovered[0])
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))
}
