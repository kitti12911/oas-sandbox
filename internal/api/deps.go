package api

import userv1 "oas-sandbox/gen/grpc/user/v1"

// Deps bundles runtime collaborators that handlers need.
type Deps struct {
	ServiceName string
	UserClient  userv1.UserServiceClient
}
