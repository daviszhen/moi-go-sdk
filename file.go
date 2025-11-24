package sdk

import (
	"context"
)

// CreateFile creates a new file in the specified volume.
//
// The file can be created in the root of the volume or within a folder.
//
// Example:
//
//	resp, err := client.CreateFile(ctx, &sdk.FileCreateRequest{
//		Name:     "my-file.txt",
//		VolumeID: "volume-id-123",
//		ParentID: "folder-id-456", // optional, empty for root
//		Size:     1024,
//		ShowType: "normal",
//	})
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Created file ID: %s\n", resp.FileID)
func (c *RawClient) CreateFile(ctx context.Context, req *FileCreateRequest, opts ...CallOption) (*FileCreateResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp FileCreateResponse
	if err := c.postJSON(ctx, "/catalog/file/create", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateFile updates file information.
//
// You can update the file name. Other properties may be limited.
//
// Example:
//
//	resp, err := client.UpdateFile(ctx, &sdk.FileUpdateRequest{
//		FileID: "file-id-123",
//		Name:   "updated-name.txt",
//	})
func (c *RawClient) UpdateFile(ctx context.Context, req *FileUpdateRequest, opts ...CallOption) (*FileUpdateResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp FileUpdateResponse
	if err := c.postJSON(ctx, "/catalog/file/update", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteFile deletes the specified file.
//
// This operation permanently deletes the file.
//
// Example:
//
//	resp, err := client.DeleteFile(ctx, &sdk.FileDeleteRequest{
//		FileID: "file-id-123",
//	})
func (c *RawClient) DeleteFile(ctx context.Context, req *FileDeleteRequest, opts ...CallOption) (*FileDeleteResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp FileDeleteResponse
	if err := c.postJSON(ctx, "/catalog/file/delete", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteFileRef deletes a file by its reference file ID.
//
// This is useful when you have a reference ID instead of the actual file ID.
//
// Example:
//
//	resp, err := client.DeleteFileRef(ctx, &sdk.FileDeleteRefRequest{
//		RefFileID: "ref-id-123",
//	})
func (c *RawClient) DeleteFileRef(ctx context.Context, req *FileDeleteRefRequest, opts ...CallOption) (*FileDeleteRefResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp FileDeleteRefResponse
	if err := c.postJSON(ctx, "/catalog/file/delete_ref", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetFile retrieves detailed information about the specified file.
//
// The response includes file name, size, type, and metadata.
//
// Example:
//
//	resp, err := client.GetFile(ctx, &sdk.FileInfoRequest{
//		FileID: "file-id-123",
//	})
//	if err != nil {
//		return err
//	}
//	fmt.Printf("File: %s, Size: %d\n", resp.Name, resp.Size)
func (c *RawClient) GetFile(ctx context.Context, req *FileInfoRequest, opts ...CallOption) (*FileInfoResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp FileInfoResponse
	if err := c.postJSON(ctx, "/catalog/file/info", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListFiles lists files in a volume or folder with optional filtering.
//
// Supports filtering by volume ID, parent ID, file type, and other criteria.
//
// Example:
//
//	resp, err := client.ListFiles(ctx, &sdk.FileListRequest{
//		CommonCondition: sdk.CommonCondition{
//			Page:     1,
//			PageSize: 10,
//			Filters: []sdk.CommonFilter{
//				{
//					Name:   "volume_id",
//					Values: []string{"volume-id-123"},
//				},
//			},
//		},
//	})
//	if err != nil {
//		return err
//	}
//	for _, file := range resp.List {
//		fmt.Printf("File: %s\n", file.Name)
//	}
func (c *RawClient) ListFiles(ctx context.Context, req *FileListRequest, opts ...CallOption) (*FileListResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp FileListResponse
	if err := c.postJSON(ctx, "/catalog/file/list", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UploadFile uploads a file to the catalog service.
//
// This is a simple file upload endpoint. For advanced features like table import,
// use UploadConnectorFile instead.
//
// Example:
//
//	resp, err := client.UploadFile(ctx, &sdk.FileUploadRequest{
//		FileID:   "file-id-123",
//		SavePath: "/tmp",
//	})
func (c *RawClient) UploadFile(ctx context.Context, req *FileUploadRequest, opts ...CallOption) (*FileUploadResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp FileUploadResponse
	if err := c.postJSON(ctx, "/catalog/file/upload", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetFileDownloadLink retrieves a signed download link for the file.
//
// The link is a temporary URL that can be used to download the file.
//
// Example:
//
//	resp, err := client.GetFileDownloadLink(ctx, &sdk.FileDownloadRequest{
//		FileID: "file-id-123",
//	})
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Download URL: %s\n", resp.Url)
func (c *RawClient) GetFileDownloadLink(ctx context.Context, req *FileDownloadRequest, opts ...CallOption) (*FileDownloadResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp FileDownloadResponse
	if err := c.postJSON(ctx, "/catalog/file/download", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetFilePreviewLink retrieves a signed preview link for the file.
//
// The link can be used to preview the file in a browser or application.
//
// Example:
//
//	resp, err := client.GetFilePreviewLink(ctx, &sdk.FilePreviewLinkRequest{
//		FileID: "file-id-123",
//	})
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Preview URL: %s\n", resp.Url)
func (c *RawClient) GetFilePreviewLink(ctx context.Context, req *FilePreviewLinkRequest, opts ...CallOption) (*FilePreviewLinkResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp FilePreviewLinkResponse
	if err := c.postJSON(ctx, "/catalog/file/preview_link", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetFilePreviewStream retrieves a preview stream URL for the file.
//
// The stream URL can be used to stream the file content for preview purposes.
//
// Example:
//
//	resp, err := client.GetFilePreviewStream(ctx, &sdk.FilePreviewStreamRequest{
//		FileID: "file-id-123",
//	})
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Stream URL: %s\n", resp.Url)
func (c *RawClient) GetFilePreviewStream(ctx context.Context, req *FilePreviewStreamRequest, opts ...CallOption) (*FilePreviewLinkResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp FilePreviewLinkResponse
	if err := c.postJSON(ctx, "/catalog/file/preview_stream", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}
