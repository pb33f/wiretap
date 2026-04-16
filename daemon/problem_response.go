// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package daemon

import (
	"net/http"

	"github.com/pb33f/wiretap/shared"
)

const problemJSONContentType = "application/problem+json"

var validationProblemHeadersToDrop = []string{
	"Accept-Ranges",
	"Age",
	"Content-Encoding",
	"Content-Language",
	"Content-Length",
	"Content-Location",
	"Content-Range",
	"Digest",
	"ETag",
	"Expires",
	"Last-Modified",
	"Trailer",
	"Transfer-Encoding",
}

func shouldReturnValidationProblem(
	config *shared.WiretapConfiguration,
	requestErrors, responseErrors []*shared.WiretapValidationError,
) bool {
	return config != nil &&
		config.HardErrorReturnProblem &&
		(len(requestErrors) > 0 || len(responseErrors) > 0)
}

func stripValidationProblemHeaders(headers http.Header) {
	for _, header := range validationProblemHeadersToDrop {
		headers.Del(header)
	}
	headers.Set("Cache-Control", "no-store")
}

// writeValidationProblemResponse substitutes the HTTP response body with an
// RFC 9457 problem document. It strips stale upstream representation headers,
// prevents caching of the substituted body, overwrites Content-Type with
// application/problem+json, writes the status code, then writes the marshalled
// problem document.
//
// Caller must have already set any headers that should remain (e.g. CORS).
func writeValidationProblemResponse(
	w http.ResponseWriter,
	status int,
	instance string,
	requestErrors, responseErrors []*shared.WiretapValidationError,
) {
	headers := w.Header()
	stripValidationProblemHeaders(headers)
	headers.Set("Content-Type", problemJSONContentType)

	problem := shared.BuildValidationProblem(status, instance, requestErrors, responseErrors)
	body := shared.MarshalValidationProblem(problem)

	w.WriteHeader(status)
	_, _ = w.Write(body)
}
