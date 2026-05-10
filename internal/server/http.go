package server

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/danielgtaylor/huma/v2/yaml"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"

	userv1 "oas-sandbox/gen/grpc/user/v1"
	workerv1 "oas-sandbox/gen/grpc/worker/v1"
	"oas-sandbox/internal/api"
	"oas-sandbox/internal/api/system"
	usersv1 "oas-sandbox/internal/api/users/v1"
	workerv1api "oas-sandbox/internal/api/worker/v1"
)

type HTTPServer struct {
	server *http.Server
}

func NewHTTPServer(
	port int,
	serviceName string,
	userClient userv1.UserServiceClient,
	workerClient workerv1.WorkerServiceClient,
) *HTTPServer {
	mux := http.NewServeMux()
	humaAPI := NewAPI(mux, serviceName, userClient, workerClient)
	registerOpenAPISpec(mux, humaAPI)
	registerSwaggerUIAssets(mux)
	registerDocs(mux, humaAPI.OpenAPI().Info.Title)

	handler := otelhttp.NewHandler(
		accessLogHandler(gzipHandler(recoverHandler(mux))),
		serviceName,
		otelhttp.WithFilter(traceableRequest),
	)
	return &HTTPServer{
		server: &http.Server{
			Addr:              fmt.Sprintf(":%d", port),
			Handler:           handler,
			ReadHeaderTimeout: 5 * time.Second,
		},
	}
}

func accessLogHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !traceableRequest(r) {
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

func recoverHandler(next http.Handler) http.Handler {
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

func gzipHandler(next http.Handler) http.Handler {
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

func traceableRequest(r *http.Request) bool {
	return r.URL.Path != "/health"
}

func extractTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().TraceID().String()
	}

	return ""
}

func NewAPI(
	mux *http.ServeMux,
	serviceName string,
	userClient userv1.UserServiceClient,
	workerClient workerv1.WorkerServiceClient,
) huma.API {
	humaConfig := huma.DefaultConfig("OAS Sandbox", "0.1.0")
	humaConfig.Info.Description = "OpenAPI sandbox for homelab API experiments."
	humaConfig.Tags = []*huma.Tag{
		{Name: api.TagSystem, Description: "Service health and operational endpoints."},
		{Name: api.TagUsers, Description: "User resource endpoints."},
		{Name: api.TagWorker, Description: "Background worker job endpoints."},
	}
	humaConfig.DocsPath = ""
	humaConfig.OpenAPIPath = ""

	humaAPI := humago.New(mux, humaConfig)
	registerAPI(humaAPI, api.Deps{
		ServiceName:  serviceName,
		UserClient:   userClient,
		WorkerClient: workerClient,
	})

	return humaAPI
}

func (s *HTTPServer) Start() error {
	slog.Info("HTTP server listening", "addr", s.server.Addr)
	return s.server.ListenAndServe()
}

func (s *HTTPServer) Stop(ctx context.Context) {
	if err := s.server.Shutdown(ctx); err != nil {
		slog.WarnContext(ctx, "HTTP server shutdown error", "error", err)
	}
}

func registerAPI(h huma.API, deps api.Deps) {
	system.Register(h, deps)

	v1 := huma.NewGroup(h, "/v1")
	usersv1.Register(v1, deps)
	workerv1api.Register(v1, deps)
}

func registerOpenAPISpec(mux *http.ServeMux, humaAPI huma.API) {
	buildJSON := func() ([]byte, error) {
		return humaAPI.OpenAPI().MarshalJSON()
	}

	serveJSON := func(w http.ResponseWriter, r *http.Request, asAttachment bool) {
		body, err := buildJSON()
		if err != nil {
			slog.ErrorContext(r.Context(), "render OpenAPI JSON", "error", err)
			http.Error(w, "failed to render OpenAPI document", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/openapi+json")
		if asAttachment {
			w.Header().Set("Content-Disposition", `attachment; filename="openapi.json"`)
		}
		_, _ = w.Write(body)
	}

	serveYAML := func(w http.ResponseWriter, r *http.Request, asAttachment bool) {
		jsonBody, err := buildJSON()
		if err != nil {
			slog.ErrorContext(r.Context(), "render OpenAPI JSON", "error", err)
			http.Error(w, "failed to render OpenAPI document", http.StatusInternalServerError)
			return
		}
		var buf bytes.Buffer
		if err := yaml.Convert(&buf, bytes.NewReader(jsonBody)); err != nil {
			slog.ErrorContext(r.Context(), "convert OpenAPI to YAML", "error", err)
			http.Error(w, "failed to render OpenAPI document", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/openapi+yaml")
		if asAttachment {
			w.Header().Set("Content-Disposition", `attachment; filename="openapi.yaml"`)
		}
		_, _ = w.Write(buf.Bytes())
	}

	mux.HandleFunc("GET /openapi.json", func(w http.ResponseWriter, r *http.Request) { serveJSON(w, r, false) })
	mux.HandleFunc("GET /openapi.yaml", func(w http.ResponseWriter, r *http.Request) { serveYAML(w, r, false) })
	mux.HandleFunc("GET /openapi.json/download", func(w http.ResponseWriter, r *http.Request) { serveJSON(w, r, true) })
	mux.HandleFunc("GET /openapi.yaml/download", func(w http.ResponseWriter, r *http.Request) { serveYAML(w, r, true) })
}

func registerSwaggerUIAssets(mux *http.ServeMux) {
	sub, err := fs.Sub(swaggerUIFS, "assets/swagger-ui")
	if err != nil {
		panic(fmt.Errorf("embed swagger ui: %w", err))
	}

	fileServer := http.FileServer(http.FS(sub))
	mux.Handle("GET /assets/swagger-ui/", http.StripPrefix("/assets/swagger-ui/", fileServer))
}

func registerDocs(mux *http.ServeMux, title string) {
	if title == "" {
		title = "SwaggerUI in HTML"
	} else {
		title += " Reference"
	}

	csp := strings.Join([]string{
		"default-src 'none'",
		"base-uri 'none'",
		"connect-src 'self'",
		"font-src 'self' data:",
		"form-action 'none'",
		"frame-ancestors 'none'",
		"img-src 'self' data:",
		"sandbox allow-same-origin allow-scripts allow-popups allow-popups-to-escape-sandbox allow-downloads",
		"script-src 'self'",
		"style-src 'self' 'unsafe-inline'",
	}, "; ")

	body := []byte(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="referrer" content="no-referrer">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>` + title + `</title>
    <link rel="stylesheet" href="/assets/swagger-ui/swagger-ui.css">
    <link rel="stylesheet" href="/assets/swagger-ui/docs-overrides.css">
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="/assets/swagger-ui/swagger-ui-bundle.js"></script>
    <script src="/assets/swagger-ui/swagger-initializer.js" data-url="/openapi.json"></script>
  </body>
</html>`)

	mux.HandleFunc("GET /docs", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Security-Policy", csp)
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write(body)
	})
}
