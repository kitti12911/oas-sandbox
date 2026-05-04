package server

import "embed"

// swaggerUIFS embeds Swagger UI assets pinned to 5.32.5 so /docs can render
// without fetching external CDN assets.
//
//go:embed assets/swagger-ui/swagger-ui.css assets/swagger-ui/docs-overrides.css assets/swagger-ui/swagger-ui-bundle.js assets/swagger-ui/swagger-initializer.js
var swaggerUIFS embed.FS
