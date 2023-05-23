// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package daemon

import (
    "github.com/pb33f/ranch/model"
    "github.com/pb33f/ranch/service"
)

func (ws *WiretapService) changeDelay(request *model.Request, core service.FabricServiceCore) {

    if dl, ok := request.Payload.(int); ok {
        ws.config.GlobalAPIDelay = dl
        core.SendResponse(request, dl)
    } else {
        core.SendErrorResponse(request, 400, "Invalid delay value")
    }
}
