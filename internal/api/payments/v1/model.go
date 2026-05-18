package paymentsv1

// StartPaymentInput is the request to start a payment saga.
//
// The Idempotency-Key header is forwarded to the orchestrator and used to
// deduplicate retries at the saga_instances (saga_type, idempotency_key)
// unique index — sending the same key twice returns the same saga.
type StartPaymentInput struct {
	IdempotencyKey string `header:"Idempotency-Key" required:"true" example:"order-1001" doc:"Client-supplied idempotency key (saga_type=payment is implicit)."`
	Body           StartPaymentRequest
}

type StartPaymentRequest struct {
	AccountID string `json:"accountId" required:"true" example:"acct-42"  doc:"Account being charged."`
	Amount    int64  `json:"amount"    required:"true" example:"1999"     doc:"Amount in minor currency units (e.g. cents). Must be positive." minimum:"1"`
	Currency  string `json:"currency"  required:"true" example:"USD"      doc:"ISO 4217 currency code." minLength:"3" maxLength:"3"`
}

type StartPaymentOutput struct {
	Body StartPaymentResult
}

type StartPaymentResult struct {
	SagaID string `json:"sagaId" example:"0198f8f0-0000-7000-8000-000000000001" doc:"Saga instance ID. Use it with GET /payments/{sagaId} to poll state."`
	State  string `json:"state"  example:"RUNNING"                              doc:"Saga state at acceptance time."`
}

type GetPaymentInput struct {
	SagaID string `path:"sagaId" required:"true" example:"0198f8f0-0000-7000-8000-000000000001" doc:"Saga instance ID returned from POST /payments."`
}

type GetPaymentOutput struct {
	Body GetPaymentResult
}

type GetPaymentResult struct {
	SagaID      string `json:"sagaId"             example:"0198f8f0-0000-7000-8000-000000000001" doc:"Saga instance ID."`
	State       string `json:"state"              example:"COMPLETED"                            doc:"Current saga state."`
	CurrentStep int32  `json:"currentStep"        example:"2"                                    doc:"Index of the step that just ran (or is running)."`
	LastError   string `json:"lastError,omitempty" example:""                                    doc:"Last recorded error, if any."`
}
