package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonv1 "oas-sandbox/gen/grpc/common/v1"
	userv1 "oas-sandbox/gen/grpc/user/v1"
)

func TestHTTPServerHealth(t *testing.T) {
	srv := NewHTTPServer(0, "oas-sandbox", fakeUserClient{})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	srv.server.Handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ok"`)
}

func TestHTTPServerOpenAPI(t *testing.T) {
	srv := NewHTTPServer(0, "oas-sandbox", fakeUserClient{})
	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	rec := httptest.NewRecorder()

	srv.server.Handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"title":"OAS Sandbox"`)
	assert.Contains(t, rec.Body.String(), `"/v1/users/{id}"`)
	assert.NotContains(t, rec.Body.String(), `"/v1/users":`)
}

func TestDocsServesSwaggerUIOfflineAndAllowsDownloads(t *testing.T) {
	srv := NewHTTPServer(0, "oas-sandbox", fakeUserClient{})
	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
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
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
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
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
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
	req := httptest.NewRequest(http.MethodGet, "/v1/users/0198f8f0-0000-7000-8000-000000000001", nil)
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

type fakeUserClient struct{}

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
	now := time.Date(2026, 5, 4, 10, 30, 0, 0, time.UTC)

	return &userv1.GetUserResponse{
		User: &userv1.User{
			Id:          req.GetId(),
			Email:       "kitti@example.com",
			Username:    "kitti",
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
		},
	}, nil
}

func (fakeUserClient) ListUsers(
	context.Context,
	*userv1.ListUsersRequest,
	...grpc.CallOption,
) (*userv1.ListUsersResponse, error) {
	return &userv1.ListUsersResponse{Pagination: &commonv1.PaginationResponse{}}, nil
}

func (fakeUserClient) CreateUser(
	context.Context,
	*userv1.CreateUserRequest,
	...grpc.CallOption,
) (*userv1.CreateUserResponse, error) {
	return &userv1.CreateUserResponse{}, nil
}

func (fakeUserClient) UpdateUser(
	context.Context,
	*userv1.UpdateUserRequest,
	...grpc.CallOption,
) (*userv1.UpdateUserResponse, error) {
	return &userv1.UpdateUserResponse{}, nil
}

func (fakeUserClient) PatchUser(
	context.Context,
	*userv1.PatchUserRequest,
	...grpc.CallOption,
) (*userv1.PatchUserResponse, error) {
	return &userv1.PatchUserResponse{}, nil
}

func (fakeUserClient) DeleteUser(
	context.Context,
	*userv1.DeleteUserRequest,
	...grpc.CallOption,
) (*userv1.DeleteUserResponse, error) {
	return &userv1.DeleteUserResponse{}, nil
}
