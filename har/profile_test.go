// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

//go:build profile

package har

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"testing"

	"github.com/pb33f/harific/motor"
	harModel "github.com/pb33f/harific/motor/model"
)

func TestProfileHARStreamer(t *testing.T) {
	harPath := os.Getenv("WIRETAP_PROFILE_HAR")
	if harPath == "" {
		t.Skip("set WIRETAP_PROFILE_HAR to profile a HAR file")
	}

	ctx := context.Background()
	runtime.GC()
	logMemStats(t, "start")

	streamer, err := NewHARStreamer(harPath, motor.DefaultStreamerOptions())
	if err != nil {
		t.Fatalf("create streamer: %v", err)
	}

	if err = streamer.Initialize(ctx); err != nil {
		t.Fatalf("initialize streamer: %v", err)
	}
	runtime.GC()
	logMemStats(t, "after initialize")
	writeHeapProfile(t, "after-initialize")

	index := streamer.GetIndex()
	if index == nil || index.TotalEntries == 0 {
		t.Fatalf("empty streamer index")
	}

	results, err := streamer.StreamRange(ctx, 0, index.TotalEntries)
	if err != nil {
		t.Fatalf("stream range: %v", err)
	}

	requests := 0
	for result := range results {
		if result.Error != nil {
			t.Fatalf("stream entry %d: %v", result.Index, result.Error)
		}
		if result.Entry == nil {
			continue
		}
		req, err := harModel.ConvertRequestIntoHttpRequest(result.Entry.Request)
		if err != nil {
			t.Fatalf("convert request %d: %v", result.Index, err)
		}
		if req.Method != http.MethodOptions {
			requests++
		}
	}

	runtime.GC()
	t.Logf("streamed requests=%d entries=%d", requests, index.TotalEntries)
	logMemStats(t, "after stream")
	writeHeapProfile(t, "after-stream")

	if err = streamer.Close(); err != nil {
		t.Fatalf("close streamer: %v", err)
	}
	runtime.GC()
	logMemStats(t, "after close")
	writeHeapProfile(t, "after-close")
}

func logMemStats(t *testing.T, label string) {
	t.Helper()

	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	t.Logf("%s: alloc=%dMB total_alloc=%dMB sys=%dMB heap_alloc=%dMB heap_inuse=%dMB heap_idle=%dMB heap_released=%dMB objects=%d",
		label,
		stats.Alloc/1024/1024,
		stats.TotalAlloc/1024/1024,
		stats.Sys/1024/1024,
		stats.HeapAlloc/1024/1024,
		stats.HeapInuse/1024/1024,
		stats.HeapIdle/1024/1024,
		stats.HeapReleased/1024/1024,
		stats.HeapObjects)
}

func writeHeapProfile(t *testing.T, label string) {
	t.Helper()

	dir := os.Getenv("WIRETAP_PROFILE_DIR")
	if dir == "" {
		return
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("create profile dir: %v", err)
	}

	profilePath := filepath.Join(dir, "har-"+label+".pprof")
	file, err := os.Create(profilePath)
	if err != nil {
		t.Fatalf("create heap profile: %v", err)
	}
	defer file.Close()

	if err = pprof.WriteHeapProfile(file); err != nil {
		t.Fatalf("write heap profile: %v", err)
	}
	t.Logf("wrote heap profile: %s", profilePath)
}
