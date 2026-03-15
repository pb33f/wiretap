// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package har

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	harModel "github.com/pb33f/harific/motor/model"
)

func BuildHAR(har []byte) (*harModel.HAR, error) {
	if har == nil {
		return nil, fmt.Errorf("HAR bytes are empty")
	}

	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	var harFile harModel.HAR

	err := json.Unmarshal(har, &harFile)

	if err != nil {
		return nil, err
	}
	return &harFile, nil
}
