// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package har

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/pb33f/harhar"
)

func BuildHAR(har []byte) (*harhar.HAR, error) {
	if har == nil {
		return nil, fmt.Errorf("HAR bytes are empty")
	}

	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	var harFile harhar.HAR

	err := json.Unmarshal(har, &harFile)
	if err != nil {
		return nil, err
	}
	return &harFile, nil
}
