// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package daemon

import "github.com/pb33f/ranch/service"

func (ws *WiretapService) Init(core service.FabricServiceCore) error {
    ws.serviceCore = core
    eventBus := core.Bus()

    // create broadcast channel and set it to galactic
    channel := eventBus.GetChannelManager().CreateChannel(WiretapBroadcastChan)
    channel.SetGalactic(WiretapBroadcastChan)

    ws.broadcastChan = channel
    ws.bus = eventBus
    core.SetDefaultJSONHeaders()
    return nil
}
