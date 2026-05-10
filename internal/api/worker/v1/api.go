package workerv1

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	humautil "github.com/kitti12911/lib-util/v3/huma"

	"oas-sandbox/internal/api"
)

func Register(h huma.API, deps api.Deps) {
	client := deps.WorkerClient

	huma.Post(h, "/worker/jobs", func(ctx context.Context, input *SubmitJobInput) (*SubmitJobOutput, error) {
		req, err := submitJobToProto(input)
		if err != nil {
			return nil, err
		}

		resp, err := client.SubmitJob(ctx, req)
		if err != nil {
			return nil, humautil.GRPCError(err)
		}

		return submitJobFromProto(resp), nil
	}, humautil.WithTag(api.TagWorker), statusAccepted)
}

func statusAccepted(op *huma.Operation) {
	op.DefaultStatus = http.StatusAccepted
}
