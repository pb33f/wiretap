// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package har

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/pb33f/harific/motor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHARStreamerStreamsEntriesFromPath(t *testing.T) {
	harPath := writeTestHAR(t, twoEntryHAR)

	streamer, err := NewHARStreamer(harPath, motor.StreamerOptions{WorkerCount: 1})
	require.NoError(t, err)
	require.NotNil(t, streamer)
	defer streamer.Close()

	ctx := context.Background()
	require.NoError(t, streamer.Initialize(ctx))
	require.Equal(t, 2, streamer.GetIndex().TotalEntries)

	results, err := streamer.StreamRange(ctx, 0, streamer.GetIndex().TotalEntries)
	require.NoError(t, err)

	var urls []string
	for result := range results {
		require.NoError(t, result.Error)
		require.NotNil(t, result.Entry)
		urls = append(urls, result.Entry.Request.URL)
	}

	assert.Equal(t, []string{
		"http://wiretap.local/api/pets/123",
		"http://wiretap.local/api/orders/abc",
	}, urls)
}

func TestCountHARMessagesUsesStreamerIndex(t *testing.T) {
	harPath := writeTestHAR(t, twoEntryHAR)

	count, err := CountHARMessages(harPath)

	require.NoError(t, err)
	assert.Equal(t, 4, count)
}

func TestCountHARMessagesReturnsMissingFileError(t *testing.T) {
	harPath := filepath.Join(t.TempDir(), "missing.har")

	count, err := CountHARMessages(harPath)

	assert.Equal(t, 0, count)
	require.Error(t, err)
}

func writeTestHAR(t *testing.T, content string) string {
	t.Helper()

	harPath := filepath.Join(t.TempDir(), "test.har")
	require.NoError(t, os.WriteFile(harPath, []byte(content), 0644))
	return harPath
}

const twoEntryHAR = `{
  "log": {
    "version": "1.2",
    "creator": {
      "name": "wiretap-test",
      "version": "1.0"
    },
    "entries": [
      {
        "startedDateTime": "2026-04-27T00:00:00Z",
        "time": 1,
        "request": {
          "method": "GET",
          "url": "http://wiretap.local/api/pets/123",
          "httpVersion": "HTTP/1.1",
          "cookies": [],
          "headers": [],
          "queryString": [],
          "headersSize": -1,
          "bodySize": 0
        },
        "response": {
          "status": 200,
          "statusText": "OK",
          "httpVersion": "HTTP/1.1",
          "cookies": [],
          "headers": [],
          "content": {
            "size": 0,
            "mimeType": "application/json",
            "text": "{}"
          },
          "headersSize": -1,
          "bodySize": 2
        },
        "cache": {},
        "timings": {
          "send": 0,
          "wait": 1,
          "receive": 0
        }
      },
      {
        "startedDateTime": "2026-04-27T00:00:01Z",
        "time": 1,
        "request": {
          "method": "POST",
          "url": "http://wiretap.local/api/orders/abc",
          "httpVersion": "HTTP/1.1",
          "cookies": [],
          "headers": [],
          "queryString": [],
          "headersSize": -1,
          "bodySize": 0
        },
        "response": {
          "status": 201,
          "statusText": "Created",
          "httpVersion": "HTTP/1.1",
          "cookies": [],
          "headers": [],
          "content": {
            "size": 0,
            "mimeType": "application/json",
            "text": "{}"
          },
          "headersSize": -1,
          "bodySize": 2
        },
        "cache": {},
        "timings": {
          "send": 0,
          "wait": 1,
          "receive": 0
        }
      }
    ]
  }
}`
