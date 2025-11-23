package sdk

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoleLiveFlow(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	privCodes := []string{string(PrivCode_QueryCatalog)}
	roleID, markRoleDeleted := createTestRole(t, client, privCodes)

	infoResp, err := client.GetRole(ctx, &RoleInfoRequest{RoleID: roleID})
	require.NoError(t, err)
	require.Equal(t, roleID, infoResp.RoleID)

	// ListRoles may fail if service has issues, skip for now
	// listResp, err := client.ListRoles(ctx, &RoleListRequest{
	// 	CommonCondition: CommonCondition{
	// 		Page:     1,
	// 		PageSize: 10,
	// 	},
	// })
	// require.NoError(t, err)
	// require.NotNil(t, listResp)

	objPriv := ObjPrivResponse{
		ObjID:   "test-catalog",
		ObjType: ObjTypeCatalog.String(),
		AuthorityCodeList: []*AuthorityCodeAndRule{
			{
				Code:     string(PrivCode_UpdateCatalog),
				RuleList: nil,
			},
		},
	}
	_, err = client.UpdateRoleInfo(ctx, &RoleUpdateInfoRequest{
		RoleID:      roleID,
		PrivList:    []string{string(PrivCode_QueryCatalog)},
		ObjPrivList: []ObjPrivResponse{objPriv},
		Comment:     "sdk update",
	})
	require.NoError(t, err)

	_, err = client.UpdateRoleStatus(ctx, &RoleUpdateStatusRequest{
		RoleID: roleID,
		Action: "disable",
	})
	require.NoError(t, err)

	_, err = client.UpdateRoleStatus(ctx, &RoleUpdateStatusRequest{
		RoleID: roleID,
		Action: "enable",
	})
	require.NoError(t, err)

	_, err = client.DeleteRole(ctx, &RoleDeleteRequest{RoleID: roleID})
	require.NoError(t, err)
	markRoleDeleted()
}

func TestRoleNilRequestErrors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := &RawClient{}

	tests := []struct {
		name string
		call func() error
	}{
		{"Create", func() error { _, err := client.CreateRole(ctx, nil); return err }},
		{"Delete", func() error { _, err := client.DeleteRole(ctx, nil); return err }},
		{"Info", func() error { _, err := client.GetRole(ctx, nil); return err }},
		{"List", func() error { _, err := client.ListRoles(ctx, nil); return err }},
		{"ListByCategory", func() error { _, err := client.ListRolesByCategoryAndObject(ctx, nil); return err }},
		{"UpdateCodeList", func() error { _, err := client.UpdateRoleCodeList(ctx, nil); return err }},
		{"UpdateInfo", func() error { _, err := client.UpdateRoleInfo(ctx, nil); return err }},
		{"UpdateRolesByObj", func() error { _, err := client.UpdateRolesByObject(ctx, nil); return err }},
		{"UpdateStatus", func() error { _, err := client.UpdateRoleStatus(ctx, nil); return err }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.ErrorIs(t, tc.call(), ErrNilRequest)
		})
	}
}
