package middleware

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newRequest(method, target string) *http.Request {
	return httptest.NewRequestWithContext(context.Background(), method, target, http.NoBody)
}

func TestGzipDoesNotCompressErrorResponse(t *testing.T) {
	handler := Gzip(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	req := newRequest(http.MethodGet, "/bad")
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Empty(t, rec.Header().Get("Content-Encoding"))
	assert.Contains(t, rec.Header().Values("Vary"), "Accept-Encoding")
	assert.Contains(t, rec.Body.String(), "bad request")
}

func TestGzipCompressesSuccessfulNonOKResponse(t *testing.T) {
	handler := Gzip(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "7")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	}))
	req := newRequest(http.MethodPost, "/created")
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))
	assert.Empty(t, rec.Header().Get("Content-Length"))

	gz, err := gzip.NewReader(rec.Body)
	require.NoError(t, err)
	defer gz.Close()

	body, err := io.ReadAll(gz)
	require.NoError(t, err)
	assert.Equal(t, "created", string(body))
}

func TestGzipSkipsWhenClientDoesNotAcceptGzip(t *testing.T) {
	handler := Gzip(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("plain"))
	}))
	req := newRequest(http.MethodGet, "/plain")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, "plain", rec.Body.String())
	assert.Empty(t, rec.Header().Get("Content-Encoding"))
}

func TestRecover(t *testing.T) {
	handler := Recover(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}))
	req := newRequest(http.MethodGet, "/panic")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), http.StatusText(http.StatusInternalServerError))
}

func TestTraceableRequestSkipsHealth(t *testing.T) {
	assert.False(t, TraceableRequest(newRequest(http.MethodGet, "/health")))
	assert.True(t, TraceableRequest(newRequest(http.MethodGet, "/v1/users")))
}

func TestAccessLogLogsRequest(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	handler := AccessLog(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("ok"))
	}))
	req := newRequest(http.MethodPost, "/v1/users")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	logBody := logs.String()
	assert.Contains(t, logBody, `"msg":"HTTP request completed"`)
	assert.Contains(t, logBody, `"method":"POST"`)
	assert.Contains(t, logBody, `"path":"/v1/users"`)
	assert.Contains(t, logBody, `"status":201`)
	assert.Contains(t, logBody, `"bytes":2`)
}

func TestAccessLogCapturesImplicitOK(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	handler := AccessLog(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	req := newRequest(http.MethodGet, "/v1/users")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	logBody := logs.String()
	assert.Contains(t, logBody, `"status":200`)
	assert.Contains(t, logBody, `"bytes":2`)
}

func TestAccessLogServerErrorLevel(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	handler := AccessLog(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	handler.ServeHTTP(httptest.NewRecorder(), newRequest(http.MethodGet, "/v1/users"))

	assert.Contains(t, logs.String(), `"level":"ERROR"`)
}

func TestAccessLogClientErrorLevel(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	handler := AccessLog(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	handler.ServeHTTP(httptest.NewRecorder(), newRequest(http.MethodGet, "/v1/users"))

	assert.Contains(t, logs.String(), `"level":"WARN"`)
}

func TestAccessLogSkipsHealth(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	handler := AccessLog(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	req := newRequest(http.MethodGet, "/health")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, logs.String())
}
