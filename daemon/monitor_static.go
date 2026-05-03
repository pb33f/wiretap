// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: BUSL-1.1

package daemon

import (
	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
	"github.com/pb33f/ranch/bus"
	"github.com/pb33f/ranch/model"
	"github.com/pb33f/wiretap/shared"
	"os"
	"path/filepath"
)

func MonitorStatic(wiretapConfig *shared.WiretapConfiguration, eventBus bus.EventBus) {

	staticChan, _ := eventBus.GetChannelManager().GetChannel(WiretapStaticChangeChan)

	go func() {
		var err error

		watcher, _ := fsnotify.NewWatcher()
		defer watcher.Close()

		watchDir := func(path string, fi os.FileInfo, err error) error {
			if fi != nil && !fi.Mode().IsDir() {
				return watcher.Add(path)
			}
			if fi == nil {
				wiretapLogger(wiretapConfig).Error("Error trying to monitor static directory", "error", err)
			}
			return nil
		}

		if wErr := filepath.Walk(wiretapConfig.StaticDir, watchDir); wErr != nil {
			wiretapLogger(wiretapConfig).Error("Error trying to monitor static directory", "error", wErr)
			return
		}

		for {
			select {
			case event := <-watcher.Events:
				if event.Has(fsnotify.Write) {
					wiretapLogger(wiretapConfig).Info("[wiretap] static file changed", "file", event.Name)

					// broadcast the change to all connected clients
					ch := make(map[string]string)
					ch["file"] = event.Name

					id, _ := uuid.NewUUID()
					staticChan.Send(&model.Message{
						Id:            &id,
						DestinationId: &id,
						Error:         err,
						Channel:       WiretapStaticChangeChan,
						Destination:   WiretapStaticChangeChan,
						Payload:       ch,
						Direction:     model.ResponseDir,
					})

				}
			case wErr := <-watcher.Errors:
				wiretapLogger(wiretapConfig).Error("[wiretap] static error", "error", wErr)
			}
		}
	}()
}
