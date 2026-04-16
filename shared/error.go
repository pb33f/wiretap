// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package shared

import (
	"encoding/json"
	"fmt"

	"github.com/pb33f/libopenapi-validator/errors"
)

// WiretapError is an RFC 9457 compliant error struct.
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
	instance string, payload any) *WiretapError {
	return &WiretapError{
		Type:     "https://pb33f.io/wiretap/error",
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: instance,
		Payload:  payload,
	}
}

func MarshalError(err *WiretapError) []byte {
	b, _ := json.Marshal(err)
	return b
}

type WiretapValidationError struct {
	errors.ValidationError
	SpecName string `json:"specName" yaml:"specName"`
}

func ConvertValidationErrors(specName string, validationErrors []*errors.ValidationError) []*WiretapValidationError {

	wiretapValidationErrors := make([]*WiretapValidationError, 0)
	for _, validationError := range validationErrors {
		wiretapValidationError := &WiretapValidationError{
			ValidationError: *validationError,
			SpecName:        specName,
		}

		wiretapValidationErrors = append(wiretapValidationErrors, wiretapValidationError)
	}

	return wiretapValidationErrors
}

// ValidationProblemType is the stable RFC 9457 problem type URI for wiretap
// validation failures. Kept distinct from the generic wiretap error type so
// clients can discriminate validation problems from other wiretap problems.
const ValidationProblemType = "https://pb33f.io/wiretap/validation-error"

// ValidationProblem extends WiretapError with first-class validation error
// lists. Produced when hard-validation triggers and HardErrorReturnProblem is
// enabled. Serialises to application/problem+json.
type ValidationProblem struct {
	*WiretapError
	RequestErrors  []*WiretapValidationError `json:"requestErrors,omitempty"`
	ResponseErrors []*WiretapValidationError `json:"responseErrors,omitempty"`
}

// BuildValidationProblem constructs an RFC 9457 problem document describing
// request and/or response validation failures. Delegates the base WiretapError
// construction to GenerateError so problem documents have a single construction
// path.
func BuildValidationProblem(
	status int,
	instance string,
	requestErrors, responseErrors []*WiretapValidationError,
) *ValidationProblem {
	reqCount := len(requestErrors)
	respCount := len(responseErrors)

	var title string
	switch {
	case reqCount > 0 && respCount > 0:
		title = "Request and response validation failed"
	case reqCount > 0:
		title = "Request validation failed"
	default:
		title = "Response validation failed"
	}

	detail := fmt.Sprintf("%d validation error(s) detected", reqCount+respCount)

	base := GenerateError(title, status, detail, instance, nil)
	base.Type = ValidationProblemType

	return &ValidationProblem{
		WiretapError:   base,
		RequestErrors:  requestErrors,
		ResponseErrors: responseErrors,
	}
}

// MarshalValidationProblem serialises a ValidationProblem to JSON suitable for
// an application/problem+json response body. Sibling to MarshalError. Guards
// against a nil embedded *WiretapError — field promotion would otherwise
// panic at marshal time if a caller constructed the outer struct directly.
func MarshalValidationProblem(p *ValidationProblem) []byte {
	if p == nil || p.WiretapError == nil {
		return []byte("{}")
	}
	b, _ := json.Marshal(p)
	return b
}
