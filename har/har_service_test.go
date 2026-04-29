// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package har

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/pb33f/ranch/bus"
	"github.com/pb33f/ranch/store"
	"github.com/stretchr/testify/assert"
)

func TestNewHARServiceUsesDefaultReplayDelay(t *testing.T) {
	service := NewHARService(nil, slog.New(slog.NewTextHandler(io.Discard, nil)), 0, newTestStoreManager())

	assert.Equal(t, 10*time.Millisecond, service.replayDelay)
}

func TestNewHARServiceUsesConfiguredReplayDelay(t *testing.T) {
	service := NewHARService(nil, slog.New(slog.NewTextHandler(io.Discard, nil)), 25, newTestStoreManager())

	assert.Equal(t, 25*time.Millisecond, service.replayDelay)
}

func newTestStoreManager() store.Manager {
	return store.NewManager(bus.NewEventBus())
}
