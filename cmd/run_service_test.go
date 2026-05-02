// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package cmd

import (
	"os"
	"testing"

	"github.com/pb33f/wiretap/shared"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunWiretapServiceRejectsInvalidWebSocketPort(t *testing.T) {
	resetPFlagCommandLine(t)

	_, err := runWiretapService(&shared.WiretapConfiguration{
		WebSocketPort: "not-a-port",
	}, nil, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), `parse websocket port "not-a-port"`)
}

func TestRootCommandRejectsCLIMockModeWithoutSpec(t *testing.T) {
	err := executeTestRootCommand(t, "--mock-mode")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot enable mock mode")
}

func TestRootCommandRejectsConfigMockModeWithoutSpec(t *testing.T) {
	configFile := writeTestConfig(t, "mockMode: true\n")

	err := executeTestRootCommand(t, "--config", configFile)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot enable mock mode")
}

func TestRootCommandPropagatesStartupErrors(t *testing.T) {
	err := executeTestRootCommand(t, "--url", "http://example.com", "--ws-port", "not-a-port")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot start wiretap")
	assert.Contains(t, err.Error(), `parse websocket port "not-a-port"`)
}

func TestRootCommandRejectsInvalidRedirectURL(t *testing.T) {
	err := executeTestRootCommand(t, "--url", "::::")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid redirect URL")
}

func TestRootCommandFailsNormalRunWhenSpecLoadFails(t *testing.T) {
	err := executeTestRootCommand(t, "--spec", "missing.yaml", "--url", "http://example.com")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load 1 OpenAPI specification")
	assert.Contains(t, err.Error(), "missing.yaml")
}

func TestRootCommandDryRunSkipsInvalidRedirectURLParsing(t *testing.T) {
	tmpDir := t.TempDir()
	users := tmpDir + "/users.yaml"
	accounts := tmpDir + "/accounts.yaml"
	writeTestSpecWithOperation(t, users, "/health", "healthUsers")
	writeTestSpecWithOperation(t, accounts, "/health", "healthAccounts")

	err := executeTestRootCommand(t, "--dry-run", "--url", "::::", "--specs", users, "--specs", accounts)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "dry run failed")
	assert.Contains(t, err.Error(), "detected 1 conflicts")
	assert.NotContains(t, err.Error(), "invalid redirect URL")
}

func TestRootCommandDryRunIgnoresClashingOperationIDWhenFlagged(t *testing.T) {
	tmpDir := t.TempDir()
	users := tmpDir + "/users.yaml"
	accounts := tmpDir + "/accounts.yaml"
	writeTestSpecWithOperation(t, users, "/users", "listThings")
	writeTestSpecWithOperation(t, accounts, "/accounts", "listThings")

	err := executeTestRootCommand(t, "--dry-run", "--specs", users, "--specs", accounts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "detected 1 conflicts")

	err = executeTestRootCommand(t, "--dry-run", "--ignore-clashing-operationid", "--specs", users, "--specs", accounts)
	require.NoError(t, err)
}

func TestResolvePrimarySpecPreservesConfiguredContract(t *testing.T) {
	tmpDir := t.TempDir()
	first := tmpDir + "/first.yaml"
	primary := tmpDir + "/primary.yaml"
	writeTestSpec(t, first)
	writeTestSpec(t, primary)

	discovered, err := resolvePrimarySpec("", []string{first, primary}, nil)
	require.NoError(t, err)
	assert.Equal(t, first, discovered)

	resolved, err := resolvePrimarySpec(primary, []string{first, primary}, nil)
	require.NoError(t, err)
	assert.Equal(t, primary, resolved)
}

func executeTestRootCommand(t *testing.T, args ...string) error {
	t.Helper()
	resetPFlagCommandLine(t)

	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(originalDir))
	})

	cmd := &cobra.Command{
		SilenceErrors: true,
		SilenceUsage:  true,
		Use:           rootCmd.Use,
		Short:         rootCmd.Short,
		Long:          rootCmd.Long,
		RunE:          rootCmd.RunE,
	}
	registerRootFlags(cmd)

	if err := cmd.ParseFlags(args); err != nil {
		return err
	}
	return cmd.RunE(cmd, args)
}

func writeTestSpec(t *testing.T, path string) {
	t.Helper()

	require.NoError(t, os.WriteFile(path, []byte(`openapi: 3.1.0
info:
  title: test
  version: "1.0"
paths: {}
`), 0644))
}

func writeTestSpecWithOperation(t *testing.T, path, pathName, operationID string) {
	t.Helper()

	require.NoError(t, os.WriteFile(path, []byte(`openapi: 3.1.0
info:
  title: test
  version: "1.0"
paths:
  "`+pathName+`":
    get:
      operationId: `+operationID+`
      responses:
        "200":
          description: ok
`), 0644))
}

func writeTestConfig(t *testing.T, contents string) string {
	t.Helper()

	path := t.TempDir() + "/wiretap.yaml"
	require.NoError(t, os.WriteFile(path, []byte(contents), 0644))
	return path
}

func resetPFlagCommandLine(t *testing.T) {
	t.Helper()

	originalCommandLine := pflag.CommandLine
	pflag.CommandLine = pflag.NewFlagSet("", pflag.ExitOnError)
	t.Cleanup(func() {
		pflag.CommandLine = originalCommandLine
	})
}
