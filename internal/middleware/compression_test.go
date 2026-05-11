package middleware_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KalessinD/gophermart/internal/middleware"
)

var compressionThreshold = 1

func simpleHandler(t *testing.T) http.Handler {
	t.Helper()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write(body)
	})
}

func TestCompression_DecodeRequest(t *testing.T) {
	originalBody := "Hello, Gzip!"
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	_, err := gzWriter.Write([]byte(originalBody))
	require.NoError(t, err)
	require.NoError(t, gzWriter.Close())

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "identity")

	rec := httptest.NewRecorder()

	handler := middleware.Compression(compressionThreshold)(simpleHandler(t))
	handler.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	respBody, _ := io.ReadAll(res.Body)
	assert.Equal(t, originalBody, string(respBody))

	assert.NotContains(t, res.Header.Get("Content-Encoding"), "gzip")
}

func TestCompression_EncodeResponse(t *testing.T) {
	originalBody := "This is a response that should be compressed"

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	rec := httptest.NewRecorder()

	handler := middleware.Compression(compressionThreshold)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(originalBody))
	}))

	handler.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, "gzip", res.Header.Get("Content-Encoding"))

	gzReader, err := gzip.NewReader(res.Body)
	require.NoError(t, err)
	defer gzReader.Close()

	decompressedBody, err := io.ReadAll(gzReader)
	require.NoError(t, err)

	assert.Equal(t, originalBody, string(decompressedBody))
}

func TestCompression_NoCompression(t *testing.T) {
	originalBody := "Plain text data"
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(originalBody))

	rec := httptest.NewRecorder()

	handler := middleware.Compression(compressionThreshold)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_, _ = w.Write(body)
	}))
	handler.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	assert.Empty(t, res.Header.Get("Content-Encoding"))

	respBody, _ := io.ReadAll(res.Body)
	assert.Equal(t, originalBody, string(respBody))
}

func TestCompression_BadGzipRequest(t *testing.T) {
	invalidGzipBody := bytes.NewBufferString("not a gzip content")

	req := httptest.NewRequest(http.MethodPost, "/", invalidGzipBody)
	req.Header.Set("Content-Encoding", "gzip")

	rec := httptest.NewRecorder()

	handler := middleware.Compression(compressionThreshold)(simpleHandler(t))
	handler.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestCompression_WriteHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	rec := httptest.NewRecorder()

	handler := middleware.Compression(compressionThreshold)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated) // 201
		_, _ = w.Write([]byte("created"))
	}))
	handler.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusCreated, res.StatusCode)
	assert.Equal(t, "gzip", res.Header.Get("Content-Encoding"))
}
