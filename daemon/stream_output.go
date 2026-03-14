// Copyright 2024 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package daemon

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/pb33f/wiretap/shared"
	"github.com/pterm/pterm"
	"os"
	"sync"
)

func (ws *WiretapService) listenForValidationErrors() {

	ws.streamViolations = []*shared.WiretapValidationError{}
	var lock sync.RWMutex
	json := jsoniter.ConfigCompatibleWithStandardLibrary

	_ = os.Remove(ws.reportFile)
	f, err := os.OpenFile(ws.reportFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		pterm.Error.Println("cannot stream violations: " + err.Error())
		return
	}

	go func() {
		defer f.Close()
		for {
			select {
			case violations := <-ws.streamChan:
				if ws.stream {
					lock.Lock()
					ws.streamViolations = append(ws.streamViolations, violations...)
					for _, v := range violations {
						bytes, _ := json.Marshal(v)
						if _, e := f.WriteString(string(bytes) + "\n"); e != nil {
							pterm.Error.Println("cannot write violation to stream: " + e.Error())
						}
					}
					lock.Unlock()
				}
			}
		}
	}()
}
