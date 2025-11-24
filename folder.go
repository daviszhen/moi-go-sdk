package sdk

import (
	"context"
)

// CreateFolder creates a new folder in the specified volume.
//
// Folders are used to organize files within a volume. A folder can be created
// in the root of the volume or within another folder.
//
// Example:
//
//	resp, err := client.CreateFolder(ctx, &sdk.FolderCreateRequest{
//		Name:     "my-folder",
//		VolumeID: "volume-id-123",
//		ParentID: "", // empty for root, or specify parent folder ID
//	})
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Created folder ID: %s\n", resp.FolderID)
func (c *RawClient) CreateFolder(ctx context.Context, req *FolderCreateRequest, opts ...CallOption) (*FolderCreateResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp FolderCreateResponse
	if err := c.postJSON(ctx, "/catalog/folder/create", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateFolder updates folder information.
//
// You can update the folder name.
//
// Example:
//
//	resp, err := client.UpdateFolder(ctx, &sdk.FolderUpdateRequest{
//		FolderID: "folder-id-123",
//		Name:     "updated-folder-name",
//	})
func (c *RawClient) UpdateFolder(ctx context.Context, req *FolderUpdateRequest, opts ...CallOption) (*FolderUpdateResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp FolderUpdateResponse
	if err := c.postJSON(ctx, "/catalog/folder/update", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteFolder deletes the specified folder.
//
// This operation will also delete all files and subfolders within the folder.
//
// Example:
//
//	resp, err := client.DeleteFolder(ctx, &sdk.FolderDeleteRequest{
//		FolderID: "folder-id-123",
//	})
func (c *RawClient) DeleteFolder(ctx context.Context, req *FolderDeleteRequest, opts ...CallOption) (*FolderDeleteResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp FolderDeleteResponse
	if err := c.postJSON(ctx, "/catalog/folder/delete", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CleanFolder removes all files and subfolders from the folder without deleting the folder itself.
//
// The folder structure remains, but all contents are removed.
//
// Example:
//
//	resp, err := client.CleanFolder(ctx, &sdk.FolderCleanRequest{
//		FolderID: "folder-id-123",
//	})
func (c *RawClient) CleanFolder(ctx context.Context, req *FolderCleanRequest, opts ...CallOption) (*FolderCleanResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp FolderCleanResponse
	if err := c.postJSON(ctx, "/catalog/folder/clean", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetFolderRefList retrieves the list of references to the specified folder.
//
// Returns a list of objects that reference this folder, such as workflows.
//
// Example:
//
//	resp, err := client.GetFolderRefList(ctx, &sdk.FolderRefListRequest{
//		FolderID: "folder-id-123",
//	})
func (c *RawClient) GetFolderRefList(ctx context.Context, req *FolderRefListRequest, opts ...CallOption) (*FolderRefListResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp FolderRefListResponse
	if err := c.postJSON(ctx, "/catalog/folder/ref_list", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}
