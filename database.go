package sdk

import (
	"context"
)

// CreateDatabase creates a new database under the specified catalog.
//
// A database is an organizational unit within a catalog that contains tables and volumes.
//
// Example:
//
//	resp, err := client.CreateDatabase(ctx, &sdk.DatabaseCreateRequest{
//		DatabaseName: "my-database",
//		Comment:      "My database description",
//		CatalogID:   123,
//	})
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Created database ID: %d\n", resp.DatabaseID)
func (c *RawClient) CreateDatabase(ctx context.Context, req *DatabaseCreateRequest, opts ...CallOption) (*DatabaseCreateResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp DatabaseCreateResponse
	if err := c.postJSON(ctx, "/catalog/database/create", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteDatabase deletes the specified database.
//
// This operation will also delete all tables and volumes within the database.
//
// Example:
//
//	resp, err := client.DeleteDatabase(ctx, &sdk.DatabaseDeleteRequest{
//		DatabaseID: 456,
//	})
func (c *RawClient) DeleteDatabase(ctx context.Context, req *DatabaseDeleteRequest, opts ...CallOption) (*DatabaseDeleteResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp DatabaseDeleteResponse
	if err := c.postJSON(ctx, "/catalog/database/delete", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateDatabase updates database information.
//
// You can update the database comment. The database name cannot be changed.
//
// Example:
//
//	resp, err := client.UpdateDatabase(ctx, &sdk.DatabaseUpdateRequest{
//		DatabaseID: 456,
//		Comment:    "Updated description",
//	})
func (c *RawClient) UpdateDatabase(ctx context.Context, req *DatabaseUpdateRequest, opts ...CallOption) (*DatabaseUpdateResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp DatabaseUpdateResponse
	if err := c.postJSON(ctx, "/catalog/database/update", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetDatabase retrieves detailed information about the specified database.
//
// The response includes the database name, description, and timestamps.
//
// Example:
//
//	resp, err := client.GetDatabase(ctx, &sdk.DatabaseInfoRequest{
//		DatabaseID: 456,
//	})
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Database: %s\n", resp.DatabaseName)
func (c *RawClient) GetDatabase(ctx context.Context, req *DatabaseInfoRequest, opts ...CallOption) (*DatabaseInfoResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp DatabaseInfoResponse
	if err := c.postJSON(ctx, "/catalog/database/info", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListDatabases lists all databases under the specified catalog.
//
// Returns a list of all databases in the catalog.
//
// Example:
//
//	resp, err := client.ListDatabases(ctx, &sdk.DatabaseListRequest{
//		CatalogID: 123,
//	})
//	if err != nil {
//		return err
//	}
//	for _, db := range resp.List {
//		fmt.Printf("Database: %s\n", db.DatabaseName)
//	}
func (c *RawClient) ListDatabases(ctx context.Context, req *DatabaseListRequest, opts ...CallOption) (*DatabaseListResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp DatabaseListResponse
	if err := c.postJSON(ctx, "/catalog/database/list", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetDatabaseChildren retrieves the children of the specified database (tables and volumes).
//
// Returns both tables and volumes that belong to the database.
//
// Example:
//
//	resp, err := client.GetDatabaseChildren(ctx, &sdk.DatabaseChildrenRequest{
//		DatabaseID: 456,
//	})
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Tables: %d, Volumes: %d\n", len(resp.Tables), len(resp.Volumes))
func (c *RawClient) GetDatabaseChildren(ctx context.Context, req *DatabaseChildrenRequest, opts ...CallOption) (*DatabaseChildrenResponseData, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp DatabaseChildrenResponseData
	if err := c.postJSON(ctx, "/catalog/database/children", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetDatabaseRefList retrieves the list of references to the specified database.
//
// Returns a list of references associated with the database.
//
// Example:
//
//	resp, err := client.GetDatabaseRefList(ctx, &sdk.DatabaseRefListRequest{
//		DatabaseID: 456,
//	})
func (c *RawClient) GetDatabaseRefList(ctx context.Context, req *DatabaseRefListRequest, opts ...CallOption) (*DatabaseRefListResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp DatabaseRefListResponse
	if err := c.postJSON(ctx, "/catalog/database/ref_list", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}
