// Copyright 2024 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: BUSL-1.1

package daemon

import (
	jsoniter "github.com/json-iterator/go"
	"os"
	"sync"
)

func (ws *WiretapService) listenForValidationErrors() {

	var lock sync.RWMutex
	json := jsoniter.ConfigCompatibleWithStandardLibrary

	_ = os.Remove(ws.reportFile)
	f, err := os.OpenFile(ws.reportFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		serviceLogger(ws).Error("cannot stream violations", "error", err)
		return
	}

	go func() {
		defer f.Close()
		for {
			select {
			case violations := <-ws.streamChan:
				if ws.stream {
					lock.Lock()
					for _, v := range violations {
						bytes, _ := json.Marshal(v)
						if _, e := f.WriteString(string(bytes) + "\n"); e != nil {
							serviceLogger(ws).Error("cannot write violation to stream", "error", e)
						}
					}
					lock.Unlock()
				}
			}
		}
	}()
}
