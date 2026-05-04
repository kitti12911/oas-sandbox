package usersv1

import (
	"github.com/kitti12911/lib-util/v3/protoutil"

	userv1 "oas-sandbox/gen/grpc/user/v1"
)

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
