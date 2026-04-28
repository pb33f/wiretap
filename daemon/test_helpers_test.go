// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package daemon

import (
	validationerrors "github.com/pb33f/libopenapi-validator/errors"
	"github.com/pb33f/wiretap/shared"
)

func buildSampleErrors(messages ...string) []*shared.WiretapValidationError {
	out := make([]*shared.WiretapValidationError, 0, len(messages))
	for _, m := range messages {
		out = append(out, &shared.WiretapValidationError{
			ValidationError: validationerrors.ValidationError{Message: m},
			SpecName:        "spec.yaml",
		})
	}
	return out
}
