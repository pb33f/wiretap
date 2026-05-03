// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: BUSL-1.1

package daemon

import (
	"log/slog"

	"github.com/pb33f/wiretap/shared"
)

func wiretapLogger(config *shared.WiretapConfiguration) *slog.Logger {
	if config != nil && config.Logger != nil {
		return config.Logger
	}
	return slog.Default()
}

func serviceLogger(ws *WiretapService) *slog.Logger {
	if ws != nil {
		return wiretapLogger(ws.config)
	}
	return slog.Default()
}
