// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package specs

import (
	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/ranch/model"
	"github.com/pb33f/ranch/service"
)

const (
	SpecServiceChan       = "specs"
	GetCurrentSpecRequest = "get-current-spec"
)

type SpecService struct {
	document    libopenapi.Document
	docModel    *v3.Document
	serviceCore service.FabricServiceCore
}

func NewSpecService(document libopenapi.Document) *SpecService {
	m, _ := document.BuildV3Model()
	return &SpecService{

		document: document,
		docModel: &m.Model,
	}
}

func (ss *SpecService) HandleServiceRequest(request *model.Request, core service.FabricServiceCore) {
	switch request.RequestCommand {
	case GetCurrentSpecRequest:
		ss.handleGetCurrentSpec(request, core)
	default:
		core.HandleUnknownRequest(request)
	}
}

func (ss *SpecService) handleGetCurrentSpec(request *model.Request, core service.FabricServiceCore) {
	core.SendResponse(request, ss.document.GetSpecInfo().SpecBytes)
}
