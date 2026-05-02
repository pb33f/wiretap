// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package daemon

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloneExistingRequestInstallsReplayableBody(t *testing.T) {
	body := []byte(`{"name":"widget"}`)
	req, err := http.NewRequest(http.MethodPost, "http://wiretap.local/widgets", bytes.NewReader(body))
	require.NoError(t, err)

	cloned := CloneExistingRequest(CloneRequest{
		Request:   req,
		Protocol:  "http",
		Host:      "upstream.local",
		BodyBytes: body,
	})

	require.NotNil(t, cloned)
	require.NotNil(t, cloned.GetBody)
	require.NotNil(t, req.GetBody)
	assert.Equal(t, int64(len(body)), cloned.ContentLength)
	assert.Equal(t, int64(len(body)), req.ContentLength)

	clonedBody, err := io.ReadAll(cloned.Body)
	require.NoError(t, err)
	assert.Equal(t, body, clonedBody)

	replayed, err := cloned.GetBody()
	require.NoError(t, err)
	replayedBody, err := io.ReadAll(replayed)
	require.NoError(t, err)
	assert.Equal(t, body, replayedBody)

	originalReplay, err := req.GetBody()
	require.NoError(t, err)
	originalReplayBody, err := io.ReadAll(originalReplay)
	require.NoError(t, err)
	assert.Equal(t, body, originalReplayBody)
}
