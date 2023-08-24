// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package report

import (
	"github.com/mitchellh/mapstructure"
	"github.com/pb33f/ranch/bus"
	"github.com/pb33f/ranch/model"
	"github.com/pb33f/ranch/service"
	"github.com/pb33f/wiretap/daemon"
)

const (
	ReportServiceChan     = "report"
	GenerateReportRequest = "generate-report-request"
)

type ReportService struct {
	transactionStore bus.BusStore
}

type GenerateReport struct {
}

type ReportResponse struct {
	Transactions []*daemon.HttpTransaction `json:"transactions,omitempty"`
}

func NewReportService() *ReportService {
	storeManager := bus.GetBus().GetStoreManager()
	transactionStore := storeManager.GetStore(daemon.WiretapServiceChan)
	return &ReportService{
		transactionStore: transactionStore,
	}
}

func (rs *ReportService) HandleServiceRequest(request *model.Request, core service.FabricServiceCore) {
	switch request.RequestCommand {
	case GenerateReportRequest:
		rs.buildReport(request, core)
	default:
		core.HandleUnknownRequest(request)
	}
}

func (rs *ReportService) buildReport(request *model.Request, core service.FabricServiceCore) {

	if dl, ok := request.Payload.(map[string]interface{}); ok {

		// decode the object into a request
		var r GenerateReport
		_ = mapstructure.Decode(dl, &r)

		// extract state from store.
		storeData := rs.transactionStore.AllValues()
		var transactions []*daemon.HttpTransaction
		for x := range storeData {
			if i, k := storeData[x].(*daemon.HttpTransaction); k {
				transactions = append(transactions, i)
			}
		}
		core.SendResponse(request, &ReportResponse{transactions})

	} else {
		core.SendErrorResponse(request, 400, "Invalid report request")
	}
}
