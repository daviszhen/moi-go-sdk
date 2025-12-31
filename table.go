package sdk

import (
	"context"
)

// CreateTable creates a new table in the specified database.
//
// The table is created with the specified schema and properties.
//
// Example:
//
//	resp, err := client.CreateTable(ctx, &sdk.TableCreateRequest{
//		DatabaseID: 123,
//		TableName:  "my_table",
//		// ... other fields
//	})
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Created table ID: %d\n", resp.TableID)
func (c *RawClient) CreateTable(ctx context.Context, req *TableCreateRequest, opts ...CallOption) (*TableCreateResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp TableCreateResponse
	if err := c.postJSON(ctx, "/catalog/table/create", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetTable retrieves detailed information about the specified table.
//
// The response includes table schema, properties, and metadata.
//
// Example:
//
//	resp, err := client.GetTable(ctx, &sdk.TableInfoRequest{
//		TableID: 456,
//	})
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Table: %s\n", resp.TableName)
func (c *RawClient) GetTable(ctx context.Context, req *TableInfoRequest, opts ...CallOption) (*TableInfoResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp TableInfoResponse
	if err := c.postJSON(ctx, "/catalog/table/info", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetTable retrieves detailed information about the specified table.
//
// The response includes table schema, properties, and metadata.
//
// Example:
//
//	resp, err := client.GetMultiTable(ctx, &sdk.MultiTableInfoRequest{
//		TableList: []TableInfoRequest{
//			{
//				TableID: 456,  //如果是普通表，传table_id
//			},
//			{
//				DatabaseID: 123, //如果是订阅表，传database_id 和 table_name
//				TableName:  "sub_table",
//				TableID:    TableIDInSubDatabase, //订阅表的table_id是一个特殊值
//			},
//		},
//	})
//	if err != nil {
//		return err
//	}
//	fmt.Printf("SubTable: %s\n", resp.InfoMap["123 sub_table"].TableName)
func (c *RawClient) GetMultiTable(ctx context.Context, req *MultiTableInfoRequest, opts ...CallOption) (*MultiTableInfoResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp MultiTableInfoResponse
	if err := c.postJSON(ctx, "/catalog/table/multi_info", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetTableOverview retrieves an overview of all tables.
//
// Returns a summary list of tables with basic information.
//
// Example:
//
//	tables, err := client.GetTableOverview(ctx)
//	if err != nil {
//		return err
//	}
//	for _, table := range tables {
//		fmt.Printf("Table: %s\n", table.TableName)
//	}
func (c *RawClient) GetTableOverview(ctx context.Context, opts ...CallOption) ([]TableOverview, error) {
	var resp []TableOverview
	if err := c.postJSON(ctx, "/catalog/table/overview", struct{}{}, &resp, opts...); err != nil {
		return nil, err
	}
	return resp, nil
}

// CheckTableExists checks if a table exists by database ID and table name.
//
// Returns true if the table exists, false otherwise.
//
// Example:
//
//	exists, err := client.CheckTableExists(ctx, &sdk.TableExistRequest{
//		DatabaseID: 123,
//		TableName:  "my_table",
//	})
//	if err != nil {
//		return err
//	}
//	if exists {
//		fmt.Println("Table exists")
//	}
func (c *RawClient) CheckTableExists(ctx context.Context, req *TableExistRequest, opts ...CallOption) (bool, error) {
	if req == nil {
		return false, ErrNilRequest
	}
	var exists bool
	if err := c.postJSON(ctx, "/catalog/table/exist", req, &exists, opts...); err != nil {
		return false, err
	}
	return exists, nil
}

// PreviewTable previews table data without loading it into memory.
//
// Returns a preview of the table data with limited rows.
//
// Example:
//
//	resp, err := client.PreviewTable(ctx, &sdk.TablePreviewRequest{
//		TableID: 456,
//		Limit:   10,
//	})
func (c *RawClient) PreviewTable(ctx context.Context, req *TablePreviewRequest, opts ...CallOption) (*TablePreviewResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp TablePreviewResponse
	if err := c.postJSON(ctx, "/catalog/table/preview", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// LoadTable loads table data into memory for processing.
//
// This operation may take time for large tables.
//
// Example:
//
//	resp, err := client.LoadTable(ctx, &sdk.TableLoadRequest{
//		TableID: 456,
//	})
func (c *RawClient) LoadTable(ctx context.Context, req *TableLoadRequest, opts ...CallOption) (*TableLoadResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp TableLoadResponse
	if err := c.postJSON(ctx, "/catalog/table/load", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetTableDownloadLink retrieves a download link for the table data.
//
// The link is a signed URL that can be used to download the table data.
//
// Example:
//
//	resp, err := client.GetTableDownloadLink(ctx, &sdk.TableDownloadRequest{
//		TableID: 456,
//	})
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Download URL: %s\n", resp.Url)
func (c *RawClient) GetTableDownloadLink(ctx context.Context, req *TableDownloadRequest, opts ...CallOption) (*TableDownloadResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp TableDownloadResponse
	if err := c.postJSON(ctx, "/catalog/table/download", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// TruncateTable removes all data from the table while keeping the table structure.
//
// This operation is irreversible. All data in the table will be deleted.
//
// Example:
//
//	_, err := client.TruncateTable(ctx, &sdk.TableTruncateRequest{
//		TableID: 456,
//	})
func (c *RawClient) TruncateTable(ctx context.Context, req *TableTruncateRequest, opts ...CallOption) (*TableTruncateResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp TableTruncateResponse
	if err := c.postJSON(ctx, "/catalog/table/truncate", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteTable deletes the specified table.
//
// This operation will permanently delete the table and all its data.
//
// Example:
//
//	_, err := client.DeleteTable(ctx, &sdk.TableDeleteRequest{
//		TableID: 456,
//	})
func (c *RawClient) DeleteTable(ctx context.Context, req *TableDeleteRequest, opts ...CallOption) (*TableDeleteResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp TableDeleteResponse
	if err := c.postJSON(ctx, "/catalog/table/delete", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetTableFullPath retrieves the full path of the table in the catalog hierarchy.
//
// The path includes catalog, database, and table names.
//
// Example:
//
//	resp, err := client.GetTableFullPath(ctx, &sdk.TableFullPathRequest{
//		TableID: 456,
//	})
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Full path: %v\n", resp.FullPath.NameList)
func (c *RawClient) GetTableFullPath(ctx context.Context, req *TableFullPathRequest, opts ...CallOption) (*TableFullPathResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp TableFullPathResponse
	if err := c.postJSON(ctx, "/catalog/table/full_path", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetTableRefList retrieves the list of references to the specified table.
//
// Returns a list of objects that reference this table.
//
// Example:
//
//	resp, err := client.GetTableRefList(ctx, &sdk.TableRefListRequest{
//		TableID: 456,
//	})
func (c *RawClient) GetTableRefList(ctx context.Context, req *TableRefListRequest, opts ...CallOption) (*TableRefListResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp TableRefListResponse
	if err := c.postJSON(ctx, "/catalog/table/ref_list", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}
