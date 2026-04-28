// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package har

import (
	"context"
	"fmt"

	"github.com/pb33f/harific/motor"
)

func NewHARStreamer(path string, opts motor.StreamerOptions) (motor.HARStreamer, error) {
	if path == "" {
		return nil, fmt.Errorf("HAR path is empty")
	}

	return motor.NewHARStreamer(path, opts)
}

func CountHARMessages(path string) (int, error) {
	streamer, err := NewHARStreamer(path, motor.DefaultStreamerOptions())
	if err != nil {
		return 0, err
	}
	defer streamer.Close()

	ctx := context.Background()
	if err = streamer.Initialize(ctx); err != nil {
		return 0, err
	}

	index := streamer.GetIndex()
	if index == nil {
		return 0, nil
	}

	count := 0
	for _, entry := range index.Entries {
		if entry.Method != "" {
			count++
		}
		if entry.StatusCode > 0 {
			count++
		}
	}
	return count, nil
}
