package sdk

import (
	"context"
)

// ListUserLogs lists user operation logs with optional filtering and pagination.
//
// Returns a list of log entries for user-related operations such as creation,
// updates, deletions, and role assignments.
//
// Example:
//
//	resp, err := client.ListUserLogs(ctx, &sdk.LogLogListRequest{
//		CommonCondition: sdk.CommonCondition{
//			Page:     1,
//			PageSize: 10,
//			Filters: []sdk.CommonFilter{
//				{
//					Name:   "user_id",
//					Values: []string{"123"},
//				},
//			},
//		},
//	})
//	if err != nil {
//		return err
//	}
//	for _, log := range resp.List {
//		fmt.Printf("Log: %s\n", log.Operation)
//	}
func (c *RawClient) ListUserLogs(ctx context.Context, req *LogLogListRequest, opts ...CallOption) (*LogLogListResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp LogLogListResponse
	if err := c.postJSON(ctx, "/log/user", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListRoleLogs lists role operation logs with optional filtering and pagination.
//
// Returns a list of log entries for role-related operations such as creation,
// updates, deletions, and privilege assignments.
//
// Example:
//
//	resp, err := client.ListRoleLogs(ctx, &sdk.LogLogListRequest{
//		CommonCondition: sdk.CommonCondition{
//			Page:     1,
//			PageSize: 10,
//			Filters: []sdk.CommonFilter{
//				{
//					Name:   "role_id",
//					Values: []string{"456"},
//				},
//			},
//		},
//	})
//	if err != nil {
//		return err
//	}
//	for _, log := range resp.List {
//		fmt.Printf("Log: %s\n", log.Operation)
//	}
func (c *RawClient) ListRoleLogs(ctx context.Context, req *LogLogListRequest, opts ...CallOption) (*LogLogListResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp LogLogListResponse
	if err := c.postJSON(ctx, "/log/role", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}
