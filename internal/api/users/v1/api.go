package usersv1

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	humautil "github.com/kitti12911/lib-util/v3/huma"

	userv1 "oas-sandbox/gen/grpc/user/v1"
	"oas-sandbox/internal/api"
)

func Register(h huma.API, deps api.Deps) {
	client := deps.UserClient

	huma.Get(h, "/users", func(ctx context.Context, input *ListUsersInput) (*UserListOutput, error) {
		resp, err := client.ListUsers(ctx, &userv1.ListUsersRequest{
			Pagination: paginationFromInput(input),
			Filters:    filtersFromInput(input),
			OrderBy:    orderByFromInput(input),
		})
		if err != nil {
			return nil, humautil.GRPCError(err)
		}

		return userListFromProto(resp), nil
	}, humautil.WithTag(api.TagUsers))

	huma.Post(h, "/users/search", func(ctx context.Context, input *AdvancedListUsersInput) (*UserListOutput, error) {
		resp, err := client.ListUsers(ctx, &userv1.ListUsersRequest{
			Pagination: paginationFromAdvancedInput(input),
			Filters:    filtersFromAdvancedInput(input),
			OrderBy:    orderByFromAdvancedInput(input),
		})
		if err != nil {
			return nil, humautil.GRPCError(err)
		}

		return userListFromProto(resp), nil
	}, humautil.WithTag(api.TagUsers))

	huma.Post(h, "/users", func(ctx context.Context, input *CreateUserInput) (*CreateUserOutput, error) {
		resp, err := client.CreateUser(ctx, &userv1.CreateUserRequest{User: userToCreateProto(input)})
		if err != nil {
			return nil, humautil.GRPCError(err)
		}

		return createUserFromProto(resp), nil
	}, humautil.WithTag(api.TagUsers), humautil.StatusCreated)

	huma.Get(h, "/users/{id}", func(ctx context.Context, input *GetUserInput) (*UserOutput, error) {
		resp, err := client.GetUser(ctx, &userv1.GetUserRequest{Id: input.ID})
		if err != nil {
			return nil, humautil.GRPCError(err)
		}

		return &UserOutput{Body: userFromProto(resp.GetUser())}, nil
	}, humautil.WithTag(api.TagUsers))
}
