// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/pb33f/wiretap/shared"
	"github.com/pterm/pterm"
	"net/http"
	"os"
	"path/filepath"
)

func serveStatic(wiretapConfig *shared.WiretapConfiguration) {
	go func() {
		var err error

		fileServer := http.FileServer(http.Dir(wiretapConfig.StaticDir))

		pterm.Info.Println(pterm.LightMagenta(fmt.Sprintf("Serving static content from '%s' on port %s...",
			wiretapConfig.StaticDir, wiretapConfig.StaticPort)))

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

		go func() {
			for {
				select {
				case event := <-watcher.Events:
					if event.Has(fsnotify.Write) {
						pterm.Info.Println(pterm.LightMagenta(fmt.Sprintf("[wiretap] reloading static file: %s", event.Name)))
					}
				case wErr := <-watcher.Errors:
					pterm.Error.Println(fmt.Sprintf("[wiretap] static error: %s", wErr.Error()))
				}
			}
		}()

		err = http.ListenAndServe(fmt.Sprintf(":%s", wiretapConfig.StaticPort), fileServer)
		if err != nil {
			pterm.Fatal.Printf("Fatal error serving static content: %s\n", err.Error())
		}
	}()
}
