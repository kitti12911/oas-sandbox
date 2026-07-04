package usersv1

import (
	"math"
	"testing"
	"time"

	commonv1 "oas-sandbox/gen/grpc/common/v1"
	userv1 "oas-sandbox/gen/grpc/user/v1"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ptr[T any](v T) *T { return &v }

func TestInt32FromInt(t *testing.T) {
	t.Parallel()
	assert.Equal(t, int32(0), int32FromInt(0))
	assert.Equal(t, int32(42), int32FromInt(42))
	assert.Equal(t, int32(math.MaxInt32), int32FromInt(math.MaxInt32+1))
	assert.Equal(t, int32(math.MinInt32), int32FromInt(math.MinInt32-1))
}

func TestUserStatusFromProto(t *testing.T) {
	t.Parallel()
	tests := map[userv1.UserStatus]string{
		userv1.UserStatus_USER_STATUS_ACTIVE:      "active",
		userv1.UserStatus_USER_STATUS_DISABLED:    "disabled",
		userv1.UserStatus_USER_STATUS_PENDING:     "pending",
		userv1.UserStatus_USER_STATUS_UNSPECIFIED: "",
	}
	for in, want := range tests {
		assert.Equal(t, want, userStatusFromProto(in))
	}
}

func TestToProtoUserStatus(t *testing.T) {
	t.Parallel()
	tests := map[string]userv1.UserStatus{
		"active":   userv1.UserStatus_USER_STATUS_ACTIVE,
		"disabled": userv1.UserStatus_USER_STATUS_DISABLED,
		"pending":  userv1.UserStatus_USER_STATUS_PENDING,
		"":         userv1.UserStatus_USER_STATUS_UNSPECIFIED,
		"unknown":  userv1.UserStatus_USER_STATUS_UNSPECIFIED,
	}
	for in, want := range tests {
		assert.Equal(t, want, toProtoUserStatus(in))
	}
}

func TestListUsersOutputFromProtoNil(t *testing.T) {
	t.Parallel()
	got := listUsersOutputFromProto(nil)
	assert.NotNil(t, got)
	assert.Empty(t, got.Body.Users)
	assert.Equal(t, 0, got.Body.Page)
}

func TestListUsersOutputFromProtoCopiesPaginationAndUsers(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	resp := &userv1.ListUsersResponse{
		Users: []*userv1.User{
			{Id: "u1", Email: "a@x", Username: "a", Status: userv1.UserStatus_USER_STATUS_ACTIVE, CreatedAt: timestamppb.New(now), UpdatedAt: timestamppb.New(now)},
		},
		Pagination: &commonv1.PaginationResponse{Page: 2, PageSize: 10, TotalPages: 3, TotalSize: 25},
	}
	got := listUsersOutputFromProto(resp)
	assert.Len(t, got.Body.Users, 1)
	assert.Equal(t, "u1", got.Body.Users[0].ID)
	assert.Equal(t, 2, got.Body.Page)
	assert.Equal(t, 10, got.Body.PageSize)
	assert.Equal(t, 3, got.Body.TotalPages)
	assert.Equal(t, 25, got.Body.TotalSize)
}

func TestUserFromProtoNil(t *testing.T) {
	t.Parallel()
	assert.Equal(t, User{}, userFromProto(nil))
}

func TestUserFromProtoCopiesFields(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	got := userFromProto(&userv1.User{
		Id:          "u1",
		Email:       "kit@example.com",
		Username:    "kit",
		DisplayName: ptr("Kit"),
		Status:      userv1.UserStatus_USER_STATUS_ACTIVE,
		CreatedAt:   timestamppb.New(now),
		UpdatedAt:   timestamppb.New(now),
		Profile: &userv1.UserProfile{
			FirstName: ptr("Kit"),
			Address:   &userv1.UserAddress{City: ptr("Bangkok")},
		},
	})
	assert.Equal(t, "u1", got.ID)
	assert.Equal(t, "kit@example.com", got.Email)
	assert.Equal(t, "Kit", *got.DisplayName)
	assert.Equal(t, "active", got.Status)
	assert.Equal(t, now, got.CreatedAt)
	assert.Equal(t, "Kit", *got.Profile.FirstName)
	assert.Equal(t, "Bangkok", *got.Profile.Address.City)
}

func TestUserToProto(t *testing.T) {
	t.Parallel()
	req := CreateUserRequest{
		Email:    "kit@example.com",
		Username: "kit",
		Status:   "pending",
		Profile: &CreateProfile{
			FirstName: ptr("Kit"),
			Address:   &CreateAddress{City: ptr("Bangkok")},
		},
	}

	got := userToProto(req)
	assert.Equal(t, "kit@example.com", got.Email)
	assert.Equal(t, userv1.UserStatus_USER_STATUS_PENDING, got.Status)
	assert.Equal(t, "Kit", *got.Profile.FirstName)
	assert.Equal(t, "Bangkok", *got.Profile.Address.City)
}

func TestProfileAndAddressToProtoNil(t *testing.T) {
	t.Parallel()
	assert.Nil(t, profileToProto(nil))
	assert.Nil(t, addressToProto(nil))
}

func TestProfileAndAddressFromProtoNil(t *testing.T) {
	t.Parallel()
	assert.Nil(t, profileFromProto(nil))
	assert.Nil(t, addressFromProto(nil))
}

func TestProfileAndAddressFromProtoCopiesFields(t *testing.T) {
	t.Parallel()
	addr := addressFromProto(&userv1.UserAddress{
		Line1: ptr("1 Road"), Line2: ptr("Apt 2"), City: ptr("BKK"),
		State: ptr("BKK"), PostalCode: ptr("10110"), CountryCode: ptr("TH"),
	})
	assert.Equal(t, "1 Road", *addr.Line1)
	assert.Equal(t, "TH", *addr.CountryCode)

	profile := profileFromProto(&userv1.UserProfile{
		FirstName: ptr("Kit"), LastName: ptr("Last"), PhoneNumber: ptr("0800"),
		Address: &userv1.UserAddress{City: ptr("BKK")},
	})
	assert.Equal(t, "Kit", *profile.FirstName)
	assert.Equal(t, "Last", *profile.LastName)
	assert.Equal(t, "0800", *profile.PhoneNumber)
	assert.Equal(t, "BKK", *profile.Address.City)
}

func TestCreateUserOutputFromProto(t *testing.T) {
	t.Parallel()
	out := createUserOutputFromProto(nil)
	assert.Equal(t, "", out.Body.ID)

	out = createUserOutputFromProto(&userv1.CreateUserResponse{Id: "u1"})
	assert.Equal(t, "u1", out.Body.ID)
}

func TestGetUserOutputFromProto(t *testing.T) {
	t.Parallel()
	out := getUserOutputFromProto(nil)
	assert.Equal(t, User{}, out.Body)

	out = getUserOutputFromProto(&userv1.GetUserResponse{User: &userv1.User{Id: "u1"}})
	assert.Equal(t, "u1", out.Body.ID)
}

func TestPaginationFromInput(t *testing.T) {
	t.Parallel()
	got := paginationFromInput(&ListUsersInput{Page: 2, PageSize: 25})
	assert.Equal(t, int32(2), got.Page)
	assert.Equal(t, int32(25), got.PageSize)
}

func TestFilterFromInputEmpty(t *testing.T) {
	t.Parallel()
	assert.Nil(t, filterFromInput(&ListUsersInput{}))
}

func TestFilterFromInputSingleValue(t *testing.T) {
	t.Parallel()
	got := filterFromInput(&ListUsersInput{
		FilterCol: "username", FilterOp: "like_ci", FilterVal: "kit",
	})
	assert.Equal(t, "username", got.Col)
	assert.Equal(t, "kit", got.Val)
	assert.Empty(t, got.Vals)
	assert.Empty(t, got.Filters)
}

func TestFilterFromInputMultipleValues(t *testing.T) {
	t.Parallel()
	got := filterFromInput(&ListUsersInput{
		FilterCol: "status", FilterOp: "in", FilterVals: "active,pending",
	})
	assert.Equal(t, []string{"active", "pending"}, got.Vals)
}

func TestOrderByFromInput(t *testing.T) {
	t.Parallel()
	assert.Nil(t, orderByFromInput(&ListUsersInput{}))
	got := orderByFromInput(&ListUsersInput{OrderBy: "username", Order: "desc"})
	assert.Len(t, got, 1)
	assert.Equal(t, "username", got[0].Col)
}

func TestPaginationFromAdvancedInput(t *testing.T) {
	t.Parallel()
	assert.Nil(t, paginationFromAdvancedInput(&AdvancedListUsersInput{}))

	got := paginationFromAdvancedInput(&AdvancedListUsersInput{
		Body: AdvancedListUsersRequest{Pagination: &Pagination{Page: 1, PageSize: 50}},
	})
	assert.Equal(t, int32(1), got.Page)
	assert.Equal(t, int32(50), got.PageSize)
}

func TestFilterFromAdvancedInput(t *testing.T) {
	t.Parallel()
	assert.Nil(t, filterFromAdvancedInput(&AdvancedListUsersInput{}))

	// leaf
	got := filterFromAdvancedInput(&AdvancedListUsersInput{
		Body: AdvancedListUsersRequest{
			Filter: &Filter{Col: "username", Op: "like_ci", Val: "kit"},
		},
	})
	assert.Equal(t, "username", got.Col)
	assert.Empty(t, got.Filters)
}

func TestFilterToProtoGroup(t *testing.T) {
	t.Parallel()
	// age >= 18 AND (status = active OR status = pending)
	got := filterToProto(&Filter{
		Logic: "and",
		Filters: []Filter{
			{Col: "age", Op: "gte", Val: "18"},
			{
				Logic: "or",
				Filters: []Filter{
					{Col: "status", Op: "exact", Val: "active"},
					{Col: "status", Op: "exact", Val: "pending"},
				},
			},
		},
	})
	assert.Equal(t, commonv1.LogicalOp_LOGICAL_OP_AND, got.Logic)
	assert.Len(t, got.Filters, 2)
	assert.Equal(t, "age", got.Filters[0].Col)
	assert.Equal(t, commonv1.LogicalOp_LOGICAL_OP_OR, got.Filters[1].Logic)
	assert.Len(t, got.Filters[1].Filters, 2)
}

func TestFilterToProtoDropsEmptyNodes(t *testing.T) {
	t.Parallel()
	// empty-col leaves are dropped; a group with no surviving children collapses
	assert.Nil(t, filterToProto(&Filter{}))
	assert.Nil(t, filterToProto(&Filter{
		Logic:   "or",
		Filters: []Filter{{Col: "", Val: "skip-me"}},
	}))

	got := filterToProto(&Filter{
		Logic: "or",
		Filters: []Filter{
			{Col: "username", Op: "like_ci", Val: "kit"},
			{Col: "", Op: "exact", Val: "skip-me"},
		},
	})
	assert.Len(t, got.Filters, 1)
	assert.Equal(t, "username", got.Filters[0].Col)
}

func TestOrderByFromAdvancedInput(t *testing.T) {
	t.Parallel()
	assert.Nil(t, orderByFromAdvancedInput(&AdvancedListUsersInput{}))

	got := orderByFromAdvancedInput(&AdvancedListUsersInput{
		Body: AdvancedListUsersRequest{
			OrderBy: []OrderBy{
				{Col: "username", Order: "asc"},
				{Col: "", Order: "desc"}, // skipped
			},
		},
	})
	assert.Len(t, got, 1)
	assert.Equal(t, "username", got[0].Col)
}
