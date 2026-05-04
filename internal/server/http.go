package server

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/danielgtaylor/huma/v2/yaml"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"oas-sandbox/internal/api"
	"oas-sandbox/internal/api/system"
)

type HTTPServer struct {
	server *http.Server
}

func NewHTTPServer(port int, serviceName string) *HTTPServer {
	mux := http.NewServeMux()
	humaConfig := huma.DefaultConfig("OAS Sandbox", "0.1.0")
	humaConfig.Info.Description = "OpenAPI sandbox for homelab API experiments."
	humaConfig.Tags = []*huma.Tag{
		{Name: api.TagSystem, Description: "Service health and operational endpoints."},
	}
	humaConfig.DocsPath = ""
	humaConfig.OpenAPIPath = ""

	humaAPI := humago.New(mux, humaConfig)
	registerAPI(humaAPI, api.Deps{
		ServiceName: serviceName,
	})
	registerOpenAPISpec(mux, humaAPI)
	registerSwaggerUIAssets(mux)
	registerDocs(mux, humaConfig.Info.Title)

	handler := otelhttp.NewHandler(mux, serviceName)
	return &HTTPServer{
		server: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: handler,
		},
	}
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
