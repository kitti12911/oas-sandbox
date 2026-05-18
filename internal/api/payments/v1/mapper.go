package paymentsv1

import (
	sagapb "oas-sandbox/gen/grpc/saga/v1"
)

func startPaymentToProto(input *StartPaymentInput) *sagapb.StartPaymentRequest {
	return &sagapb.StartPaymentRequest{
		IdempotencyKey: input.IdempotencyKey,
		AccountId:      input.Body.AccountID,
		Amount:         input.Body.Amount,
		Currency:       input.Body.Currency,
	}
}

func startPaymentFromProto(resp *sagapb.StartPaymentResponse) *StartPaymentOutput {
	if resp == nil {
		return &StartPaymentOutput{}
	}
	return &StartPaymentOutput{
		Body: StartPaymentResult{
			SagaID: resp.GetSagaId(),
			State:  resp.GetState(),
		},
	}
}

func getPaymentFromProto(resp *sagapb.GetSagaResponse) *GetPaymentOutput {
	if resp == nil {
		return &GetPaymentOutput{}
	}
	return &GetPaymentOutput{
		Body: GetPaymentResult{
			SagaID:      resp.GetSagaId(),
			State:       resp.GetState(),
			CurrentStep: resp.GetCurrentStep(),
			LastError:   resp.GetLastError(),
		},
	}
}
