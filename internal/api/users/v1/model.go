package usersv1

import (
	"time"
)

type GetUserInput struct {
	ID string `path:"id" example:"0198f8f0-0000-7000-8000-000000000001" doc:"User ID"`
}

type UserOutput struct {
	Body User
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
