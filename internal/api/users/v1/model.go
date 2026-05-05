package usersv1

import (
	"time"

	humautil "github.com/kitti12911/lib-util/v3/huma"

	userv1 "oas-sandbox/gen/grpc/user/v1"
)

var _ userv1.User

type GetUserInput struct {
	ID string `path:"id" example:"0198f8f0-0000-7000-8000-000000000001" doc:"User ID"`
}

type ListUsersInput struct {
	Page       int    `query:"page"       example:"1"        doc:"Page number"`
	PageSize   int    `query:"pageSize"   example:"10"       doc:"Items per page"`
	FilterCol  string `query:"filterCol"  example:"username" doc:"Filter field"`
	FilterOp   string `query:"filterOp"   example:"like_ci"  doc:"Filter operation: exact, like, like_ci, gt, lt, gte, lte, null, not_null, in, between, between_exclusive"`
	FilterVal  string `query:"filterVal"  example:"kit"      doc:"Single filter value"`
	FilterVals string `query:"filterVals" example:"active,pending" doc:"Comma-separated multi values for in and between operations"`
	OrderBy    string `query:"orderBy"    example:"username" doc:"Order field"`
	Order      string `query:"order"      example:"asc"      doc:"Order direction: asc or desc"`
}

type AdvancedListUsersInput struct {
	Body AdvancedListUsersRequest
}

type AdvancedListUsersRequest struct {
	Pagination *Pagination `json:"pagination,omitempty" doc:"Pagination options"`
	Filters    []Filter    `json:"filters,omitempty"    doc:"Filter clauses"`
	OrderBy    []OrderBy   `json:"orderBy,omitempty"    doc:"Order clauses"`
}

type Pagination struct {
	Page     int `json:"page,omitempty"     example:"1"  doc:"Page number"`
	PageSize int `json:"pageSize,omitempty" example:"10" doc:"Items per page"`
}

type Filter struct {
	Col  string   `json:"col"            example:"username" doc:"Filter field"`
	Op   string   `json:"op"             example:"like_ci"  doc:"Filter operation"`
	Val  string   `json:"val,omitempty"  example:"kit"      doc:"Single filter value"`
	Vals []string `json:"vals,omitempty" example:"active"   doc:"Multiple filter values"`
}

type OrderBy struct {
	Col   string `json:"col"   example:"username" doc:"Order field"`
	Order string `json:"order" example:"asc"      doc:"Order direction: asc or desc"`
}

type CreateUserInput struct {
	Body CreateUserRequest
}

type UpdateUserInput struct {
	ID string `path:"id" example:"0198f8f0-0000-7000-8000-000000000001" doc:"User ID"`

	Body CreateUserRequest
}

type PatchUserInput struct {
	ID string `path:"id" example:"0198f8f0-0000-7000-8000-000000000001" doc:"User ID"`

	Body PatchUserRequest
}

type DeleteUserInput struct {
	ID string `path:"id" example:"0198f8f0-0000-7000-8000-000000000001" doc:"User ID"`
}

type MockUserOutput struct {
	Body MockUser
}

type MockUser struct {
	ID       string `json:"id"       example:"mock-user-v1" doc:"Mock user ID"`
	Username string `json:"username" example:"mock"         doc:"Mock username"`
}

type CreateUserRequest struct {
	Email       string         `json:"email"                 example:"kitti@example.com" doc:"Email address"`
	Username    string         `json:"username"              example:"kitti"             doc:"Username"`
	DisplayName *string        `json:"displayName,omitempty" example:"Kitti"             doc:"Display name"`
	Status      string         `json:"status"                example:"active"            doc:"User status"`
	Profile     *CreateProfile `json:"profile,omitempty"                               doc:"User profile"`
}

//openapi:patch proto=userv1.User
type PatchUserRequest struct {
	Email       humautil.Patch[string]       `json:"email"       required:"false" example:"kitti@example.com" doc:"Email address"`
	Username    humautil.Patch[string]       `json:"username"    required:"false" example:"kitti"             doc:"Username"`
	DisplayName humautil.Patch[string]       `json:"displayName" required:"false" example:"Kitti"             doc:"Display name" patch:"ptr"`
	Status      humautil.Patch[string]       `json:"status"      required:"false" example:"active"            doc:"User status" patch:"converter=statusToProto"`
	Profile     humautil.Patch[PatchProfile] `json:"profile"     required:"false"                           doc:"User profile" patch:"proto=UserProfile"`
}

type PatchProfile struct {
	FirstName   humautil.Patch[string]       `json:"firstName"   required:"false" example:"Kitti"  doc:"First name" patch:"ptr"`
	LastName    humautil.Patch[string]       `json:"lastName"    required:"false" example:"User"   doc:"Last name" patch:"ptr"`
	PhoneNumber humautil.Patch[string]       `json:"phoneNumber" required:"false" example:"+66000" doc:"Phone number" patch:"ptr"`
	Address     humautil.Patch[PatchAddress] `json:"address"     required:"false"                  doc:"Address" patch:"proto=UserAddress"`
}

type PatchAddress struct {
	Line1       humautil.Patch[string] `json:"line1"       required:"false" example:"123 Main St" doc:"Address line 1" patch:"ptr"`
	Line2       humautil.Patch[string] `json:"line2"       required:"false" example:"Unit 10"     doc:"Address line 2" patch:"ptr"`
	City        humautil.Patch[string] `json:"city"        required:"false" example:"Bangkok"     doc:"City" patch:"ptr"`
	State       humautil.Patch[string] `json:"state"       required:"false" example:"Bangkok"     doc:"State or province" patch:"ptr"`
	PostalCode  humautil.Patch[string] `json:"postalCode"  required:"false" example:"10110"       doc:"Postal code" patch:"ptr"`
	CountryCode humautil.Patch[string] `json:"countryCode" required:"false" example:"TH"          doc:"ISO country code" patch:"ptr"`
}

type CreateProfile struct {
	FirstName   *string        `json:"firstName,omitempty"   example:"Kitti"  doc:"First name"`
	LastName    *string        `json:"lastName,omitempty"    example:"User"   doc:"Last name"`
	PhoneNumber *string        `json:"phoneNumber,omitempty" example:"+66000" doc:"Phone number"`
	Address     *CreateAddress `json:"address,omitempty"                      doc:"Address"`
}

type CreateAddress struct {
	Line1       *string `json:"line1,omitempty"       example:"123 Main St" doc:"Address line 1"`
	Line2       *string `json:"line2,omitempty"       example:"Unit 10"     doc:"Address line 2"`
	City        *string `json:"city,omitempty"        example:"Bangkok"     doc:"City"`
	State       *string `json:"state,omitempty"       example:"Bangkok"     doc:"State or province"`
	PostalCode  *string `json:"postalCode,omitempty"  example:"10110"       doc:"Postal code"`
	CountryCode *string `json:"countryCode,omitempty" example:"TH"          doc:"ISO country code"`
}

type CreateUserOutput struct {
	Body CreateUserResult
}

type CreateUserResult struct {
	ID string `json:"id" example:"0198f8f0-0000-7000-8000-000000000001" doc:"Created user ID"`
}

type UserOutput struct {
	Body User
}

type UserListOutput struct {
	Body UserList
}

type UserList struct {
	Users      []User `json:"users"      doc:"Users for the current page"`
	Page       int    `json:"page"       example:"1"  doc:"Current page"`
	PageSize   int    `json:"pageSize"   example:"10" doc:"Items per page"`
	TotalPages int    `json:"totalPages" example:"1"  doc:"Total pages"`
	TotalSize  int    `json:"totalSize"  example:"1"  doc:"Total item count"`
}

type User struct {
	ID          string    `json:"id"          example:"0198f8f0-0000-7000-8000-000000000001" doc:"User ID"`
	Email       string    `json:"email"       example:"kitti@example.com"                     doc:"Email address"`
	Username    string    `json:"username"    example:"kitti"                                 doc:"Username"`
	DisplayName *string   `json:"displayName,omitempty" example:"Kitti"                       doc:"Display name"`
	Status      string    `json:"status"      example:"active"                                doc:"User status"`
	Profile     *Profile  `json:"profile,omitempty"                                          doc:"User profile"`
	CreatedAt   time.Time `json:"createdAt"                                                   doc:"Creation time"`
	UpdatedAt   time.Time `json:"updatedAt"                                                   doc:"Update time"`
}

type Profile struct {
	FirstName   *string  `json:"firstName,omitempty"   example:"Kitti"  doc:"First name"`
	LastName    *string  `json:"lastName,omitempty"    example:"User"   doc:"Last name"`
	PhoneNumber *string  `json:"phoneNumber,omitempty" example:"+66000" doc:"Phone number"`
	Address     *Address `json:"address,omitempty"                    doc:"Address"`
}

type Address struct {
	Line1       *string `json:"line1,omitempty"       example:"123 Main St" doc:"Address line 1"`
	Line2       *string `json:"line2,omitempty"       example:"Unit 10"     doc:"Address line 2"`
	City        *string `json:"city,omitempty"        example:"Bangkok"     doc:"City"`
	State       *string `json:"state,omitempty"       example:"Bangkok"     doc:"State or province"`
	PostalCode  *string `json:"postalCode,omitempty"  example:"10110"       doc:"Postal code"`
	CountryCode *string `json:"countryCode,omitempty" example:"TH"          doc:"ISO country code"`
}
