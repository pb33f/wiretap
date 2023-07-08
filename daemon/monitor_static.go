// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package daemon

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
	"github.com/pb33f/ranch/bus"
	"github.com/pb33f/ranch/model"
	"github.com/pb33f/wiretap/shared"
	"github.com/pterm/pterm"
	"os"
	"path/filepath"
)

func MonitorStatic(wiretapConfig *shared.WiretapConfiguration) {

	b := bus.GetBus()
	staticChan, _ := b.GetChannelManager().GetChannel(WiretapStaticChangeChan)

	go func() {
		var err error

		watcher, _ := fsnotify.NewWatcher()
		defer watcher.Close()

		watchDir := func(path string, fi os.FileInfo, err error) error {
			if fi.Mode().IsDir() {
				return watcher.Add(path)
			}
			return nil
		}

		if wErr := filepath.Walk(wiretapConfig.StaticDir, watchDir); err != nil {
			pterm.Fatal.Println(fmt.Sprintf("Error trying to monitor static directory: %s", wErr))
		}

		for {
			select {
			case event := <-watcher.Events:
				if event.Has(fsnotify.Write) {
					pterm.Info.Println(pterm.LightMagenta(fmt.Sprintf("[wiretap] static file changed: %s", event.Name)))

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
				pterm.Error.Println(fmt.Sprintf("[wiretap] static error: %s", wErr.Error()))
			}
		}
	}()
}
