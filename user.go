package sdk

import (
	"context"
)

// CreateUser creates a new user account.
//
// The user can be assigned roles and privileges after creation.
//
// Example:
//
//	resp, err := client.CreateUser(ctx, &sdk.UserCreateRequest{
//		UserName: "john.doe",
//		Password: "secure-password",
//		Comment:  "User description",
//	})
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Created user ID: %d\n", resp.UserID)
func (c *RawClient) CreateUser(ctx context.Context, req *UserCreateRequest, opts ...CallOption) (*UserCreateResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp UserCreateResponse
	if err := c.postJSON(ctx, "/user/create", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteUser deletes the specified user account.
//
// This operation permanently removes the user and all associated data.
//
// Example:
//
//	resp, err := client.DeleteUser(ctx, &sdk.UserDeleteUserRequest{
//		UserID: 123,
//	})
func (c *RawClient) DeleteUser(ctx context.Context, req *UserDeleteUserRequest, opts ...CallOption) (*UserDeleteUserResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp UserDeleteUserResponse
	if err := c.postJSON(ctx, "/user/delete", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetUserDetail retrieves detailed information about the specified user.
//
// The response includes user name, status, roles, and other metadata.
//
// Example:
//
//	resp, err := client.GetUserDetail(ctx, &sdk.UserDetailInfoRequest{
//		UserID: 123,
//	})
//	if err != nil {
//		return err
//	}
//	fmt.Printf("User: %s\n", resp.UserName)
func (c *RawClient) GetUserDetail(ctx context.Context, req *UserDetailInfoRequest, opts ...CallOption) (*UserDetailInfoResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp UserDetailInfoResponse
	if err := c.postJSON(ctx, "/user/detail_info", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListUsers lists users with optional filtering and pagination.
//
// Supports filtering by name, status, and other criteria.
//
// Example:
//
//	resp, err := client.ListUsers(ctx, &sdk.UserListRequest{
//		Keyword: "john",
//		CommonCondition: sdk.CommonCondition{
//			Page:     1,
//			PageSize: 10,
//		},
//	})
//	if err != nil {
//		return err
//	}
//	for _, user := range resp.List {
//		fmt.Printf("User: %s\n", user.UserName)
//	}
func (c *RawClient) ListUsers(ctx context.Context, req *UserListRequest, opts ...CallOption) (*UserListResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp UserListResponse
	if err := c.postJSON(ctx, "/user/list", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateUserPassword updates the password for the specified user.
//
// This operation requires appropriate permissions to change another user's password.
//
// Example:
//
//	resp, err := client.UpdateUserPassword(ctx, &sdk.UserUpdatePasswordRequest{
//		UserID:  123,
//		NewPassword: "new-secure-password",
//	})
func (c *RawClient) UpdateUserPassword(ctx context.Context, req *UserUpdatePasswordRequest, opts ...CallOption) (*UserUpdatePasswordResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp UserUpdatePasswordResponse
	if err := c.postJSON(ctx, "/user/update_password", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateUserInfo updates user information such as name, email, phone, etc.
//
// You can update various user profile fields.
//
// Example:
//
//	resp, err := client.UpdateUserInfo(ctx, &sdk.UserUpdateInfoRequest{
//		UserID: 123,
//		Email:  "newemail@example.com",
//		Phone:  "1234567890",
//	})
func (c *RawClient) UpdateUserInfo(ctx context.Context, req *UserUpdateInfoRequest, opts ...CallOption) (*UserUpdateInfoResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp UserUpdateInfoResponse
	if err := c.postJSON(ctx, "/user/update_info", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateUserRoles updates the roles assigned to a user.
//
// This replaces the existing role assignments with the new list.
//
// Example:
//
//	resp, err := client.UpdateUserRoles(ctx, &sdk.UserUpdateRoleListRequest{
//		UserID:  123,
//		RoleIDs: []RoleID{456, 789},
//	})
func (c *RawClient) UpdateUserRoles(ctx context.Context, req *UserUpdateRoleListRequest, opts ...CallOption) (*UserUpdateRoleListResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp UserUpdateRoleListResponse
	if err := c.postJSON(ctx, "/user/update_role_list", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateUserStatus updates the status of the specified user.
//
// User status controls whether the user account is active or inactive.
//
// Example:
//
//	resp, err := client.UpdateUserStatus(ctx, &sdk.UserUpdateStatusRequest{
//		UserID: 123,
//		Status: 1, // 1 for active, 0 for inactive
//	})
func (c *RawClient) UpdateUserStatus(ctx context.Context, req *UserUpdateStatusRequest, opts ...CallOption) (*UserUpdateStatusResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp UserUpdateStatusResponse
	if err := c.postJSON(ctx, "/user/update_status", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetMyAPIKey retrieves the API key for the current authenticated user.
//
// The API key can be used for programmatic access to the API.
//
// Example:
//
//	resp, err := client.GetMyAPIKey(ctx)
//	if err != nil {
//		return err
//	}
//	fmt.Printf("API Key: %s\n", resp.ApiKey)
func (c *RawClient) GetMyAPIKey(ctx context.Context, opts ...CallOption) (*UserApiKeyResponse, error) {
	var resp UserApiKeyResponse
	if err := c.postJSON(ctx, "/user/me/api-key", nil, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// RefreshMyAPIKey generates a new API key for the current authenticated user.
//
// The old API key will be invalidated and a new one will be generated.
//
// Example:
//
//	resp, err := client.RefreshMyAPIKey(ctx)
//	if err != nil {
//		return err
//	}
//	fmt.Printf("New API Key: %s\n", resp.ApiKey)
func (c *RawClient) RefreshMyAPIKey(ctx context.Context, opts ...CallOption) (*UserApiKeyRefreshResonse, error) {
	var resp UserApiKeyRefreshResonse
	if err := c.postJSON(ctx, "/user/me/api-key/refresh", nil, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetMyInfo retrieves information about the current authenticated user.
//
// Returns the user profile and metadata for the user making the request.
//
// Example:
//
//	resp, err := client.GetMyInfo(ctx)
//	if err != nil {
//		return err
//	}
//	fmt.Printf("User: %s\n", resp.UserName)
func (c *RawClient) GetMyInfo(ctx context.Context, opts ...CallOption) (*UserMeInfoResponse, error) {
	var resp UserMeInfoResponse
	if err := c.postJSON(ctx, "/user/me/info", nil, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateMyInfo updates information for the current authenticated user.
//
// You can update your own profile information such as email, phone, etc.
//
// Example:
//
//	resp, err := client.UpdateMyInfo(ctx, &sdk.UserMeUpdateInfoRequest{
//		Email: "newemail@example.com",
//		Phone: "1234567890",
//	})
func (c *RawClient) UpdateMyInfo(ctx context.Context, req *UserMeUpdateInfoRequest, opts ...CallOption) (*UserMeUpdateInfoResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp UserMeUpdateInfoResponse
	if err := c.postJSON(ctx, "/user/me/update_info", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateMyPassword updates the password for the current authenticated user.
//
// This allows users to change their own password.
//
// Example:
//
//	resp, err := client.UpdateMyPassword(ctx, &sdk.UserMeUpdatePasswordRequest{
//		OldPassword: "old-password",
//		NewPassword: "new-secure-password",
//	})
func (c *RawClient) UpdateMyPassword(ctx context.Context, req *UserMeUpdatePasswordRequest, opts ...CallOption) (*UserMeUpdatePasswordResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp UserMeUpdatePasswordResponse
	if err := c.postJSON(ctx, "/user/me/update_password", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}
