// Copyright 2026 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package mockproxy

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pb33f/ranch/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleStaticResponseWritesResponseAndBroadcasts(t *testing.T) {
	id := uuid.New()
	request := &model.Request{
		Id:                 &id,
		HttpRequest:        httptest.NewRequest(http.MethodGet, "http://wiretap.local/static", nil),
		HttpResponseWriter: httptest.NewRecorder(),
	}
	response := &http.Response{
		StatusCode: http.StatusCreated,
		Header:     http.Header{"X-Test": []string{"yes"}},
		Body:       io.NopCloser(strings.NewReader("static body")),
	}

	type broadcastedResponse struct {
		status int
		body   string
	}
	broadcastedC := make(chan broadcastedResponse, 1)
	NewHandler().HandleStaticResponse(request, response, func(resp *http.Response) {
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		broadcastedC <- broadcastedResponse{status: resp.StatusCode, body: string(body)}
	})

	rec := request.HttpResponseWriter.(*httptest.ResponseRecorder)
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "yes", rec.Header().Get("X-Test"))
	assert.Equal(t, "static body", rec.Body.String())
	select {
	case broadcasted := <-broadcastedC:
		assert.Equal(t, http.StatusCreated, broadcasted.status)
		assert.Equal(t, "static body", broadcasted.body)
	case <-time.After(500 * time.Millisecond):
		require.Fail(t, "expected static response to be broadcast")
	}
}

func TestHandleStaticResponseDoesNotPanicOnBodyReadError(t *testing.T) {
	id := uuid.New()
	request := &model.Request{
		Id:                 &id,
		HttpRequest:        httptest.NewRequest(http.MethodGet, "http://wiretap.local/static", nil),
		HttpResponseWriter: httptest.NewRecorder(),
	}
	response := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{},
		Body:       failingReadCloser{},
	}

	var broadcasted bool
	assert.NotPanics(t, func() {
		NewHandler().HandleStaticResponse(request, response, func(_ *http.Response) {
			broadcasted = true
		})
	})

	rec := request.HttpResponseWriter.(*httptest.ResponseRecorder)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.False(t, broadcasted)
}

func TestHandleStaticResponseDoesNotPanicOnBodyWriteError(t *testing.T) {
	id := uuid.New()
	writer := &failingResponseWriter{header: http.Header{}}
	request := &model.Request{
		Id:                 &id,
		HttpRequest:        httptest.NewRequest(http.MethodGet, "http://wiretap.local/static", nil),
		HttpResponseWriter: writer,
	}
	response := &http.Response{
		StatusCode: http.StatusAccepted,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader("static body")),
	}

	assert.NotPanics(t, func() {
		NewHandler().HandleStaticResponse(request, response, nil)
	})
	assert.Equal(t, http.StatusAccepted, writer.code)
}

type failingReadCloser struct{}

func (failingReadCloser) Read(_ []byte) (int, error) {
	return 0, errors.New("read failed")
}

func (failingReadCloser) Close() error {
	return nil
}

type failingResponseWriter struct {
	header http.Header
	code   int
}

func (w *failingResponseWriter) Header() http.Header {
	return w.header
}

func (w *failingResponseWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write failed")
}

func (w *failingResponseWriter) WriteHeader(statusCode int) {
	w.code = statusCode
}
