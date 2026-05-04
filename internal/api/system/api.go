package system

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	humautil "github.com/kitti12911/lib-util/v3/huma"

	"oas-sandbox/internal/api"
)

func Register(h huma.API, deps api.Deps) {
	huma.Get(h, "/health", func(_ context.Context, _ *struct{}) (*HealthOutput, error) {
		out := &HealthOutput{}
		out.Body.Status = "ok"
		out.Body.Service = deps.ServiceName
		return out, nil
	}, humautil.WithTag(api.TagSystem))
}
