package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// FileMeta represents file metadata for upload.
type FileMeta struct {
	Filename string `json:"filename"`
	Path     string `json:"path"`
}

// LocalFileUploadRequest represents a request to upload local files.
type LocalFileUploadRequest struct {
	Files []FileUploadItem `json:"-"`
	Meta  []FileMeta       `json:"meta"`
}

// FileUploadItem represents a file to be uploaded.
type FileUploadItem struct {
	File     io.Reader
	FileName string
}

// LocalFileUploadResponse represents a response from local file upload.
type LocalFileUploadResponse struct {
	ConnFileIds []string `json:"conn_file_ids"`
}

// UploadFileRequest represents a request to upload files to connector.
// This is used for the /connectors/upload endpoint.
type UploadFileRequest struct {
	// VolumeID is the volume ID (required)
	VolumeID VolumeID
	// Files are the files to upload (required)
	Files []FileUploadItem
	// Meta is the file metadata array (optional)
	Meta []FileMeta
	// FileTypes is the list of allowed file types (optional)
	FileTypes []int32
	// PathRegex is the path regex filter (optional)
	PathRegex string
	// UnzipKeepStructure indicates whether to keep directory structure when unzipping (optional)
	UnzipKeepStructure bool
	// DedupConfig is the deduplication configuration (optional)
	DedupConfig *DedupConfig
	// TableConfig is the table configuration (optional)
	TableConfig *TableConfig
}

// ConflictPolicy represents the conflict resolution policy when importing data.
type ConflictPolicy int

const (
	// ConflictPolicyFail indicates that import should fail on conflict (0).
	ConflictPolicyFail ConflictPolicy = 0
	// ConflictPolicySkip indicates that conflicting rows should be skipped (1).
	ConflictPolicySkip ConflictPolicy = 1
	// ConflictPolicyReplace indicates that conflicting rows should be replaced (2).
	ConflictPolicyReplace ConflictPolicy = 2
)

// FileAndTableColumnMapping represents the mapping between file columns and table columns.
// Used when importing to an existing table.
type FileAndTableColumnMapping struct {
	// TableColumn is the table column name
	TableColumn string `json:"tableColumn"`
	// Column is the file column name or default value or NULL
	Column string `json:"column"`
	// ColNumInFile is the column number in the file (1-based)
	ColNumInFile int32 `json:"col_num_in_file"`
}

type ExistedTableOpts struct {
	// Method means overwrite the table by new data or append new data to the table
	// options:
	// 	"append"
	// 	"overwrite"
	// 	"" same as "append"
	Method string `json:"method"`
}

// TableConfig represents table configuration for file upload.
type TableConfig struct {
	CreateTable   *CreateTableConfig `json:"create_table"`
	IsColumnName  bool               `json:"isColumnName"`
	ColumnNameRow int                `json:"columnNameRow"`
	RowStart      int                `json:"rowStart"`
	// Conflict specifies the conflict resolution policy:
	// 0: 导入失败 (ConflictPolicyFail)
	// 1: 跳过冲突行 (ConflictPolicySkip)
	// 2: 替换冲突行 (ConflictPolicyReplace)
	Conflict    ConflictPolicy `json:"conflict"`
	ConnFileIDs []string       `json:"conn_file_ids"`
	NewTable    bool           `json:"new_table"`
	DatabaseID  DatabaseID     `json:"database_id"`
	// TableID is the target table ID.
	// Required when using an existing table (new_table = false).
	// Will be filled after table creation when creating a new table (new_table = true).
	TableID TableID `json:"table_id,omitempty"`
	// ExistedTable is the mapping between file columns and table columns.
	// Used when importing to an existing table (new_table = false).
	ExistedTable []FileAndTableColumnMapping `json:"existed_table,omitempty"`
	// ExistedTableOpts denotes the choice when import data into the existed table
	ExistedTableOpts ExistedTableOpts `json:"existed_table_opts,omitempty"`
}

// CreateTableConfig represents the table creation configuration.
type CreateTableConfig struct {
	TableColumn []TableColumn `json:"tableColumn"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
}

// TableColumn represents a column definition in the table configuration.
type TableColumn struct {
	Number         int      `json:"number"`
	ColumnName     string   `json:"columnName"`
	ColumnValues   []string `json:"columnValues"`
	CharNumber     string   `json:"charNumber"`
	CharColumnName string   `json:"charColumnName"`
	Column         string   `json:"column"`
	DataType       string   `json:"dataType"`
	IsKey          bool     `json:"isKey"`
	ColNumInFile   int      `json:"col_num_in_file"`
	Precision      []int    `json:"precision"`
}

// UploadFileResponse represents a response from upload file endpoint.
type UploadFileResponse struct {
	FileID  string              `json:"file_id"`
	Message string              `json:"message"`
	Success bool                `json:"success"`
	Results []*FileUploadResult `json:"results"`
	TaskId  int64               `json:"task_id"`
}

// FileUploadResult represents a single file upload result.
type FileUploadResult struct {
	FileID  string `json:"file_id"`
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// FilePreviewRequest represents a request to preview a file.
type FilePreviewRequest struct {
	// ConnectorId is the connector ID (required for connector file preview)
	ConnectorId uint64 `json:"connector_id"`
	// ConnFileId is the connector rpc file ID (required for local upload file preview)
	ConnFileId string `json:"conn_file_id"`
	// Uri is the file URI (optional, used with connector_id)
	Uri string `json:"uri"`
	// IsColumnName indicates whether to use column names
	IsColumnName bool `json:"isColumnName"`
	// ColumnNameRow is the row number where column names are located
	ColumnNameRow int32 `json:"columnNameRow"`
	// RowStart is the starting row number (must be between 1 and 1000 inclusive)
	RowStart int32 `json:"rowStart"`
	// Csv is the CSV configuration (optional)
	Csv *ConnectorCsvConfig `json:"csv"`
	// FileType is the file type (0 = auto detect, or specific file type)
	FileType int32 `json:"file_type,omitempty"`
}

// ConnectorCsvConfig represents CSV parsing configuration for connector file preview.
type ConnectorCsvConfig struct {
	// Separator is the field separator (default: ",")
	Separator string `json:"separator"`
	// Delimiter is the quote character (default: "\"")
	Delimiter string `json:"delimiter"`
	// IsEscape indicates whether to escape quotes (default: true)
	IsEscape bool `json:"isEscape"`
}

// FilePreviewResponse represents a response from file preview.
type FilePreviewResponse struct {
	// ConnFileId is the connector file ID
	ConnFileId string `json:"conn_file_id"`
	// Rows contains the previewed file rows
	Rows []*PreviewRow `json:"rows"`
	// FileType is the file type
	FileType int32 `json:"file_type"`
}

// PreviewRow represents a single row in file preview.
type PreviewRow struct {
	// Number is the row number
	Number int32 `json:"number"`
	// ColumnName is the column name
	ColumnName string `json:"columnName"`
	// ColumnValues contains the column values for this row
	ColumnValues []string `json:"columnValues"`
	// CharNumber is the Excel column number (e.g., "A", "B", "C")
	CharNumber string `json:"charNumber"`
	// CharColumnName is the Excel column name
	CharColumnName string `json:"charColumnName"`
}

// ConnectorFileDownloadRequest represents a request to generate a download URL
// for a previously uploaded connector file.
type ConnectorFileDownloadRequest struct {
	ConnFileId string `json:"conn_file_id"`
}

// ConnectorFileDownloadResponse contains the signed download URL that can be
// used to download the connector file.
type ConnectorFileDownloadResponse struct {
	URL string `json:"url"`
}

// ConnectorFileDeleteRequest represents a request to delete a connector file by
// its conn_file_id.
type ConnectorFileDeleteRequest struct {
	ConnFileId string `json:"conn_file_id"`
}

// ConnectorFileDeleteResponse represents the response from deleting a connector
// file. The backend currently does not return additional fields, but the type
// is defined for future compatibility.
type ConnectorFileDeleteResponse struct {
	Success bool `json:"success"`
}

// UploadLocalFiles uploads local files to connector.
// files is a map of form field name to file reader and filename.
// meta is the file metadata array in JSON format.
// UploadLocalFiles uploads multiple local files to the connector service.
//
// This method uploads files that will be used for data import tasks. The files
// are uploaded to a temporary storage and you receive conn_file_ids that can be
// used with FilePreview and UploadConnectorFile.
//
// Example:
//
//	file1, _ := os.Open("data1.csv")
//	file2, _ := os.Open("data2.csv")
//	defer file1.Close()
//	defer file2.Close()
//
//	resp, err := client.UploadLocalFiles(ctx, []sdk.FileUploadItem{
//		{File: file1, FileName: "data1.csv"},
//		{File: file2, FileName: "data2.csv"},
//	}, []sdk.FileMeta{
//		{Filename: "data1.csv", Path: "/"},
//		{Filename: "data2.csv", Path: "/"},
//	})
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Uploaded files: %v\n", resp.ConnFileIds)
func (c *RawClient) UploadLocalFiles(ctx context.Context, files []FileUploadItem, meta []FileMeta, opts ...CallOption) (*LocalFileUploadResponse, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("at least one file is required")
	}
	if len(meta) == 0 {
		return nil, fmt.Errorf("meta is required")
	}

	// Create multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add meta field
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return nil, fmt.Errorf("marshal meta: %w", err)
	}
	metaField, err := writer.CreateFormField("meta")
	if err != nil {
		return nil, fmt.Errorf("create meta field: %w", err)
	}
	if _, err := metaField.Write(metaJSON); err != nil {
		return nil, fmt.Errorf("write meta field: %w", err)
	}

	// Add files
	for _, item := range files {
		fileField, err := writer.CreateFormFile("file", item.FileName)
		if err != nil {
			return nil, fmt.Errorf("create file field for %s: %w", item.FileName, err)
		}
		if _, err := io.Copy(fileField, item.File); err != nil {
			return nil, fmt.Errorf("copy file %s: %w", item.FileName, err)
		}
	}

	// Get content type before closing writer
	contentType := writer.FormDataContentType()

	// Close writer to finalize the multipart message
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close multipart writer: %w", err)
	}

	// Make request
	callOpts := newCallOptions(opts...)
	fullURL := c.baseURL + ensureLeadingSlash("/connectors/file/upload")
	if len(callOpts.query) > 0 {
		delimiter := "?"
		if strings.Contains(fullURL, "?") {
			delimiter = "&"
		}
		fullURL = fullURL + delimiter + callOpts.query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", contentType)
	req.Header.Set(headerAPIKey, c.apiKey)
	if c.userAgent != "" {
		req.Header.Set(headerUserAgent, c.userAgent)
	}
	mergeHeaders(req.Header, c.defaultHeaders, false)
	if callOpts.requestID != "" {
		req.Header.Set(headerRequestID, callOpts.requestID)
	}
	mergeHeaders(req.Header, callOpts.headers, true)

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		data, _ := io.ReadAll(resp.Body)
		return nil, &HTTPError{StatusCode: resp.StatusCode, Body: data}
	}

	// Parse response
	var envelope apiEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if envelope.Code != "" && envelope.Code != "OK" {
		return nil, &APIError{
			Code:       envelope.Code,
			Message:    envelope.Msg,
			RequestID:  envelope.RequestID,
			HTTPStatus: resp.StatusCode,
		}
	}

	var uploadResp LocalFileUploadResponse
	if len(envelope.Data) > 0 && string(envelope.Data) != "null" {
		if err := json.Unmarshal(envelope.Data, &uploadResp); err != nil {
			return nil, fmt.Errorf("decode data field: %w", err)
		}
	}

	return &uploadResp, nil
}

// UploadLocalFile uploads a single local file to the connector service.
//
// This is a convenience method that wraps UploadLocalFiles for a single file.
// The fileReader provides the file content, fileName is the filename, and meta
// is the file metadata array.
//
// Example:
//
//	file, _ := os.Open("data.csv")
//	defer file.Close()
//
//	resp, err := client.UploadLocalFile(ctx, file, "data.csv", []sdk.FileMeta{
//		{Filename: "data.csv", Path: "/"},
//	})
//	if err != nil {
//		return err
//	}
//	connFileID := resp.ConnFileIds[0]
func (c *RawClient) UploadLocalFile(ctx context.Context, fileReader io.Reader, fileName string, meta []FileMeta, opts ...CallOption) (*LocalFileUploadResponse, error) {
	return c.UploadLocalFiles(ctx, []FileUploadItem{
		{
			File:     fileReader,
			FileName: fileName,
		},
	}, meta, opts...)
}

// UploadLocalFileFromPath uploads a local file from the file system path to the connector service.
//
// This is a convenience method that opens the file from the given path and uploads it.
// The filename is automatically extracted from the path.
//
// Example:
//
//	resp, err := client.UploadLocalFileFromPath(ctx, "/path/to/data.csv", []sdk.FileMeta{
//		{Filename: "data.csv", Path: "/"},
//	})
//	if err != nil {
//		return err
//	}
//	connFileID := resp.ConnFileIds[0]
func (c *RawClient) UploadLocalFileFromPath(ctx context.Context, filePath string, meta []FileMeta, opts ...CallOption) (*LocalFileUploadResponse, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file %s: %w", filePath, err)
	}
	defer file.Close()

	// Extract filename from path
	fileName := filepath.Base(filePath)

	return c.UploadLocalFile(ctx, file, fileName, meta, opts...)
}

// FilePreview previews a file from connector or local upload to analyze its structure.
//
// The request must specify either:
//   - connector_id with uri or conn_file_id (for connector files)
//   - conn_file_id without connector_id (for local upload files)
//
// The response includes file structure information such as column names, data types,
// and sample data, which can be used to build TableConfig for data import.
//
// Example:
//
//	resp, err := client.FilePreview(ctx, &sdk.FilePreviewRequest{
//		ConnFileId: "conn-file-id-123",
//	})
//	if err != nil {
//		return err
//	}
//	// Use resp to build TableConfig for import
//	for _, col := range resp.TableColumn {
//		fmt.Printf("Column: %s, Type: %s\n", col.ColumnName, col.DataType)
//	}
func (c *RawClient) FilePreview(ctx context.Context, req *FilePreviewRequest, opts ...CallOption) (*FilePreviewResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	// Validate request parameters
	if req.ConnectorId > 0 {
		// Connector file preview: need uri or conn_file_id
		if strings.TrimSpace(req.Uri) == "" && strings.TrimSpace(req.ConnFileId) == "" {
			return nil, fmt.Errorf("file preview needs uri or conn_file_id when connector_id is provided")
		}
	} else {
		// Local upload file preview: need conn_file_id
		if strings.TrimSpace(req.ConnFileId) == "" {
			return nil, fmt.Errorf("file preview needs conn_file_id for local upload file")
		}
	}

	// Make request
	callOpts := newCallOptions(opts...)
	fullURL := c.baseURL + ensureLeadingSlash("/connectors/file/preview")

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, strings.NewReader(string(reqBody)))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set(headerContentType, mimeJSON)
	httpReq.Header.Set(headerAccept, mimeJSON)
	httpReq.Header.Set(headerAPIKey, c.apiKey)
	if c.userAgent != "" {
		httpReq.Header.Set(headerUserAgent, c.userAgent)
	}
	mergeHeaders(httpReq.Header, c.defaultHeaders, false)
	if callOpts.requestID != "" {
		httpReq.Header.Set(headerRequestID, callOpts.requestID)
	}
	mergeHeaders(httpReq.Header, callOpts.headers, true)

	// Execute request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		data, _ := io.ReadAll(resp.Body)
		return nil, &HTTPError{StatusCode: resp.StatusCode, Body: data}
	}

	// Parse response
	var envelope apiEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if envelope.Code != "" && envelope.Code != "OK" {
		return nil, &APIError{
			Code:       envelope.Code,
			Message:    envelope.Msg,
			RequestID:  envelope.RequestID,
			HTTPStatus: resp.StatusCode,
		}
	}

	var previewResp FilePreviewResponse
	if len(envelope.Data) > 0 && string(envelope.Data) != "null" {
		if err := json.Unmarshal(envelope.Data, &previewResp); err != nil {
			return nil, fmt.Errorf("decode data field: %w", err)
		}
	}

	return &previewResp, nil
}

// UploadConnectorFile uploads files to connector and creates a data import task.
//
// This endpoint supports advanced features like file filtering, deduplication, and table configuration.
// It can either upload new files or reference already uploaded files via TableConfig.ConnFileIDs.
//
// Note: This is different from the UploadFile method in file.go which uploads to /catalog/file/upload.
//
// Example - Upload new files and import to new table:
//
//	file, _ := os.Open("data.csv")
//	defer file.Close()
//
//	resp, err := client.UploadConnectorFile(ctx, &sdk.UploadFileRequest{
//		VolumeID: "123456",
//		Files: []sdk.FileUploadItem{
//			{File: file, FileName: "data.csv"},
//		},
//		Meta: []sdk.FileMeta{
//			{Filename: "data.csv", Path: "/"},
//		},
//		TableConfig: &sdk.TableConfig{
//			NewTable:    true,
//			DatabaseID:  123,
//			ConnFileIDs: []string{}, // Will be filled from uploaded files
//			// ... other table config
//		},
//	})
//
// Example - Import already uploaded files to existing table:
//
//	resp, err := client.UploadConnectorFile(ctx, &sdk.UploadFileRequest{
//		VolumeID: "123456",
//		Files:    []sdk.FileUploadItem{}, // Empty, files already uploaded
//		Meta:     []sdk.FileMeta{{Filename: "data.csv", Path: "/"}},
//		TableConfig: &sdk.TableConfig{
//			NewTable:    false,
//			DatabaseID:  123,
//			TableID:     456,
//			ConnFileIDs: []string{"conn-file-id-123"},
//			// ... column mappings
//		},
//	})
func (c *RawClient) UploadConnectorFile(ctx context.Context, req *UploadFileRequest, opts ...CallOption) (*UploadFileResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	if req.VolumeID == "" {
		return nil, fmt.Errorf("volume_id is required")
	}
	// Allow empty Files if TableConfig.ConnFileIDs is provided (files already uploaded)
	if len(req.Files) == 0 && (req.TableConfig == nil || len(req.TableConfig.ConnFileIDs) == 0) {
		return nil, fmt.Errorf("at least one file is required, or TableConfig.ConnFileIDs must be provided")
	}

	// Create multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add VolumeID field (required)
	volumeIDField, err := writer.CreateFormField("VolumeID")
	if err != nil {
		return nil, fmt.Errorf("create VolumeID field: %w", err)
	}
	if _, err := volumeIDField.Write([]byte(string(req.VolumeID))); err != nil {
		return nil, fmt.Errorf("write VolumeID field: %w", err)
	}

	// Add meta field (optional)
	if len(req.Meta) > 0 {
		metaJSON, err := json.Marshal(req.Meta)
		if err != nil {
			return nil, fmt.Errorf("marshal meta: %w", err)
		}
		metaField, err := writer.CreateFormField("meta")
		if err != nil {
			return nil, fmt.Errorf("create meta field: %w", err)
		}
		if _, err := metaField.Write(metaJSON); err != nil {
			return nil, fmt.Errorf("write meta field: %w", err)
		}
	}

	// Add file_types field (optional)
	if len(req.FileTypes) > 0 {
		fileTypesJSON, err := json.Marshal(req.FileTypes)
		if err != nil {
			return nil, fmt.Errorf("marshal file_types: %w", err)
		}
		fileTypesField, err := writer.CreateFormField("file_types")
		if err != nil {
			return nil, fmt.Errorf("create file_types field: %w", err)
		}
		if _, err := fileTypesField.Write(fileTypesJSON); err != nil {
			return nil, fmt.Errorf("write file_types field: %w", err)
		}
	}

	// Add path_regex field (optional)
	if req.PathRegex != "" {
		pathRegexField, err := writer.CreateFormField("path_regex")
		if err != nil {
			return nil, fmt.Errorf("create path_regex field: %w", err)
		}
		if _, err := pathRegexField.Write([]byte(req.PathRegex)); err != nil {
			return nil, fmt.Errorf("write path_regex field: %w", err)
		}
	}

	// Add unzip_keep_structure field (optional)
	if req.UnzipKeepStructure {
		unzipField, err := writer.CreateFormField("unzip_keep_structure")
		if err != nil {
			return nil, fmt.Errorf("create unzip_keep_structure field: %w", err)
		}
		if _, err := unzipField.Write([]byte("true")); err != nil {
			return nil, fmt.Errorf("write unzip_keep_structure field: %w", err)
		}
	}

	// Add dedup field (optional)
	if req.DedupConfig != nil {
		dedupJSON, err := json.Marshal(req.DedupConfig)
		if err != nil {
			return nil, fmt.Errorf("marshal dedup: %w", err)
		}
		dedupField, err := writer.CreateFormField("dedup")
		if err != nil {
			return nil, fmt.Errorf("create dedup field: %w", err)
		}
		if _, err := dedupField.Write(dedupJSON); err != nil {
			return nil, fmt.Errorf("write dedup field: %w", err)
		}
	}

	// Add table_config field (optional)
	if req.TableConfig != nil {
		tableConfigJSON, err := json.Marshal(req.TableConfig)
		if err != nil {
			return nil, fmt.Errorf("marshal table_config: %w", err)
		}
		tableConfigField, err := writer.CreateFormField("table_config")
		if err != nil {
			return nil, fmt.Errorf("create table_config field: %w", err)
		}
		if _, err := tableConfigField.Write(tableConfigJSON); err != nil {
			return nil, fmt.Errorf("write table_config field: %w", err)
		}
	}

	// Add files (required, unless TableConfig.ConnFileIDs is provided)
	for _, item := range req.Files {
		fileField, err := writer.CreateFormFile("file", item.FileName)
		if err != nil {
			return nil, fmt.Errorf("create file field for %s: %w", item.FileName, err)
		}
		if _, err := io.Copy(fileField, item.File); err != nil {
			return nil, fmt.Errorf("copy file %s: %w", item.FileName, err)
		}
	}

	// Get content type before closing writer
	contentType := writer.FormDataContentType()

	// Close writer to finalize the multipart message
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close multipart writer: %w", err)
	}

	// Make request
	callOpts := newCallOptions(opts...)
	fullURL := c.baseURL + ensureLeadingSlash("/connectors/upload")
	if len(callOpts.query) > 0 {
		delimiter := "?"
		if strings.Contains(fullURL, "?") {
			delimiter = "&"
		}
		fullURL = fullURL + delimiter + callOpts.query.Encode()
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", contentType)
	httpReq.Header.Set(headerAPIKey, c.apiKey)
	if c.userAgent != "" {
		httpReq.Header.Set(headerUserAgent, c.userAgent)
	}
	mergeHeaders(httpReq.Header, c.defaultHeaders, false)
	if callOpts.requestID != "" {
		httpReq.Header.Set(headerRequestID, callOpts.requestID)
	}
	mergeHeaders(httpReq.Header, callOpts.headers, true)

	// Execute request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		data, _ := io.ReadAll(resp.Body)
		return nil, &HTTPError{StatusCode: resp.StatusCode, Body: data}
	}

	// Parse response
	var envelope apiEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if envelope.Code != "" && envelope.Code != "OK" {
		return nil, &APIError{
			Code:       envelope.Code,
			Message:    envelope.Msg,
			RequestID:  envelope.RequestID,
			HTTPStatus: resp.StatusCode,
		}
	}

	var uploadResp UploadFileResponse
	if len(envelope.Data) > 0 && string(envelope.Data) != "null" {
		if err := json.Unmarshal(envelope.Data, &uploadResp); err != nil {
			return nil, fmt.Errorf("decode data field: %w", err)
		}
	}

	return &uploadResp, nil
}

// DownloadConnectorFile retrieves a signed download URL for a connector file.
//
// Example:
//
//	resp, err := client.DownloadConnectorFile(ctx, &sdk.ConnectorFileDownloadRequest{
//		ConnFileId: "conn-file-id-123",
//	})
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Download URL: %s\n", resp.URL)
func (c *RawClient) DownloadConnectorFile(ctx context.Context, req *ConnectorFileDownloadRequest, opts ...CallOption) (*ConnectorFileDownloadResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	if strings.TrimSpace(req.ConnFileId) == "" {
		return nil, fmt.Errorf("conn_file_id is required")
	}

	var resp ConnectorFileDownloadResponse
	if err := c.postJSON(ctx, "/connectors/file/download", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteConnectorFile deletes a connector file by its conn_file_id.
//
// Example:
//
//	_, err := client.DeleteConnectorFile(ctx, &sdk.ConnectorFileDeleteRequest{
//		ConnFileId: "conn-file-id-123",
//	})
func (c *RawClient) DeleteConnectorFile(ctx context.Context, req *ConnectorFileDeleteRequest, opts ...CallOption) (*ConnectorFileDeleteResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	if strings.TrimSpace(req.ConnFileId) == "" {
		return nil, fmt.Errorf("conn_file_id is required")
	}

	var resp ConnectorFileDeleteResponse
	if err := c.postJSON(ctx, "/connectors/file/delete", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}
