// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package problems

import (
	"net/http"

	"github.com/pb33f/wiretap/shared"
)

const JSONContentType = "application/problem+json"

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

func ShouldReturnValidationProblem(
	config *shared.WiretapConfiguration,
	requestErrors, responseErrors []*shared.WiretapValidationError,
) bool {
	return config != nil &&
		config.HardErrorReturnProblem &&
		(len(requestErrors) > 0 || len(responseErrors) > 0)
}

func PickHardErrorStatus(
	isHardError bool,
	requestErrors, responseErrors []*shared.WiretapValidationError,
	config *shared.WiretapConfiguration,
	upstreamStatus int,
) int {
	if !isHardError {
		return upstreamStatus
	}
	hasReq := len(requestErrors) > 0
	hasResp := len(responseErrors) > 0
	switch {
	case hasReq && !hasResp:
		return config.HardErrorCode
	case hasResp:
		return config.HardErrorReturnCode
	default:
		return upstreamStatus
	}
}

func StripValidationProblemHeaders(headers http.Header) {
	for _, header := range validationProblemHeadersToDrop {
		headers.Del(header)
	}
	headers.Set("Cache-Control", "no-store")
}

// WriteValidationProblemResponse substitutes the HTTP response body with an
// RFC 9457 problem document. It strips stale upstream representation headers,
// prevents caching of the substituted body, overwrites Content-Type with
// application/problem+json, writes the status code, then writes the marshalled
// problem document.
//
// Caller must have already set any headers that should remain (e.g. CORS).
func WriteValidationProblemResponse(
	w http.ResponseWriter,
	status int,
	instance string,
	requestErrors, responseErrors []*shared.WiretapValidationError,
) {
	headers := w.Header()
	StripValidationProblemHeaders(headers)
	headers.Set("Content-Type", JSONContentType)

	problem := shared.BuildValidationProblem(status, instance, requestErrors, responseErrors)
	body := shared.MarshalValidationProblem(problem)

	w.WriteHeader(status)
	_, _ = w.Write(body)
}
