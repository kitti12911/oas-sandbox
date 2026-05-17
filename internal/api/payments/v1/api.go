package paymentsv1

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	humautil "github.com/kitti12911/lib-util/v3/huma"

	sagapb "oas-sandbox/gen/grpc/saga/v1"
	"oas-sandbox/internal/api"
)

// Register mounts the payments HTTP front door routes onto h.
//
// POST /payments         — start a payment saga (Idempotency-Key header required).
// GET  /payments/{sagaId} — fetch saga state.
func Register(h huma.API, deps api.Deps) {
	client := deps.SagaClient

	huma.Post(h, "/payments", func(ctx context.Context, input *StartPaymentInput) (*StartPaymentOutput, error) {
		resp, err := client.StartPayment(ctx, startPaymentToProto(input))
		if err != nil {
			return nil, humautil.GRPCError(err)
		}
		return startPaymentFromProto(resp), nil
	}, humautil.WithTag(api.TagPayments), humautil.StatusCreated)

	huma.Get(h, "/payments/{sagaId}", func(ctx context.Context, input *GetPaymentInput) (*GetPaymentOutput, error) {
		resp, err := client.GetSaga(ctx, &sagapb.GetSagaRequest{SagaId: input.SagaID})
		if err != nil {
			return nil, humautil.GRPCError(err)
		}
		return getPaymentFromProto(resp), nil
	}, humautil.WithTag(api.TagPayments))
}
