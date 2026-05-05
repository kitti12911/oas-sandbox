package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonv1 "oas-sandbox/gen/grpc/common/v1"
	userv1 "oas-sandbox/gen/grpc/user/v1"
)

func newRequest(method, target string, body io.Reader) *http.Request {
	if body == nil {
		body = http.NoBody
	}
	return httptest.NewRequestWithContext(context.Background(), method, target, body)
}

func TestHTTPServerHealth(t *testing.T) {
	srv := NewHTTPServer(0, "oas-sandbox", fakeUserClient{})
	req := newRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	srv.server.Handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)
}

func TestHTTPServerOpenAPI(t *testing.T) {
	srv := NewHTTPServer(0, "oas-sandbox", fakeUserClient{})
	req := newRequest(http.MethodGet, "/openapi.json", nil)
	rec := httptest.NewRecorder()

	srv.server.Handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"title":"OAS Sandbox"`)
	assert.Contains(t, rec.Body.String(), `"/v1/users"`)
	assert.Contains(t, rec.Body.String(), `"/v1/users/{id}"`)
}

func TestDocsServesSwaggerUIOfflineAndAllowsDownloads(t *testing.T) {
	srv := NewHTTPServer(0, "oas-sandbox", fakeUserClient{})
	req := newRequest(http.MethodGet, "/docs", nil)
	rec := httptest.NewRecorder()

	srv.server.Handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	csp := rec.Header().Get("Content-Security-Policy")
	assert.Contains(t, csp, "allow-downloads")
	assert.NotContains(t, csp, "unpkg.com")

	body := rec.Body.String()
	assert.NotContains(t, body, "unpkg.com")
	assert.Contains(t, body, `href="/assets/swagger-ui/swagger-ui.css"`)
	assert.Contains(t, body, `href="/assets/swagger-ui/docs-overrides.css"`)
	assert.Contains(t, body, `src="/assets/swagger-ui/swagger-ui-bundle.js"`)
	assert.Contains(t, body, `src="/assets/swagger-ui/swagger-initializer.js"`)
	assert.Contains(t, body, `data-url="/openapi.json"`)
}

func TestSwaggerUIAssetsServed(t *testing.T) {
	srv := NewHTTPServer(0, "oas-sandbox", fakeUserClient{})

	tests := []struct {
		path      string
		minLength int
	}{
		{path: "/assets/swagger-ui/swagger-ui.css", minLength: 100_000},
		{path: "/assets/swagger-ui/docs-overrides.css", minLength: 100},
		{path: "/assets/swagger-ui/swagger-ui-bundle.js", minLength: 1_000_000},
		{path: "/assets/swagger-ui/swagger-initializer.js", minLength: 100},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := newRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			srv.server.Handler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.GreaterOrEqual(t, rec.Body.Len(), tt.minLength)
		})
	}
}

func TestOpenAPIDownload(t *testing.T) {
	srv := NewHTTPServer(0, "oas-sandbox", fakeUserClient{})

	tests := []struct {
		path        string
		contentType string
		filename    string
		bodyMustHas string
	}{
		{
			path:        "/openapi.json/download",
			contentType: "application/openapi+json",
			filename:    "openapi.json",
			bodyMustHas: `"openapi"`,
		},
		{
			path:        "/openapi.yaml/download",
			contentType: "application/openapi+yaml",
			filename:    "openapi.yaml",
			bodyMustHas: "openapi:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := newRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			srv.server.Handler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, tt.contentType, rec.Header().Get("Content-Type"))
			assert.Equal(t, `attachment; filename="`+tt.filename+`"`, rec.Header().Get("Content-Disposition"))
			assert.True(t, strings.Contains(rec.Body.String(), tt.bodyMustHas))
		})
	}
}

func TestGetUserEndpoint(t *testing.T) {
	client := fakeUserClient{}
	srv := NewHTTPServer(0, "oas-sandbox", client)
	req := newRequest(http.MethodGet, "/v1/users/0198f8f0-0000-7000-8000-000000000001", nil)
	rec := httptest.NewRecorder()

	srv.server.Handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, `"id":"0198f8f0-0000-7000-8000-000000000001"`)
	assert.Contains(t, body, `"email":"kitti@example.com"`)
	assert.Contains(t, body, `"username":"kitti"`)
	assert.Contains(t, body, `"displayName":"Kitti"`)
	assert.Contains(t, body, `"status":"active"`)
	assert.Contains(t, body, `"firstName":"Kitti"`)
	assert.Contains(t, body, `"countryCode":"TH"`)
}

func TestListUsersEndpoint(t *testing.T) {
	var got *userv1.ListUsersRequest
	client := fakeUserClient{listReq: &got}
	srv := NewHTTPServer(0, "oas-sandbox", client)
	req := newRequest(
		http.MethodGet,
		"/v1/users?page=2&pageSize=5"+
			"&filterCol=username&filterOp=like_ci&filterVal=kit"+
			"&orderBy=username&order=desc",
		nil,
	)
	rec := httptest.NewRecorder()

	srv.server.Handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, `"users":[`)
	assert.Contains(t, body, `"username":"kitti"`)
	assert.Contains(t, body, `"page":2`)
	assert.Contains(t, body, `"pageSize":5`)
	assert.Contains(t, body, `"totalPages":4`)
	assert.Contains(t, body, `"totalSize":20`)

	require.NotNil(t, got)
	assert.Equal(t, int32(2), got.GetPagination().GetPage())
	assert.Equal(t, int32(5), got.GetPagination().GetPageSize())

	require.Len(t, got.GetFilters(), 1)
	assert.Equal(t, "username", got.GetFilters()[0].GetCol())
	assert.Equal(t, commonv1.FilterOp_FILTER_OP_LIKE_CI, got.GetFilters()[0].GetOp())
	assert.Equal(t, "kit", got.GetFilters()[0].GetVal())

	require.Len(t, got.GetOrderBy(), 1)
	assert.Equal(t, "username", got.GetOrderBy()[0].GetCol())
	assert.Equal(t, commonv1.OrderDirection_ORDER_DIRECTION_DESC, got.GetOrderBy()[0].GetOrder())
}

func TestAdvancedListUsersEndpoint(t *testing.T) {
	var got *userv1.ListUsersRequest
	client := fakeUserClient{listReq: &got}
	srv := NewHTTPServer(0, "oas-sandbox", client)
	body := strings.NewReader(`{
		"pagination": {
			"page": 3,
			"pageSize": 15
		},
		"filters": [
			{
				"col": "createdAt",
				"op": "between",
				"vals": ["2026-05-03T19:52:28.202566Z", "2026-05-03T19:54:28.202566Z"]
			},
			{
				"col": "username",
				"op": "like",
				"val": "new"
			},
			{
				"col": "deletedAt",
				"op": "null"
			},
			{
				"col": "status",
				"op": "in",
				"vals": ["active", "pending"]
			}
		],
		"orderBy": [
			{
				"col": "username",
				"order": "desc"
			},
			{
				"col": "createdAt",
				"order": "asc"
			}
		]
	}`)
	req := newRequest(http.MethodPost, "/v1/users/search", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.server.Handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, got)
	assert.Equal(t, int32(3), got.GetPagination().GetPage())
	assert.Equal(t, int32(15), got.GetPagination().GetPageSize())

	require.Len(t, got.GetFilters(), 4)

	assert.Equal(t, "createdAt", got.GetFilters()[0].GetCol())
	assert.Equal(t, commonv1.FilterOp_FILTER_OP_BETWEEN, got.GetFilters()[0].GetOp())
	assert.Equal(t, []string{
		"2026-05-03T19:52:28.202566Z",
		"2026-05-03T19:54:28.202566Z",
	}, got.GetFilters()[0].GetVals())

	assert.Equal(t, "username", got.GetFilters()[1].GetCol())
	assert.Equal(t, commonv1.FilterOp_FILTER_OP_LIKE, got.GetFilters()[1].GetOp())
	assert.Equal(t, "new", got.GetFilters()[1].GetVal())

	assert.Equal(t, "deletedAt", got.GetFilters()[2].GetCol())
	assert.Equal(t, commonv1.FilterOp_FILTER_OP_NULL, got.GetFilters()[2].GetOp())
	assert.Empty(t, got.GetFilters()[2].GetVal())
	assert.Empty(t, got.GetFilters()[2].GetVals())

	assert.Equal(t, "status", got.GetFilters()[3].GetCol())
	assert.Equal(t, commonv1.FilterOp_FILTER_OP_IN, got.GetFilters()[3].GetOp())
	assert.Equal(t, []string{"active", "pending"}, got.GetFilters()[3].GetVals())

	require.Len(t, got.GetOrderBy(), 2)
	assert.Equal(t, "username", got.GetOrderBy()[0].GetCol())
	assert.Equal(t, commonv1.OrderDirection_ORDER_DIRECTION_DESC, got.GetOrderBy()[0].GetOrder())
	assert.Equal(t, "createdAt", got.GetOrderBy()[1].GetCol())
	assert.Equal(t, commonv1.OrderDirection_ORDER_DIRECTION_ASC, got.GetOrderBy()[1].GetOrder())
}

func TestCreateUserEndpoint(t *testing.T) {
	var got *userv1.CreateUserRequest
	client := fakeUserClient{createReq: &got}
	srv := NewHTTPServer(0, "oas-sandbox", client)
	body := strings.NewReader(`{
		"email": "new@example.com",
		"username": "new-user",
		"displayName": "New User",
		"status": "active",
		"profile": {
			"firstName": "New",
			"lastName": "User",
			"phoneNumber": "+66123",
			"address": {
				"line1": "123 Main St",
				"line2": "Unit 10",
				"city": "Bangkok",
				"state": "Bangkok",
				"postalCode": "10110",
				"countryCode": "TH"
			}
		}
	}`)
	req := newRequest(http.MethodPost, "/v1/users", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.server.Handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), `"id":"0198f8f0-0000-7000-8000-000000000099"`)

	require.NotNil(t, got)
	assert.Equal(t, "new@example.com", got.GetUser().GetEmail())
	assert.Equal(t, "new-user", got.GetUser().GetUsername())
	assert.Equal(t, "New User", got.GetUser().GetDisplayName())
	assert.Equal(t, userv1.UserStatus_USER_STATUS_ACTIVE, got.GetUser().GetStatus())
	assert.Equal(t, "New", got.GetUser().GetProfile().GetFirstName())
	assert.Equal(t, "User", got.GetUser().GetProfile().GetLastName())
	assert.Equal(t, "+66123", got.GetUser().GetProfile().GetPhoneNumber())
	assert.Equal(t, "123 Main St", got.GetUser().GetProfile().GetAddress().GetLine1())
	assert.Equal(t, "Unit 10", got.GetUser().GetProfile().GetAddress().GetLine2())
	assert.Equal(t, "Bangkok", got.GetUser().GetProfile().GetAddress().GetCity())
	assert.Equal(t, "Bangkok", got.GetUser().GetProfile().GetAddress().GetState())
	assert.Equal(t, "10110", got.GetUser().GetProfile().GetAddress().GetPostalCode())
	assert.Equal(t, "TH", got.GetUser().GetProfile().GetAddress().GetCountryCode())
}

func TestUpdateUserEndpoint(t *testing.T) {
	var got *userv1.UpdateUserRequest
	client := fakeUserClient{updateReq: &got}
	srv := NewHTTPServer(0, "oas-sandbox", client)
	body := strings.NewReader(`{
		"email": "updated@example.com",
		"username": "updated-user",
		"displayName": "Updated User",
		"status": "disabled",
		"profile": {
			"firstName": "Updated",
			"lastName": "User",
			"phoneNumber": "+66999",
			"address": {
				"line1": "456 Main St",
				"city": "Chiang Mai",
				"countryCode": "TH"
			}
		}
	}`)
	req := newRequest(http.MethodPut, "/v1/users/0198f8f0-0000-7000-8000-000000000001", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.server.Handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"affectedRows":1`)

	require.NotNil(t, got)
	assert.Equal(t, "0198f8f0-0000-7000-8000-000000000001", got.GetId())
	assert.Equal(t, "updated@example.com", got.GetUser().GetEmail())
	assert.Equal(t, "updated-user", got.GetUser().GetUsername())
	assert.Equal(t, "Updated User", got.GetUser().GetDisplayName())
	assert.Equal(t, userv1.UserStatus_USER_STATUS_DISABLED, got.GetUser().GetStatus())
	assert.Equal(t, "Updated", got.GetUser().GetProfile().GetFirstName())
	assert.Equal(t, "User", got.GetUser().GetProfile().GetLastName())
	assert.Equal(t, "+66999", got.GetUser().GetProfile().GetPhoneNumber())
	assert.Equal(t, "456 Main St", got.GetUser().GetProfile().GetAddress().GetLine1())
	assert.Equal(t, "Chiang Mai", got.GetUser().GetProfile().GetAddress().GetCity())
	assert.Equal(t, "TH", got.GetUser().GetProfile().GetAddress().GetCountryCode())
}

func TestPatchUserEndpoint(t *testing.T) {
	var got *userv1.PatchUserRequest
	client := fakeUserClient{patchReq: &got}
	srv := NewHTTPServer(0, "oas-sandbox", client)
	body := strings.NewReader(`{
		"username": "patched-user",
		"displayName": null,
		"profile": {
			"firstName": null,
			"address": {
				"city": "Phuket"
			}
		}
	}`)
	req := newRequest(http.MethodPatch, "/v1/users/0198f8f0-0000-7000-8000-000000000001", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.server.Handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"affectedRows":1`)

	require.NotNil(t, got)
	assert.Equal(t, "0198f8f0-0000-7000-8000-000000000001", got.GetId())
	assert.Equal(t, []string{"username", "display_name", "profile.first_name", "profile.address.city"}, got.GetUpdateMask().GetPaths())
	assert.Equal(t, "patched-user", got.GetUser().GetUsername())
	assert.Nil(t, got.GetUser().DisplayName)
	assert.Nil(t, got.GetUser().GetProfile().FirstName)
	assert.Equal(t, "Phuket", got.GetUser().GetProfile().GetAddress().GetCity())
}

func TestPatchUserEndpointRejectsEmptyBody(t *testing.T) {
	client := fakeUserClient{}
	srv := NewHTTPServer(0, "oas-sandbox", client)
	req := newRequest(http.MethodPatch, "/v1/users/0198f8f0-0000-7000-8000-000000000001", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.server.Handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "at least one field is required")
}

type fakeUserClient struct {
	listReq   **userv1.ListUsersRequest
	createReq **userv1.CreateUserRequest
	updateReq **userv1.UpdateUserRequest
	patchReq  **userv1.PatchUserRequest
}

func (fakeUserClient) GetUser(
	_ context.Context,
	req *userv1.GetUserRequest,
	_ ...grpc.CallOption,
) (*userv1.GetUserResponse, error) {
	displayName := "Kitti"
	firstName := "Kitti"
	lastName := "User"
	phoneNumber := "+66000"
	line1 := "123 Main St"
	city := "Bangkok"
	countryCode := "TH"

	return &userv1.GetUserResponse{
		User: fakeUser(req.GetId(), "kitti@example.com", "kitti", displayName, firstName, lastName, phoneNumber, line1, city, countryCode),
	}, nil
}

func (c fakeUserClient) ListUsers(
	_ context.Context,
	req *userv1.ListUsersRequest,
	_ ...grpc.CallOption,
) (*userv1.ListUsersResponse, error) {
	if c.listReq != nil {
		*c.listReq = req
	}

	return &userv1.ListUsersResponse{
		Users: []*userv1.User{
			fakeUser(
				"0198f8f0-0000-7000-8000-000000000001",
				"kitti@example.com",
				"kitti",
				"Kitti",
				"Kitti",
				"User",
				"+66000",
				"123 Main St",
				"Bangkok",
				"TH",
			),
		},
		Pagination: &commonv1.PaginationResponse{
			Page:       2,
			PageSize:   5,
			TotalPages: 4,
			TotalSize:  20,
		},
	}, nil
}

func (c fakeUserClient) CreateUser(
	_ context.Context,
	req *userv1.CreateUserRequest,
	_ ...grpc.CallOption,
) (*userv1.CreateUserResponse, error) {
	if c.createReq != nil {
		*c.createReq = req
	}

	return &userv1.CreateUserResponse{Id: "0198f8f0-0000-7000-8000-000000000099"}, nil
}

func (c fakeUserClient) UpdateUser(
	_ context.Context,
	req *userv1.UpdateUserRequest,
	_ ...grpc.CallOption,
) (*userv1.UpdateUserResponse, error) {
	if c.updateReq != nil {
		*c.updateReq = req
	}

	return &userv1.UpdateUserResponse{AffectedRows: 1}, nil
}

func (c fakeUserClient) PatchUser(
	_ context.Context,
	req *userv1.PatchUserRequest,
	_ ...grpc.CallOption,
) (*userv1.PatchUserResponse, error) {
	if c.patchReq != nil {
		*c.patchReq = req
	}

	return &userv1.PatchUserResponse{AffectedRows: 1}, nil
}

func (fakeUserClient) DeleteUser(
	context.Context,
	*userv1.DeleteUserRequest,
	...grpc.CallOption,
) (*userv1.DeleteUserResponse, error) {
	return &userv1.DeleteUserResponse{}, nil
}

func fakeUser(
	id string,
	email string,
	username string,
	displayName string,
	firstName string,
	lastName string,
	phoneNumber string,
	line1 string,
	city string,
	countryCode string,
) *userv1.User {
	now := time.Date(2026, 5, 4, 10, 30, 0, 0, time.UTC)

	return &userv1.User{
		Id:          id,
		Email:       email,
		Username:    username,
		DisplayName: &displayName,
		Status:      userv1.UserStatus_USER_STATUS_ACTIVE,
		Profile: &userv1.UserProfile{
			FirstName:   &firstName,
			LastName:    &lastName,
			PhoneNumber: &phoneNumber,
			Address: &userv1.UserAddress{
				Line1:       &line1,
				City:        &city,
				CountryCode: &countryCode,
			},
		},
		CreatedAt: timestamppb.New(now),
		UpdatedAt: timestamppb.New(now),
	}
}
