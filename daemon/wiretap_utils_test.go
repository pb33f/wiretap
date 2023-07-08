// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

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
	assert.Equal(t, "http://localhost:8000/", reconstructURL(r, protocol, host, port))
	r, _ = http.NewRequest("GET", "http://localhost:1337", nil)
	assert.Equal(t, "http://localhost:8000", reconstructURL(r, protocol, host, port))
	// Adding port correctly
	r, _ = http.NewRequest("GET", "http://localhost/", nil)
	assert.Equal(t, "http://localhost/", reconstructURL(r, protocol, host, ""))
	r, _ = http.NewRequest("GET", "http://localhost:8000", nil)
	assert.Equal(t, "http://localhost", reconstructURL(r, protocol, host, ""))
	// Adding paths correctly
	r, _ = http.NewRequest("POST", "http://localhost/dalek", nil)
	assert.Equal(t, "http://localhost:8000/dalek", reconstructURL(r, protocol, host, port))
	r, _ = http.NewRequest("PUT", "http://localhost/dalek/1337", nil)
	assert.Equal(t, "http://localhost/dalek/1337", reconstructURL(r, protocol, host, ""))
	// Adding query params correctly
	r, _ = http.NewRequest("GET", "http://localhost?doctor=who", nil)
	assert.Equal(t, "http://localhost:8000?doctor=who", reconstructURL(r, protocol, host, port))
	r, _ = http.NewRequest("GET", "http://localhost:1337?doctor=who", nil)
	assert.Equal(t, "http://localhost?doctor=who", reconstructURL(r, protocol, host, ""))
}
