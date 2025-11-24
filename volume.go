package sdk

import (
	"context"
)

// CreateVolume creates a new volume in the specified database.
//
// A volume is a storage unit that can contain files and folders.
//
// Example:
//
//	resp, err := client.CreateVolume(ctx, &sdk.VolumeCreateRequest{
//		DatabaseID: 123,
//		Name:       "my-volume",
//		Comment:    "My volume description",
//	})
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Created volume ID: %s\n", resp.VolumeID)
func (c *RawClient) CreateVolume(ctx context.Context, req *VolumeCreateRequest, opts ...CallOption) (*VolumeCreateResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp VolumeCreateResponse
	if err := c.postJSON(ctx, "/catalog/volume/create", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteVolume deletes the specified volume.
//
// This operation will also delete all files and folders within the volume.
//
// Example:
//
//	resp, err := client.DeleteVolume(ctx, &sdk.VolumeDeleteRequest{
//		VolumeID: "volume-id-123",
//	})
func (c *RawClient) DeleteVolume(ctx context.Context, req *VolumeDeleteRequest, opts ...CallOption) (*VolumeDeleteResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp VolumeDeleteResponse
	if err := c.postJSON(ctx, "/catalog/volume/delete", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateVolume updates volume information.
//
// You can update the volume name and/or comment.
//
// Example:
//
//	resp, err := client.UpdateVolume(ctx, &sdk.VolumeUpdateRequest{
//		VolumeID: "volume-id-123",
//		Name:     "updated-name",
//		Comment:  "Updated description",
//	})
func (c *RawClient) UpdateVolume(ctx context.Context, req *VolumeUpdateRequest, opts ...CallOption) (*VolumeUpdateResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp VolumeUpdateResponse
	if err := c.postJSON(ctx, "/catalog/volume/update", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetVolume retrieves detailed information about the specified volume.
//
// The response includes volume name, description, and metadata.
//
// Example:
//
//	resp, err := client.GetVolume(ctx, &sdk.VolumeInfoRequest{
//		VolumeID: "volume-id-123",
//	})
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Volume: %s\n", resp.Name)
func (c *RawClient) GetVolume(ctx context.Context, req *VolumeInfoRequest, opts ...CallOption) (*VolumeInfoResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp VolumeInfoResponse
	if err := c.postJSON(ctx, "/catalog/volume/info", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetVolumeRefList retrieves the list of references to the specified volume.
//
// Returns a list of objects that reference this volume.
//
// Example:
//
//	resp, err := client.GetVolumeRefList(ctx, &sdk.VolumeRefListRequest{
//		VolumeID: "volume-id-123",
//	})
func (c *RawClient) GetVolumeRefList(ctx context.Context, req *VolumeRefListRequest, opts ...CallOption) (*VolumeRefListResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp VolumeRefListResponse
	if err := c.postJSON(ctx, "/catalog/volume/ref_list", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetVolumeFullPath retrieves the full path of the volume or folder in the catalog hierarchy.
//
// The path includes catalog, database, volume, and folder names.
//
// Example:
//
//	resp, err := client.GetVolumeFullPath(ctx, &sdk.VolumeFullPathRequest{
//		FolderIDList: []FileID{"folder-id-123"},
//	})
//	if err != nil {
//		return err
//	}
//	for _, path := range resp.FolderFullPath {
//		fmt.Printf("Path: %v\n", path.NameList)
//	}
func (c *RawClient) GetVolumeFullPath(ctx context.Context, req *VolumeFullPathRequest, opts ...CallOption) (*VolumeFullPathResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp VolumeFullPathResponse
	if err := c.postJSON(ctx, "/catalog/volume/full_path", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AddVolumeWorkflowRef adds a workflow reference to the volume.
//
// This associates a workflow with the volume for tracking purposes.
//
// Example:
//
//	resp, err := client.AddVolumeWorkflowRef(ctx, &sdk.VolumeAddRefWorkflowRequest{
//		VolumeID:    "volume-id-123",
//		WorkflowID:  "workflow-id-456",
//	})
func (c *RawClient) AddVolumeWorkflowRef(ctx context.Context, req *VolumeAddRefWorkflowRequest, opts ...CallOption) (*VolumeAddRefWorkflowResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp VolumeAddRefWorkflowResponse
	if err := c.postJSON(ctx, "/catalog/volume/add_ref_workflow", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// RemoveVolumeWorkflowRef removes a workflow reference from the volume.
//
// This disassociates a workflow from the volume.
//
// Example:
//
//	resp, err := client.RemoveVolumeWorkflowRef(ctx, &sdk.VolumeRemoveRefWorkflowRequest{
//		VolumeID:   "volume-id-123",
//		WorkflowID: "workflow-id-456",
//	})
func (c *RawClient) RemoveVolumeWorkflowRef(ctx context.Context, req *VolumeRemoveRefWorkflowRequest, opts ...CallOption) (*VolumeRemoveRefWorkflowResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp VolumeRemoveRefWorkflowResponse
	if err := c.postJSON(ctx, "/catalog/volume/remove_ref_workflow", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}
