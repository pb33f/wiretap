// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package cmd

import (
	"bufio"
	"fmt"
	"github.com/pb33f/wiretap/shared"
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
		htmlContent, er := fs.Sub(staticFS, "ui/dist")
		if er != nil {
			log.Fatal(err)
			return
		}
		assetContent, er := fs.Sub(staticFS, "ui/dist/assets")
		if er != nil {
			log.Fatal(err)
			return
		}

		// read in the index
		index, ierr := htmlContent.Open("index.html")
		if ierr != nil {
			log.Fatal(ierr)
			return
		}
		indexReader := bufio.NewReader(index)
		bytes, berr := io.ReadAll(indexReader)
		if berr != nil {
			log.Fatal(berr)
			return
		}

		// handle index will serve a modified index.html from the embedded filesystem.
		// this is so the monitor can connect to the websocket on the correct port.
		handleIndex := func(w http.ResponseWriter, r *http.Request) {
			indexString := string(bytes)

			// replace the port in the index.html file and serve it.
			indexString = strings.ReplaceAll(indexString, "%WIRETAP_PORT%", wiretapConfig.Port)
			io.WriteString(w, indexString)
		}

		// create a new mux.
		mux := http.NewServeMux()

		// create a new fileserver for the assets.
		fs := http.FileServer(http.FS(assetContent))

		// handle the index
		mux.HandleFunc("/", handleIndex)

		// handle the assets
		mux.Handle("/assets/", http.StripPrefix("/assets", fs))

		log.Printf("Monitor UI booting on port %s...", wiretapConfig.MonitorPort)
		err = http.ListenAndServe(fmt.Sprintf(":%s", wiretapConfig.MonitorPort), mux)
		if err != nil {
			log.Fatal(err)
		}
	}()
}
