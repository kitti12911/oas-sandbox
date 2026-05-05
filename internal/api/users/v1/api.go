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

	huma.Get(h, "/users/mock", func(ctx context.Context, input *struct{}) (*MockUserOutput, error) {
		return &MockUserOutput{
			Body: MockUser{
				ID:       "mock-user-v1",
				Username: "mock",
				Status:   "active",
			},
		}, nil
	}, humautil.WithTag(api.TagUsers))

	huma.Get(h, "/users/{id}", func(ctx context.Context, input *GetUserInput) (*UserOutput, error) {
		resp, err := client.GetUser(ctx, &userv1.GetUserRequest{Id: input.ID})
		if err != nil {
			return nil, humautil.GRPCError(err)
		}

		return &UserOutput{Body: userFromProto(resp.GetUser())}, nil
	}, humautil.WithTag(api.TagUsers))

	huma.Put(h, "/users/{id}", func(ctx context.Context, input *UpdateUserInput) (*humautil.AffectedRowsOutput, error) {
		resp, err := client.UpdateUser(ctx, &userv1.UpdateUserRequest{
			Id:   input.ID,
			User: userToUpdateProto(input),
		})
		if err != nil {
			return nil, humautil.GRPCError(err)
		}

		return humautil.AffectedRows(resp.GetAffectedRows()), nil
	}, humautil.WithTag(api.TagUsers))

	huma.Patch(h, "/users/{id}", func(ctx context.Context, input *PatchUserInput) (*humautil.AffectedRowsOutput, error) {
		user, updateMask := patchUserPatch(input)
		if len(updateMask.GetPaths()) == 0 {
			return nil, huma.Error400BadRequest("at least one field is required")
		}

		resp, err := client.PatchUser(ctx, &userv1.PatchUserRequest{
			Id:         input.ID,
			User:       user,
			UpdateMask: updateMask,
		})
		if err != nil {
			return nil, humautil.GRPCError(err)
		}

		return humautil.AffectedRows(resp.GetAffectedRows()), nil
	}, humautil.WithTag(api.TagUsers))

	huma.Delete(h, "/users/{id}", func(ctx context.Context, input *DeleteUserInput) (*humautil.AffectedRowsOutput, error) {
		resp, err := client.DeleteUser(ctx, &userv1.DeleteUserRequest{Id: input.ID})
		if err != nil {
			return nil, humautil.GRPCError(err)
		}

		return humautil.AffectedRows(resp.GetAffectedRows()), nil
	}, humautil.WithTag(api.TagUsers))
}

func RegisterMockV2(h huma.API) {
	huma.Get(h, "/users/mock", func(ctx context.Context, input *struct{}) (*MockUserOutput, error) {
		return &MockUserOutput{
			Body: MockUser{
				ID:       "mock-user-v2",
				Username: "mock",
				Status:   "active",
			},
		}, nil
	}, humautil.WithTag(api.TagUsers))
}
