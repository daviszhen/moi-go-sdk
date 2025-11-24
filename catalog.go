package sdk

import (
	"context"
)

// CreateCatalog creates a new catalog.
//
// The catalog is the top-level organizational structure for managing databases, tables, and volumes.
//
// Example:
//
//	resp, err := client.CreateCatalog(ctx, &sdk.CatalogCreateRequest{
//		CatalogName: "my-catalog",
//		Comment:     "My catalog description",
//	})
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Created catalog ID: %d\n", resp.CatalogID)
func (c *RawClient) CreateCatalog(ctx context.Context, req *CatalogCreateRequest, opts ...CallOption) (*CatalogCreateResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp CatalogCreateResponse
	if err := c.postJSON(ctx, "/catalog/create", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteCatalog deletes the specified catalog.
//
// This operation will also delete all databases, tables, and volumes within the catalog.
//
// Example:
//
//	resp, err := client.DeleteCatalog(ctx, &sdk.CatalogDeleteRequest{
//		CatalogID: 123,
//	})
func (c *RawClient) DeleteCatalog(ctx context.Context, req *CatalogDeleteRequest, opts ...CallOption) (*CatalogDeleteResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp CatalogDeleteResponse
	if err := c.postJSON(ctx, "/catalog/delete", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateCatalog updates catalog information.
//
// You can update the catalog name and/or comment. Omitted fields will remain unchanged.
//
// Example:
//
//	resp, err := client.UpdateCatalog(ctx, &sdk.CatalogUpdateRequest{
//		CatalogID:   123,
//		CatalogName: "updated-name",
//		Comment:     "Updated description",
//	})
func (c *RawClient) UpdateCatalog(ctx context.Context, req *CatalogUpdateRequest, opts ...CallOption) (*CatalogUpdateResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp CatalogUpdateResponse
	if err := c.postJSON(ctx, "/catalog/update", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCatalog retrieves detailed information about the specified catalog.
//
// The response includes the catalog name, description, and counts of databases, tables, volumes, and files.
//
// Example:
//
//	resp, err := client.GetCatalog(ctx, &sdk.CatalogInfoRequest{
//		CatalogID: 123,
//	})
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Catalog: %s, Databases: %d\n", resp.CatalogName, resp.DatabaseCount)
func (c *RawClient) GetCatalog(ctx context.Context, req *CatalogInfoRequest, opts ...CallOption) (*CatalogInfoResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp CatalogInfoResponse
	if err := c.postJSON(ctx, "/catalog/info", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListCatalogs lists all catalogs.
//
// Returns a list of all catalogs in the system.
//
// Example:
//
//	resp, err := client.ListCatalogs(ctx)
//	if err != nil {
//		return err
//	}
//	for _, catalog := range resp.List {
//		fmt.Printf("Catalog: %s\n", catalog.CatalogName)
//	}
func (c *RawClient) ListCatalogs(ctx context.Context, opts ...CallOption) (*CatalogListResponse, error) {
	var resp CatalogListResponse
	if err := c.postJSON(ctx, "/catalog/list", struct{}{}, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCatalogTree retrieves the hierarchical tree structure of catalogs, databases, tables, and volumes.
//
// The tree structure shows the complete organizational hierarchy of all resources.
//
// Example:
//
//	resp, err := client.GetCatalogTree(ctx)
//	if err != nil {
//		return err
//	}
//	// Traverse the tree structure
//	for _, node := range resp.Tree {
//		fmt.Printf("Type: %s, Name: %s\n", node.Type, node.Name)
//	}
func (c *RawClient) GetCatalogTree(ctx context.Context, opts ...CallOption) (*CatalogTreeResponse, error) {
	var resp CatalogTreeResponse
	if err := c.postJSON(ctx, "/catalog/tree", struct{}{}, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCatalogRefList retrieves the list of references to the specified catalog.
//
// Returns a list of volume references associated with the catalog.
//
// Example:
//
//	resp, err := client.GetCatalogRefList(ctx, &sdk.CatalogRefListRequest{
//		CatalogID: 123,
//	})
func (c *RawClient) GetCatalogRefList(ctx context.Context, req *CatalogRefListRequest, opts ...CallOption) (*CatalogRefListResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp CatalogRefListResponse
	if err := c.postJSON(ctx, "/catalog/ref_list", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}
