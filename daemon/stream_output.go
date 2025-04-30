// Copyright 2024 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package daemon

import (
	"fmt"
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
		if _, e := f.WriteString("[]"); e != nil {
			pterm.Error.Println("cannot write violation to stream: " + err.Error())
		}
		for {
			select {
			case violations := <-ws.streamChan:

				if ws.stream {
					lock.Lock()

					fi, _ := f.Stat()
					_ = os.Truncate(ws.reportFile, fi.Size()-1)
					if fi.Size() > 2 {
						_, _ = f.WriteString(",\n")
					}
					ws.streamViolations = append(ws.streamViolations, violations...)

					for i, v := range violations {
						bytes, _ := json.Marshal(v)
						if _, e := f.WriteString(fmt.Sprintf("%s", bytes)); e != nil {
							pterm.Error.Println("cannot write violation to stream: " + err.Error())
						}
						if i > len(violations)-1 {
							_, _ = f.WriteString(",\n")
						}
					}
					_, _ = f.WriteString("]")
					lock.Unlock()
				}
			}
		}
	}()
}
