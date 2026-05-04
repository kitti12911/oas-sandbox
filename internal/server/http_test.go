package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPServerHealth(t *testing.T) {
	srv := NewHTTPServer(0, "oas-sandbox")
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	srv.server.Handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)
}

func TestHTTPServerOpenAPI(t *testing.T) {
	srv := NewHTTPServer(0, "oas-sandbox")
	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	rec := httptest.NewRecorder()

	srv.server.Handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"title":"OAS Sandbox"`)
}

func TestDocsServesSwaggerUIOfflineAndAllowsDownloads(t *testing.T) {
	srv := NewHTTPServer(0, "oas-sandbox")
	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	rec := httptest.NewRecorder()

	srv.server.Handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	csp := rec.Header().Get("Content-Security-Policy")
	assert.Contains(t, csp, "allow-downloads")
	assert.NotContains(t, csp, "unpkg.com")

	body := rec.Body.String()
	assert.NotContains(t, body, "unpkg.com")
	assert.Contains(t, body, `href="/assets/swagger-ui/swagger-ui.css"`)
	assert.Contains(t, body, `href="/assets/swagger-ui/docs-overrides.css"`)
	assert.Contains(t, body, `src="/assets/swagger-ui/swagger-ui-bundle.js"`)
	assert.Contains(t, body, `src="/assets/swagger-ui/swagger-initializer.js"`)
	assert.Contains(t, body, `data-url="/openapi.json"`)
}

func TestSwaggerUIAssetsServed(t *testing.T) {
	srv := NewHTTPServer(0, "oas-sandbox")

	tests := []struct {
		path      string
		minLength int
	}{
		{path: "/assets/swagger-ui/swagger-ui.css", minLength: 100_000},
		{path: "/assets/swagger-ui/docs-overrides.css", minLength: 100},
		{path: "/assets/swagger-ui/swagger-ui-bundle.js", minLength: 1_000_000},
		{path: "/assets/swagger-ui/swagger-initializer.js", minLength: 100},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			srv.server.Handler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.GreaterOrEqual(t, rec.Body.Len(), tt.minLength)
		})
	}
}

func TestOpenAPIDownload(t *testing.T) {
	srv := NewHTTPServer(0, "oas-sandbox")

	tests := []struct {
		path        string
		contentType string
		filename    string
		bodyMustHas string
	}{
		{
			path:        "/openapi.json/download",
			contentType: "application/openapi+json",
			filename:    "openapi.json",
			bodyMustHas: `"openapi"`,
		},
		{
			path:        "/openapi.yaml/download",
			contentType: "application/openapi+yaml",
			filename:    "openapi.yaml",
			bodyMustHas: "openapi:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			srv.server.Handler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, tt.contentType, rec.Header().Get("Content-Type"))
			assert.Equal(t, `attachment; filename="`+tt.filename+`"`, rec.Header().Get("Content-Disposition"))
			assert.True(t, strings.Contains(rec.Body.String(), tt.bodyMustHas))
		})
	}
}
