package api

import (
	sagav1 "oas-sandbox/gen/grpc/saga/v1"
	userv1 "oas-sandbox/gen/grpc/user/v1"
	workerv1 "oas-sandbox/gen/grpc/worker/v1"
)

// Deps bundles runtime collaborators that handlers need.
type Deps struct {
	ServiceName  string
	UserClient   userv1.UserServiceClient
	WorkerClient workerv1.WorkerServiceClient
	SagaClient   sagav1.SagaServiceClient
}
