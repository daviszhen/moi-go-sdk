package sdk

import (
	"context"
)

// CreateRole creates a new role with specified privileges.
//
// Roles are used to manage permissions. You can assign global privileges
// and object-level privileges (e.g., table privileges) to a role.
//
// Example:
//
//	resp, err := client.CreateRole(ctx, &sdk.RoleCreateRequest{
//		RoleName: "my-role",
//		Comment:  "Role description",
//		PrivList: []string{"U1", "R1"}, // global privileges
//		ObjPrivList: []sdk.ObjPrivResponse{
//			{
//				ObjID:   "123",
//				ObjType: "table",
//				AuthorityCodeList: []*sdk.AuthorityCodeAndRule{
//					{Code: "DT8", RuleList: nil}, // SELECT permission
//				},
//			},
//		},
//	})
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Created role ID: %d\n", resp.RoleID)
func (c *RawClient) CreateRole(ctx context.Context, req *RoleCreateRequest, opts ...CallOption) (*RoleCreateResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp RoleCreateResponse
	if err := c.postJSON(ctx, "/role/create", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteRole deletes the specified role.
//
// This operation will remove the role and all its privilege assignments.
//
// Example:
//
//	resp, err := client.DeleteRole(ctx, &sdk.RoleDeleteRequest{
//		RoleID: 456,
//	})
func (c *RawClient) DeleteRole(ctx context.Context, req *RoleDeleteRequest, opts ...CallOption) (*RoleDeleteResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp RoleDeleteResponse
	if err := c.postJSON(ctx, "/role/delete", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetRole retrieves detailed information about the specified role.
//
// The response includes role name, description, global privileges, and object-level privileges.
//
// Example:
//
//	resp, err := client.GetRole(ctx, &sdk.RoleInfoRequest{
//		RoleID: 456,
//	})
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Role: %s\n", resp.RoleName)
func (c *RawClient) GetRole(ctx context.Context, req *RoleInfoRequest, opts ...CallOption) (*RoleInfoResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp RoleInfoResponse
	if err := c.postJSON(ctx, "/role/info", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListRoles lists roles with optional filtering and pagination.
//
// Supports filtering by name, description, and other criteria.
//
// Example:
//
//	resp, err := client.ListRoles(ctx, &sdk.RoleListRequest{
//		Keyword: "admin",
//		CommonCondition: sdk.CommonCondition{
//			Page:     1,
//			PageSize: 10,
//		},
//	})
//	if err != nil {
//		return err
//	}
//	for _, role := range resp.List {
//		fmt.Printf("Role: %s\n", role.RoleName)
//	}
func (c *RawClient) ListRoles(ctx context.Context, req *RoleListRequest, opts ...CallOption) (*RoleListResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp RoleListResponse
	if err := c.postJSON(ctx, "/role/list", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListRolesByCategoryAndObject lists roles filtered by category and object.
//
// This is useful for finding roles that have privileges on specific objects.
//
// Example:
//
//	resp, err := client.ListRolesByCategoryAndObject(ctx, &sdk.RoleListByCategoryAndObjectRequest{
//		Category: "table",
//		ObjectID: "123",
//	})
func (c *RawClient) ListRolesByCategoryAndObject(ctx context.Context, req *RoleListByCategoryAndObjectRequest, opts ...CallOption) (*RoleListByCategoryAndObjectResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp RoleListByCategoryAndObjectResponse
	if err := c.postJSON(ctx, "/role/list_by_category_and_obj", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateRoleCodeList updates the global privilege codes for a role.
//
// This replaces the existing global privileges with the new list.
//
// Example:
//
//	resp, err := client.UpdateRoleCodeList(ctx, &sdk.RoleUpdateCodeListRequest{
//		RoleID:  456,
//		PrivList: []string{"U1", "R1", "C1"},
//	})
func (c *RawClient) UpdateRoleCodeList(ctx context.Context, req *RoleUpdateCodeListRequest, opts ...CallOption) (*RoleUpdateCodeListResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp RoleUpdateCodeListResponse
	if err := c.postJSON(ctx, "/role/update_code_list", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateRoleInfo updates role information including privileges.
//
// The request can include:
//   - PrivList: global privilege codes (string array)
//   - ObjPrivList: object privileges with optional rules (ObjPrivResponse array)
//     Each ObjPrivResponse contains AuthorityCodeList which is []*AuthorityCodeAndRule,
//     where each AuthorityCodeAndRule can have RuleList for row/column level permissions.
//   - Comment: role description
//
// Example:
//
//	resp, err := client.UpdateRoleInfo(ctx, &sdk.RoleUpdateInfoRequest{
//		RoleID:  456,
//		Comment: "Updated description",
//		PrivList: []string{"U1", "R1"},
//		ObjPrivList: []sdk.ObjPrivResponse{
//			{
//				ObjID:   "123",
//				ObjType: "table",
//				AuthorityCodeList: []*sdk.AuthorityCodeAndRule{
//					{
//						Code: "DT8",
//						RuleList: []*sdk.TableRowColRule{
//							{
//								Column:   "department",
//								Relation: "and",
//								ExpressionList: []*sdk.TableRowColExpression{
//									{Operator: "=", Expression: "IT"},
//								},
//							},
//						},
//					},
//				},
//			},
//		},
//	})
func (c *RawClient) UpdateRoleInfo(ctx context.Context, req *RoleUpdateInfoRequest, opts ...CallOption) (*RoleUpdateInfoResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp RoleUpdateInfoResponse
	if err := c.postJSON(ctx, "/role/update_info", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateRolesByObject updates roles associated with a specific object.
//
// This is useful for bulk updating role assignments for an object.
//
// Example:
//
//	resp, err := client.UpdateRolesByObject(ctx, &sdk.RoleUpdateRolesByObjectRequest{
//		ObjectID: "123",
//		Category: "table",
//		RoleIDs:  []RoleID{456, 789},
//	})
func (c *RawClient) UpdateRolesByObject(ctx context.Context, req *RoleUpdateRolesByObjectRequest, opts ...CallOption) (*RoleUpdateRolesByObjectResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp RoleUpdateRolesByObjectResponse
	if err := c.postJSON(ctx, "/role/update_roles_by_obj", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateRoleStatus updates the status of the specified role.
//
// Role status controls whether the role is active or inactive.
//
// Example:
//
//	resp, err := client.UpdateRoleStatus(ctx, &sdk.RoleUpdateStatusRequest{
//		RoleID: 456,
//		Status: 1, // 1 for active, 0 for inactive
//	})
func (c *RawClient) UpdateRoleStatus(ctx context.Context, req *RoleUpdateStatusRequest, opts ...CallOption) (*RoleUpdateStatusResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp RoleUpdateStatusResponse
	if err := c.postJSON(ctx, "/role/update_status", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}
