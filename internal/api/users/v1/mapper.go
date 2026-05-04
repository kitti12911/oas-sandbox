package usersv1

import (
	"strings"

	"github.com/kitti12911/lib-util/v3/protoutil"
	"github.com/kitti12911/lib-util/v3/query"

	commonv1 "oas-sandbox/gen/grpc/common/v1"
	userv1 "oas-sandbox/gen/grpc/user/v1"
)

func userListFromProto(resp *userv1.ListUsersResponse) *UserListOutput {
	if resp == nil {
		return &UserListOutput{}
	}

	users := make([]User, 0, len(resp.GetUsers()))
	for _, user := range resp.GetUsers() {
		users = append(users, userFromProto(user))
	}

	pagination := resp.GetPagination()
	return &UserListOutput{
		Body: UserList{
			Users:      users,
			Page:       int(pagination.GetPage()),
			PageSize:   int(pagination.GetPageSize()),
			TotalPages: int(pagination.GetTotalPages()),
			TotalSize:  int(pagination.GetTotalSize()),
		},
	}
}

func userFromProto(user *userv1.User) User {
	if user == nil {
		return User{}
	}

	out := User{
		ID:        user.GetId(),
		Email:     user.GetEmail(),
		Username:  user.GetUsername(),
		Status:    statusFromProto(user.GetStatus()),
		CreatedAt: protoutil.TimeFromProto(user.GetCreatedAt()),
		UpdatedAt: protoutil.TimeFromProto(user.GetUpdatedAt()),
	}
	if user.DisplayName != nil {
		out.DisplayName = user.DisplayName
	}
	if user.GetProfile() != nil {
		out.Profile = profileFromProto(user.GetProfile())
	}

	return out
}

func userToCreateProto(input *CreateUserInput) *userv1.User {
	return userRequestToProto(input.Body)
}

func userToUpdateProto(input *UpdateUserInput) *userv1.User {
	return userRequestToProto(input.Body)
}

func userRequestToProto(user CreateUserRequest) *userv1.User {
	return &userv1.User{
		Email:       user.Email,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Status:      statusToProto(user.Status),
		Profile:     profileToCreateProto(user.Profile),
	}
}

func profileToCreateProto(profile *CreateProfile) *userv1.UserProfile {
	if profile == nil {
		return nil
	}

	return &userv1.UserProfile{
		FirstName:   profile.FirstName,
		LastName:    profile.LastName,
		PhoneNumber: profile.PhoneNumber,
		Address:     addressToCreateProto(profile.Address),
	}
}

func addressToCreateProto(address *CreateAddress) *userv1.UserAddress {
	if address == nil {
		return nil
	}

	return &userv1.UserAddress{
		Line1:       address.Line1,
		Line2:       address.Line2,
		City:        address.City,
		State:       address.State,
		PostalCode:  address.PostalCode,
		CountryCode: address.CountryCode,
	}
}

func profileFromProto(profile *userv1.UserProfile) *Profile {
	if profile == nil {
		return nil
	}

	out := &Profile{
		FirstName:   profile.FirstName,
		LastName:    profile.LastName,
		PhoneNumber: profile.PhoneNumber,
	}
	if profile.GetAddress() != nil {
		out.Address = addressFromProto(profile.GetAddress())
	}

	return out
}

func addressFromProto(address *userv1.UserAddress) *Address {
	if address == nil {
		return nil
	}

	return &Address{
		Line1:       address.Line1,
		Line2:       address.Line2,
		City:        address.City,
		State:       address.State,
		PostalCode:  address.PostalCode,
		CountryCode: address.CountryCode,
	}
}

func createUserFromProto(resp *userv1.CreateUserResponse) *CreateUserOutput {
	if resp == nil {
		return &CreateUserOutput{}
	}

	return &CreateUserOutput{
		Body: CreateUserResult{
			ID: resp.GetId(),
		},
	}
}

func statusFromProto(status userv1.UserStatus) string {
	switch status {
	case userv1.UserStatus_USER_STATUS_ACTIVE:
		return "active"
	case userv1.UserStatus_USER_STATUS_DISABLED:
		return "disabled"
	case userv1.UserStatus_USER_STATUS_PENDING:
		return "pending"
	default:
		return "unspecified"
	}
}

func statusToProto(status string) userv1.UserStatus {
	switch status {
	case "active":
		return userv1.UserStatus_USER_STATUS_ACTIVE
	case "disabled":
		return userv1.UserStatus_USER_STATUS_DISABLED
	case "pending":
		return userv1.UserStatus_USER_STATUS_PENDING
	default:
		return userv1.UserStatus_USER_STATUS_UNSPECIFIED
	}
}

func paginationFromInput(input *ListUsersInput) *commonv1.PaginationRequest {
	return &commonv1.PaginationRequest{
		Page:     int32(input.Page),
		PageSize: int32(input.PageSize),
	}
}

func filtersFromInput(input *ListUsersInput) []*commonv1.Filter {
	if input.FilterCol == "" {
		return nil
	}

	var vals []string
	if input.FilterVals != "" {
		vals = strings.Split(input.FilterVals, ",")
	}

	return []*commonv1.Filter{
		{
			Col:  input.FilterCol,
			Op:   query.FilterOpFromString[commonv1.FilterOp](input.FilterOp),
			Val:  input.FilterVal,
			Vals: vals,
		},
	}
}

func orderByFromInput(input *ListUsersInput) []*commonv1.OrderBy {
	if input.OrderBy == "" {
		return nil
	}

	return []*commonv1.OrderBy{
		{
			Col:   input.OrderBy,
			Order: query.OrderDirectionFromString[commonv1.OrderDirection](input.Order),
		},
	}
}

func paginationFromAdvancedInput(input *AdvancedListUsersInput) *commonv1.PaginationRequest {
	if input.Body.Pagination == nil {
		return nil
	}

	return &commonv1.PaginationRequest{
		Page:     int32(input.Body.Pagination.Page),
		PageSize: int32(input.Body.Pagination.PageSize),
	}
}

func filtersFromAdvancedInput(input *AdvancedListUsersInput) []*commonv1.Filter {
	if len(input.Body.Filters) == 0 {
		return nil
	}

	filters := make([]*commonv1.Filter, 0, len(input.Body.Filters))
	for _, filter := range input.Body.Filters {
		if filter.Col == "" {
			continue
		}

		filters = append(filters, &commonv1.Filter{
			Col:  filter.Col,
			Op:   query.FilterOpFromString[commonv1.FilterOp](filter.Op),
			Val:  filter.Val,
			Vals: filter.Vals,
		})
	}

	return filters
}

func orderByFromAdvancedInput(input *AdvancedListUsersInput) []*commonv1.OrderBy {
	if len(input.Body.OrderBy) == 0 {
		return nil
	}

	orderBy := make([]*commonv1.OrderBy, 0, len(input.Body.OrderBy))
	for _, order := range input.Body.OrderBy {
		if order.Col == "" {
			continue
		}

		orderBy = append(orderBy, &commonv1.OrderBy{
			Col:   order.Col,
			Order: query.OrderDirectionFromString[commonv1.OrderDirection](order.Order),
		})
	}

	return orderBy
}
