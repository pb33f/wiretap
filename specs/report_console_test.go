// Copyright 2026 Princess Beef Heavy Industries LLC
// SPDX-License-Identifier: AGPL

package specs

import (
	"bytes"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderConsoleUsesReadableConflictBlocks(t *testing.T) {
	tmpDir := t.TempDir()
	specA := filepath.Join(tmpDir, "api", "accounts.yaml")
	specB := filepath.Join(tmpDir, "rulesets", "testdata", "accounts-shadow.yaml")

	report := &ConflictReport{
		Conflicts: []Conflict{
			{
				Kind:       KindCrossSpecDuplicate,
				Method:     "GET",
				Paths:      []string{"/api/{service}/health", "/api/{service}/health"},
				RoutePaths: []string{"/api/{service}/health", "/api/{service}/health"},
				Specs:      []string{specA, specB},
			},
		},
		LoadErrors: []LoadError{
			{Spec: specA, Error: fmt.Errorf("stat %s: no such file or directory", specA)},
		},
		SpecCount: 2,
	}

	var out bytes.Buffer
	RenderConsole(report, &out)
	rendered := out.String()

	assert.Contains(t, rendered, "Cross-spec duplicate routes (1)")
	assert.Contains(t, rendered, "GET /api/{service}/health")
	assert.Contains(t, rendered, "api/accounts.yaml")
	assert.Contains(t, rendered, "rulesets/testdata/accounts-shadow.yaml")
	assert.Contains(t, rendered, "Load errors (1)")
	assert.NotContains(t, rendered, tmpDir)
	assert.NotContains(t, rendered, "METHOD\tROUTES\tSPECS")
}

func TestRenderConsoleShowsOperationIDIgnoreHint(t *testing.T) {
	report := &ConflictReport{
		Conflicts: []Conflict{
			{
				Kind:        KindDuplicateOperationID,
				Paths:       []string{"GET /health", "GET /health"},
				Specs:       []string{"api/accounts.yaml", "api/users.yaml"},
				OperationID: "getHealth",
			},
		},
		SpecCount: 2,
	}

	var out bytes.Buffer
	RenderConsole(report, &out)
	rendered := out.String()

	assert.Contains(t, rendered, "Duplicate operationIds (1)")
	assert.Contains(t, rendered, "Use --ignore-clashing-operationid to ignore duplicate operation IDs clashing across specs.")
}

func TestConsoleStyleHighlightsConflictDetails(t *testing.T) {
	var out bytes.Buffer
	style := consoleStyle{enabled: true}

	renderDetailLines([]detailLine{
		{Label: "spec", Value: "api/accounts.yaml", Kind: detailSpec},
		{Label: "conflict path", Value: "/health", Kind: detailPath},
	}, style, &out)

	rendered := out.String()
	assert.Contains(t, rendered, "├─")
	assert.Contains(t, rendered, "└─")
	assert.Contains(t, rendered, ansiDarkGrey+"spec:"+ansiReset)
	assert.Contains(t, rendered, ansiSecondaryPink+"api/accounts.yaml"+ansiReset)
	assert.NotContains(t, rendered, ansiHeaderPink+"api/accounts.yaml"+ansiReset)
	assert.NotContains(t, rendered, "\033[1;38;2;248;58;255mapi/accounts.yaml")
	assert.Contains(t, rendered, ansiPrimaryCyan+"/health"+ansiReset)
	assert.NotContains(t, rendered, "`-")
}

func TestConsoleStyleUsesDoctorMethodColors(t *testing.T) {
	style := consoleStyle{enabled: true}

	assert.Equal(t, ansiMethodGreen+"GET"+ansiReset, style.method("GET"))
	assert.Equal(t, ansiMethodGreen+"QUERY"+ansiReset, style.method("QUERY"))
	assert.Equal(t, ansiMethodYellow+"PATCH"+ansiReset, style.method("PATCH"))
	assert.Equal(t, ansiMethodBlue+"PUT"+ansiReset, style.method("PUT"))
	assert.Equal(t, ansiMethodBlue+"POST"+ansiReset, style.method("POST"))
	assert.Equal(t, ansiMethodRed+"DELETE"+ansiReset, style.method("DELETE"))
	assert.Equal(t, "OPTIONS", style.method("OPTIONS"))

	rendered := style.operationPath("POST /widgets")
	assert.Contains(t, rendered, ansiMethodBlue+"POST"+ansiReset)
	assert.Contains(t, rendered, ansiPrimaryCyan+"/widgets"+ansiReset)
}
