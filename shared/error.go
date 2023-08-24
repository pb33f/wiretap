// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package shared

import "encoding/json"

// WiretapError is an rfc7807 compliant error struct
type WiretapError struct {
	Type     string `json:"type,omitempty"`     // URI reference to the type of problem
	Title    string `json:"title"`              // A short description of the issue
	Status   int    `json:"status,omitempty"`   // HTTP status code.
	Detail   string `json:"detail"`             // explanation of the issue in detail.
	Instance string `json:"instance,omitempty"` // URI to the specific problem.
	Payload  any    `json:"payload,omitempty"`  // if added, this is the payload that caused the error
}

func GenerateError(title string,
	status int,
	detail string,
	instance string) *WiretapError {
	return &WiretapError{
		Type:     "https://pb33f.io/wiretap/error",
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: instance,
	}
}

func MarshalError(err *WiretapError) []byte {
	b, _ := json.Marshal(err)
	return b
}
