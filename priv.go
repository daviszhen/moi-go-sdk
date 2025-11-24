package sdk

import (
	"context"
)

// ListObjectsByCategory lists objects by category for privilege management.
//
// This is useful for finding objects (tables, volumes, etc.) that can be assigned
// privileges in role management.
//
// Example:
//
//	resp, err := client.ListObjectsByCategory(ctx, &sdk.PrivListObjByCategoryRequest{
//		Category: "table",
//		DatabaseID: 123,
//	})
//	if err != nil {
//		return err
//	}
//	for _, obj := range resp.List {
//		fmt.Printf("Object: %s (ID: %s)\n", obj.Name, obj.ID)
//	}
func (c *RawClient) ListObjectsByCategory(ctx context.Context, req *PrivListObjByCategoryRequest, opts ...CallOption) (*PrivListObjByCategoryResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp PrivListObjByCategoryResponse
	if err := c.postJSON(ctx, "/rbac/priv/list_obj_by_category", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}
