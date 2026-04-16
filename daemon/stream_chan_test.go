// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package daemon

import (
	"testing"
	"time"

	"github.com/pb33f/wiretap/shared"
	"github.com/stretchr/testify/assert"
)

// TestSendToStreamChan_DoesNotBlockWhenFull proves that the sync hard-error
// path cannot deadlock on a stalled stream consumer. Previously streamChan was
// unbuffered and the direct send (`ws.streamChan <- errs`) would block the
// proxy goroutine forever if listenForValidationErrors had not started or had
// stalled writing to disk.
func TestSendToStreamChan_DoesNotBlockWhenFull(t *testing.T) {
	ws := &WiretapService{
		// tiny buffer + no consumer; any second send would block if we did a
		// naive channel send
		streamChan: make(chan []*shared.WiretapValidationError, 1),
	}

	done := make(chan struct{})
	go func() {
		for i := 0; i < 10; i++ {
			sendToStreamChan(ws, buildSampleErrors("boom"))
		}
		close(done)
	}()

	select {
	case <-done:
		// pass — all sends returned even though only the first filled the buffer
		// and there was no consumer
	case <-time.After(500 * time.Millisecond):
		t.Fatal("sendToStreamChan blocked with no consumer — sync path would deadlock under real load")
	}
}

func TestSendToStreamChan_DeliversWhenBufferAvailable(t *testing.T) {
	ws := &WiretapService{
		streamChan: make(chan []*shared.WiretapValidationError, 4),
	}

	sendToStreamChan(ws, buildSampleErrors("a"))
	sendToStreamChan(ws, buildSampleErrors("b"))

	assert.Equal(t, 2, len(ws.streamChan), "both sends should land in the buffer")
}
