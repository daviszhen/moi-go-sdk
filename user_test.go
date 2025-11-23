package sdk

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserLiveFlow(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	roleID, markRoleDeleted := createTestRole(t, client, []string{string(PrivCode_QueryCatalog)})

	createResp, err := client.CreateUser(ctx, &UserCreateRequest{
		UserName:    strings.ToLower(randomUserName()),
		Password:    "TestPwd123!",
		RoleIDList:  []RoleID{roleID},
		Description: "sdk test user",
		Phone:       "12345678901",
		Email:       "sdk@example.com",
	})
	require.NoError(t, err)
	userID := createResp.UserID

	userDeleted := false
	t.Cleanup(func() {
		if userDeleted {
			return
		}
		if _, err := client.DeleteUser(ctx, &UserDeleteUserRequest{UserID: userID}); err != nil {
			t.Logf("cleanup delete user failed: %v", err)
		}
	})

	_, err = client.GetUserDetail(ctx, &UserDetailInfoRequest{UserID: userID})
	require.NoError(t, err)

	_, err = client.ListUsers(ctx, &UserListRequest{})
	require.NoError(t, err)

	_, err = client.UpdateUserInfo(ctx, &UserUpdateInfoRequest{
		UserID:      userID,
		Phone:       "10987654321",
		Email:       "sdk-updated@example.com",
		Description: "updated",
	})
	require.NoError(t, err)

	_, err = client.UpdateUserRoles(ctx, &UserUpdateRoleListRequest{
		UserID:     userID,
		RoleIDList: []RoleID{roleID},
	})
	require.NoError(t, err)

	_, err = client.UpdateUserStatus(ctx, &UserUpdateStatusRequest{
		UserID: userID,
		Action: "disable",
	})
	require.NoError(t, err)

	_, err = client.UpdateUserStatus(ctx, &UserUpdateStatusRequest{
		UserID: userID,
		Action: "enable",
	})
	require.NoError(t, err)

	_, err = client.DeleteUser(ctx, &UserDeleteUserRequest{UserID: userID})
	require.NoError(t, err)
	userDeleted = true

	_, err = client.DeleteRole(ctx, &RoleDeleteRequest{RoleID: roleID})
	require.NoError(t, err)
	markRoleDeleted()
}

func TestUserNilRequestErrors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := &RawClient{}

	tests := []struct {
		name string
		call func() error
	}{
		{"Create", func() error { _, err := client.CreateUser(ctx, nil); return err }},
		{"Delete", func() error { _, err := client.DeleteUser(ctx, nil); return err }},
		{"Detail", func() error { _, err := client.GetUserDetail(ctx, nil); return err }},
		{"List", func() error { _, err := client.ListUsers(ctx, nil); return err }},
		{"UpdatePassword", func() error { _, err := client.UpdateUserPassword(ctx, nil); return err }},
		{"UpdateInfo", func() error { _, err := client.UpdateUserInfo(ctx, nil); return err }},
		{"UpdateRoles", func() error { _, err := client.UpdateUserRoles(ctx, nil); return err }},
		{"UpdateStatus", func() error { _, err := client.UpdateUserStatus(ctx, nil); return err }},
		{"UpdateMyInfo", func() error { _, err := client.UpdateMyInfo(ctx, nil); return err }},
		{"UpdateMyPassword", func() error { _, err := client.UpdateMyPassword(ctx, nil); return err }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.ErrorIs(t, tc.call(), ErrNilRequest)
		})
	}
}
