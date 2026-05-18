package paymentsv1

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sagapb "oas-sandbox/gen/grpc/saga/v1"
)

func TestStartPaymentToProto(t *testing.T) {
	t.Parallel()
	got := startPaymentToProto(&StartPaymentInput{
		IdempotencyKey: "order-1001",
		Body: StartPaymentRequest{
			AccountID: "acct-42",
			Amount:    1999,
			Currency:  "USD",
		},
	})
	assert.Equal(t, "order-1001", got.GetIdempotencyKey())
	assert.Equal(t, "acct-42", got.GetAccountId())
	assert.Equal(t, int64(1999), got.GetAmount())
	assert.Equal(t, "USD", got.GetCurrency())
}

func TestStartPaymentFromProtoNil(t *testing.T) {
	t.Parallel()
	got := startPaymentFromProto(nil)
	assert.Equal(t, "", got.Body.SagaID)
	assert.Equal(t, "", got.Body.State)
}

func TestStartPaymentFromProto(t *testing.T) {
	t.Parallel()
	got := startPaymentFromProto(&sagapb.StartPaymentResponse{
		SagaId: "saga-1",
		State:  "RUNNING",
	})
	assert.Equal(t, "saga-1", got.Body.SagaID)
	assert.Equal(t, "RUNNING", got.Body.State)
}

func TestGetPaymentFromProtoNil(t *testing.T) {
	t.Parallel()
	got := getPaymentFromProto(nil)
	assert.Equal(t, "", got.Body.SagaID)
}

func TestGetPaymentFromProto(t *testing.T) {
	t.Parallel()
	got := getPaymentFromProto(&sagapb.GetSagaResponse{
		SagaId:      "saga-1",
		State:       "COMPLETED",
		CurrentStep: 2,
		LastError:   "",
	})
	assert.Equal(t, "saga-1", got.Body.SagaID)
	assert.Equal(t, "COMPLETED", got.Body.State)
	assert.Equal(t, int32(2), got.Body.CurrentStep)
	assert.Equal(t, "", got.Body.LastError)
}
