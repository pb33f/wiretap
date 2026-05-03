// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package controls

import (
	"testing"

	"github.com/pb33f/ranch/bus"
	"github.com/pb33f/ranch/store"
	"github.com/pb33f/wiretap/shared"
	"github.com/stretchr/testify/assert"
)

func TestResetRuntimeStateClearsTransactionAndHARState(t *testing.T) {
	storeManager := store.NewManager(bus.NewEventBus())
	controlsStore := storeManager.CreateStore(ControlServiceChan)
	transactionStore := storeManager.CreateStore(shared.WiretapServiceChan)
	harStore := storeManager.CreateStore(shared.HARServiceChan)

	config := &shared.WiretapConfiguration{GlobalAPIDelay: 250}
	controlsStore.Put(shared.ConfigKey, config, nil)
	transactionStore.Put("transaction-1", "stored transaction", nil)
	harStore.Put(shared.HARKey, "/tmp/example.har", nil)

	controlService := NewControlsService(storeManager)
	resetConfig := controlService.resetRuntimeState()

	assert.Empty(t, transactionStore.AllValues())
	assert.Empty(t, harStore.AllValues())
	assert.Equal(t, 0, resetConfig.GlobalAPIDelay)
	assert.Equal(t, 0, config.GlobalAPIDelay)
}
