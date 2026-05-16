// Package middleware holds the HTTP request-pipeline handlers (access logging,
// panic recovery, gzip compression) and the helpers they share. It is the HTTP
// analog of grpc-sandbox's interceptor package: request-scoped logic that is
// worth testing in isolation, kept separate from server bootstrap and route
// registration.
package middleware

import (
	"compress/gzip"
	"context"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel/trace"
)

// TraceableRequest reports whether a request should be traced and access
// logged. Health checks are excluded to keep the signal clean. It is also used
// as the otelhttp span filter.
func TraceableRequest(r *http.Request) bool {
	return r.URL.Path != "/health"
}

func extractTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().TraceID().String()
	}

	return ""
}

// AccessLog records method, path, status, duration and response size for every
// traceable request, choosing the log level from the response status class.
func AccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !TraceableRequest(r) {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		recorder := &accessLogResponseWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(recorder, r)

		attrs := []any{
			"method", r.Method,
			"path", r.URL.Path,
			"status", recorder.status,
			"duration", time.Since(start),
			"bytes", recorder.bytes,
		}
		if route := r.Pattern; route != "" {
			attrs = append(attrs, "route", route)
		}
		if traceID := extractTraceID(r.Context()); traceID != "" {
			attrs = append(attrs, "trace_id", traceID)
		}

		switch {
		case recorder.status >= http.StatusInternalServerError:
			slog.ErrorContext(r.Context(), "HTTP request completed", attrs...)
		case recorder.status >= http.StatusBadRequest:
			slog.WarnContext(r.Context(), "HTTP request completed", attrs...)
		default:
			slog.InfoContext(r.Context(), "HTTP request completed", attrs...)
		}
	})
}

type accessLogResponseWriter struct {
	http.ResponseWriter
	status      int
	bytes       int
	wroteHeader bool
}

func (w *accessLogResponseWriter) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}
	w.status = status
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(status)
}

func (w *accessLogResponseWriter) Write(body []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	n, err := w.ResponseWriter.Write(body)
	w.bytes += n
	return n, err
}

// Recover converts a panic in a downstream handler into a 500 response and an
// error log instead of crashing the server.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				slog.ErrorContext(r.Context(), "HTTP request panic", "error", recovered)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

type gzipResponseWriter struct {
	http.ResponseWriter
	writer      *gzip.Writer
	wroteHeader bool
	gzipEnabled bool
}

func (grw *gzipResponseWriter) WriteHeader(status int) {
	if grw.wroteHeader {
		return
	}

	grw.wroteHeader = true
	if status >= http.StatusOK && status < http.StatusMultipleChoices {
		grw.Header().Del("Content-Length")
		grw.Header().Set("Content-Encoding", "gzip")
		grw.writer = gzipPool.Get().(*gzip.Writer)
		grw.writer.Reset(grw.ResponseWriter)
		grw.gzipEnabled = true
	}

	grw.ResponseWriter.WriteHeader(status)
}

func (grw *gzipResponseWriter) Write(b []byte) (int, error) {
	if !grw.wroteHeader {
		grw.WriteHeader(http.StatusOK)
	}
	if !grw.gzipEnabled {
		return grw.ResponseWriter.Write(b)
	}
	return grw.writer.Write(b)
}

func (grw *gzipResponseWriter) Close() error {
	if grw.writer == nil {
		return nil
	}

	err := grw.writer.Close()
	gzipPool.Put(grw.writer)
	grw.writer = nil
	return err
}

var gzipPool = sync.Pool{
	New: func() any {
		return gzip.NewWriter(io.Discard)
	},
}

// Gzip compresses successful responses when the client advertises gzip support.
func Gzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Add("Vary", "Accept-Encoding")

		grw := &gzipResponseWriter{ResponseWriter: w}
		defer func() {
			if err := grw.Close(); err != nil {
				slog.WarnContext(r.Context(), "close gzip response writer", "error", err)
			}
		}()

		next.ServeHTTP(grw, r)
	})
}
