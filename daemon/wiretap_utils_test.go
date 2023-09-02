// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package daemon

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReconstructURL(t *testing.T) {
	protocol := "http"
	host := "localhost"
	port := "8000"
	// Making sure trailing slashes are accounted for correctly
	r, _ := http.NewRequest("GET", "http://localhost:1337/", nil)
	assert.Equal(t, "http://localhost:8000/", ReconstructURL(r, protocol, host, port))
	r, _ = http.NewRequest("GET", "http://localhost:1337", nil)
	assert.Equal(t, "http://localhost:8000", ReconstructURL(r, protocol, host, port))
	// Adding port correctly
	r, _ = http.NewRequest("GET", "http://localhost/", nil)
	assert.Equal(t, "http://localhost/", ReconstructURL(r, protocol, host, ""))
	r, _ = http.NewRequest("GET", "http://localhost:8000", nil)
	assert.Equal(t, "http://localhost", ReconstructURL(r, protocol, host, ""))
	// Adding paths correctly
	r, _ = http.NewRequest("POST", "http://localhost/dalek", nil)
	assert.Equal(t, "http://localhost:8000/dalek", ReconstructURL(r, protocol, host, port))
	r, _ = http.NewRequest("PUT", "http://localhost/dalek/1337", nil)
	assert.Equal(t, "http://localhost/dalek/1337", ReconstructURL(r, protocol, host, ""))
	// Adding query params correctly
	r, _ = http.NewRequest("GET", "http://localhost?doctor=who", nil)
	assert.Equal(t, "http://localhost:8000?doctor=who", ReconstructURL(r, protocol, host, port))
	r, _ = http.NewRequest("GET", "http://localhost:1337?doctor=who", nil)
	assert.Equal(t, "http://localhost?doctor=who", ReconstructURL(r, protocol, host, ""))
}
