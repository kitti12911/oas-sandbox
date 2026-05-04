package usersv1

import (
	"time"
)

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
	Val  string   `json:"val,omitempty"  example:"new"      doc:"Single filter value"`
	Vals []string `json:"vals,omitempty" example:"active"   doc:"Multiple filter values"`
}

type OrderBy struct {
	Col   string `json:"col"   example:"username" doc:"Order field"`
	Order string `json:"order" example:"asc"      doc:"Order direction: asc or desc"`
}

type CreateUserInput struct {
	Body CreateUserRequest
}

type CreateUserRequest struct {
	Email       string         `json:"email"                 example:"kitti@example.com" doc:"Email address"`
	Username    string         `json:"username"              example:"kitti"             doc:"Username"`
	DisplayName *string        `json:"displayName,omitempty" example:"Kitti"             doc:"Display name"`
	Status      string         `json:"status"                example:"active"            doc:"User status"`
	Profile     *CreateProfile `json:"profile,omitempty"                               doc:"User profile"`
}

type CreateProfile struct {
	FirstName   *string        `json:"firstName,omitempty"   example:"Kitti"  doc:"First name"`
	LastName    *string        `json:"lastName,omitempty"    example:"User"   doc:"Last name"`
	PhoneNumber *string        `json:"phoneNumber,omitempty" example:"+66000" doc:"Phone number"`
	Address     *CreateAddress `json:"address,omitempty"                    doc:"Address"`
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
