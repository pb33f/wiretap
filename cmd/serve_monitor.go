// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package cmd

import (
	"bufio"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/pb33f/wiretap/shared"
	"github.com/pterm/pterm"
	"io"
	"io/fs"
	"log"
	"net/http"
	"strings"
)

func serveMonitor(wiretapConfig *shared.WiretapConfiguration) {
	go func() {
		var err error
		var staticFS = fs.FS(wiretapConfig.FS)
		htmlContent, er := fs.Sub(staticFS, shared.UILocation)
		if er != nil {
			log.Fatal(err)
			return
		}
		assetContent, er := fs.Sub(staticFS, shared.UIAssetsLocation)
		if er != nil {
			log.Fatal(err)
			return
		}

		// read in the index
		index, iErr := htmlContent.Open(shared.IndexFile)
		if iErr != nil {
			log.Fatal(iErr)
			return
		}
		indexReader := bufio.NewReader(index)
		bytes, bErr := io.ReadAll(indexReader)
		if bErr != nil {
			log.Fatal(bErr)
			return
		}

		indexString := string(bytes)

		// replace the port in the index.html file and serve it.
		indexString = strings.ReplaceAll(strings.ReplaceAll(indexString, shared.WiretapPortPlaceholder, wiretapConfig.WebSocketPort),
			shared.WiretapVersionPlaceholder, wiretapConfig.Version)

		// handle index will serve a modified index.html from the embedded filesystem.
		// this is so the monitor can connect to the websocket on the correct port.
		handleIndex := func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.WriteString(w, indexString)
		}

		// create a new mux.
		mux := http.NewServeMux()

		// create a new file server for the assets.
		fileServer := http.FileServer(http.FS(assetContent))

		// handle the index
		mux.HandleFunc("/", handleIndex)

		// compress everything!
		fileServer = handlers.CompressHandler(fileServer)

		// handle the assets
		mux.Handle("/assets/", http.StripPrefix("/assets", fileServer))

		pterm.Info.Println(pterm.LightMagenta(fmt.Sprintf("Monitor UI booting on port %s...", wiretapConfig.MonitorPort)))

		err = http.ListenAndServe(fmt.Sprintf(":%s", wiretapConfig.MonitorPort), mux)
		if err != nil {
			log.Fatal(err)
		}
	}()
}
