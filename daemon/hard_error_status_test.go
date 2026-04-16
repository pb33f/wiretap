// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package daemon

import (
	"testing"

	"github.com/pb33f/wiretap/shared"
	"github.com/stretchr/testify/assert"
)

func TestPickHardErrorStatus(t *testing.T) {
	config := &shared.WiretapConfiguration{
		HardErrorCode:       400,
		HardErrorReturnCode: 502,
	}
	reqErr := buildSampleErrors("bad request")
	respErr := buildSampleErrors("bad response")

	tests := []struct {
		name           string
		isHardError    bool
		requestErrors  []*shared.WiretapValidationError
		responseErrors []*shared.WiretapValidationError
		upstreamStatus int
		want           int
	}{
		{"hard off falls through to upstream", false, reqErr, respErr, 200, 200},
		{"hard on, request invalid only", true, reqErr, nil, 200, 400},
		{"hard on, response invalid only", true, nil, respErr, 200, 502},
		{"hard on, both invalid prefers response code", true, reqErr, respErr, 200, 502},
		{"hard on, neither invalid falls through", true, nil, nil, 201, 201},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := pickHardErrorStatus(tc.isHardError, tc.requestErrors, tc.responseErrors, config, tc.upstreamStatus)
			assert.Equal(t, tc.want, got)
		})
	}
}
