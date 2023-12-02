// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package har

import (
    "fmt"
    "github.com/pb33f/harhar"
    "github.com/pb33f/libopenapi-validator/errors"
)

type Transaction struct {
    Request  *harhar.Request
    Response *harhar.Response
}

func ValidateHAR(har *harhar.HAR) []*errors.ValidationError {

    //var validationErrors []*errors.ValidationError

    for _, entry := range har.Log.Entries {

        httpRequest, err := harhar.ConvertRequestIntoHttpRequest(entry.Request)
        if err != nil {

            return nil
        }
        fmt.Sprintf("httpRequest: %v", httpRequest)

        //validationErrors = append(validationErrors, validateRequest(entry.Request)...)
        //validationErrors = append(validationErrors, validateResponse(entry.Request, entry.Response)...)
    }

    return nil

}
