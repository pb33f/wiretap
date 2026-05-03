// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package report

import (
	"github.com/go-viper/mapstructure/v2"
	"github.com/pb33f/ranch/model"
	"github.com/pb33f/ranch/service"
	"github.com/pb33f/ranch/store"
	"github.com/pb33f/wiretap/daemon"
	"github.com/pb33f/wiretap/transaction"
)

const (
	ReportServiceChan     = "report"
	GenerateReportRequest = "generate-report-request"
)

type ReportService struct {
	transactionStore store.BusStore
}

type GenerateReport struct {
	Download *bool `json:"download,omitempty" mapstructure:"download"`
}

type ReportResponse struct {
	Transactions []*transaction.HttpTransaction `json:"transactions,omitempty"`
	Download     *bool                          `json:"download,omitempty"`
}

func NewReportService(storeManager store.Manager) *ReportService {
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
		var transactions []*transaction.HttpTransaction
		for x := range storeData {
			if i, k := storeData[x].(*transaction.HttpTransaction); k {
				transactions = append(transactions, i)
			}
		}
		download := true
		if r.Download != nil {
			download = *r.Download
		}
		core.SendResponse(request, &ReportResponse{
			Transactions: transactions,
			Download:     &download,
		})

	} else {
		core.SendErrorResponse(request, 400, "Invalid report request")
	}
}
