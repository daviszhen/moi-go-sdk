package sdk

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SDKClient is a high-level client that provides convenient business-oriented APIs.
// It wraps RawClient and combines multiple raw API calls to implement higher-level functionality.
type SDKClient struct {
	raw *RawClient
}

// NewSDKClient creates a new high-level SDK client using the provided RawClient.
func NewSDKClient(raw *RawClient) *SDKClient {
	if raw == nil {
		panic("RawClient cannot be nil")
	}
	return &SDKClient{
		raw: raw,
	}
}

// WithSpecialUser creates a new SDKClient with the same configuration but a different API key.
// The cloned client uses a cloned RawClient with the new API key.
// Panics if the client is nil or if the API key is empty.
func (c *SDKClient) WithSpecialUser(apiKey string) *SDKClient {
	if c == nil {
		panic("cannot clone nil client")
	}
	clonedRaw := c.raw.WithSpecialUser(apiKey)
	return &SDKClient{
		raw: clonedRaw,
	}
}

// TablePrivInfo represents table privilege information for role creation.
type TablePrivInfo struct {
	// TableID is the table ID
	TableID TableID
	// PrivCodes are the privilege codes for this table (deprecated, use AuthorityCodeList instead)
	PrivCodes []PrivCode
	// AuthorityCodeList contains privilege codes with optional rules for this table
	// If both PrivCodes and AuthorityCodeList are provided, AuthorityCodeList takes precedence
	AuthorityCodeList []*AuthorityCodeAndRule
}

// CreateTableRole creates a role for table privileges, or returns the existing role if it already exists.
//
// It first queries for the role by name using RawClient. If the role exists, it returns
// the role ID and created=false. If the role doesn't exist, it creates a new role with
// the specified name, comment, and table privileges.
//
// The tablePrivs parameter can use either AuthorityCodeList (recommended, supports rules)
// or PrivCodes (deprecated, for backward compatibility). If both are provided,
// AuthorityCodeList takes precedence.
//
// Example:
//
//	roleID, created, err := sdkClient.CreateTableRole(ctx, "my-role", "Role description", []sdk.TablePrivInfo{
//		{
//			TableID: 123,
//			AuthorityCodeList: []*sdk.AuthorityCodeAndRule{
//				{
//					Code:     "DT8", // SELECT permission
//					RuleList: nil,
//				},
//			},
//		},
//	})
//	if err != nil {
//		return err
//	}
//	if created {
//		fmt.Printf("Created new role: %d\n", roleID)
//	} else {
//		fmt.Printf("Role already exists: %d\n", roleID)
//	}
//
// Returns:
//   - roleID: the ID of the role (existing or newly created)
//   - created: true if the role was newly created, false if it already existed
//   - error: any error that occurred
func (c *SDKClient) CreateTableRole(ctx context.Context, roleName string, comment string, tablePrivs []TablePrivInfo) (roleID RoleID, created bool, err error) {
	if roleName == "" {
		return 0, false, fmt.Errorf("role name is required")
	}

	// Step 1: Query for existing role by name using filters (as per frontend example)
	// Use server-side filter with fuzzy search, then verify exact match client-side
	var existingRole *RoleInfoResponse
	page := 1
	pageSize := 100
	maxPages := 1000 // Safety limit to avoid infinite loops

	for page <= maxPages {
		// Use filters to search by role name (matching frontend example format)
		roleListReq := &RoleListRequest{
			Keyword: "",
			CommonCondition: CommonCondition{
				Page:     page,
				PageSize: pageSize,
				Order:    "desc",
				OrderBy:  "created_at",
				Filters: []CommonFilter{
					{
						Name:   "name_description",
						Values: []string{roleName},
						Fuzzy:  true,
					},
				},
			},
		}

		roleListResp, err := c.raw.ListRoles(ctx, roleListReq)
		if err != nil {
			return 0, false, err
		}

		if roleListResp == nil || len(roleListResp.List) == 0 {
			// No more roles to check
			break
		}

		// Check if role with exact name exists in current page
		for i := range roleListResp.List {
			if roleListResp.List[i].RoleName == roleName {
				existingRole = &roleListResp.List[i]
				break
			}
		}

		if existingRole != nil {
			// Found the role
			break
		}

		// Check if there are more pages
		// Stop if current page has fewer results than pageSize (indicates last page)
		if len(roleListResp.List) < pageSize {
			// No more pages (last page returned fewer items than pageSize)
			break
		}

		// Also check Total to avoid infinite loops
		// If we've processed all items according to Total, stop
		if roleListResp.Total > 0 && page*pageSize >= roleListResp.Total {
			// Reached the total number of roles
			break
		}

		// Continue to next page
		page++
	}

	// Step 2: If role exists, return its ID
	if existingRole != nil {
		return existingRole.RoleID, false, nil
	}

	// Step 3: Convert table privilege info to ObjPrivResponse
	objPrivList := make([]ObjPrivResponse, 0, len(tablePrivs))
	for _, tablePriv := range tablePrivs {
		var authorityCodeList []*AuthorityCodeAndRule

		// Use AuthorityCodeList if provided, otherwise fall back to PrivCodes for backward compatibility
		if len(tablePriv.AuthorityCodeList) > 0 {
			// Use the provided AuthorityCodeList with rules
			authorityCodeList = tablePriv.AuthorityCodeList
		} else if len(tablePriv.PrivCodes) > 0 {
			// Convert PrivCode slice to AuthorityCodeAndRule slice (backward compatibility)
			authorityCodeList = make([]*AuthorityCodeAndRule, 0, len(tablePriv.PrivCodes))
			for _, privCode := range tablePriv.PrivCodes {
				authorityCodeList = append(authorityCodeList, &AuthorityCodeAndRule{
					Code:     string(privCode),
					RuleList: nil, // No rules by default
				})
			}
		} else {
			// Skip if neither is provided
			continue
		}

		objPrivList = append(objPrivList, ObjPrivResponse{
			ObjID:             fmt.Sprintf("%d", tablePriv.TableID),
			ObjType:           ObjTypeTable.String(), // "table"
			ObjName:           "",                    // Table name is optional, can be left empty
			AuthorityCodeList: authorityCodeList,
		})
	}

	// Step 4: Create new role
	createReq := &RoleCreateRequest{
		RoleName:    roleName,
		Comment:     comment,
		PrivList:    []string{}, // No global privileges, only object-level privileges
		ObjPrivList: objPrivList,
	}

	createResp, err := c.raw.CreateRole(ctx, createReq)
	if err != nil {
		// If creation fails due to role already existing, try to find it again
		// This handles the case where ListRoles failed but the role exists
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr != nil {
			// Check if error indicates role already exists
			errMsg := strings.ToLower(apiErr.Message)
			if strings.Contains(errMsg, "already exists") || strings.Contains(errMsg, "duplicate") {
				// Try to list roles one more time to find the existing role with pagination
				// Use the same pagination logic as initial search
				retryPage := 1
				retryPageSize := 100
				retryMaxPages := 1000 // Safety limit
				for retryPage <= retryMaxPages {
					retryListReq := &RoleListRequest{
						Keyword: "",
						CommonCondition: CommonCondition{
							Page:     retryPage,
							PageSize: retryPageSize,
							Order:    "desc",
							OrderBy:  "created_at",
							Filters: []CommonFilter{
								{
									Name:   "name_description",
									Values: []string{roleName},
									Fuzzy:  true,
								},
							},
						},
					}
					retryListResp, retryErr := c.raw.ListRoles(ctx, retryListReq)
					if retryErr != nil {
						// If listing fails for this page, try next page (might be a transient error)
						// But if it's the first page, break
						if retryPage == 1 {
							break
						}
						// For subsequent pages, if error occurs, assume we've reached the end
						break
					}

					if retryListResp == nil || len(retryListResp.List) == 0 {
						// No more results
						break
					}

					// Search for the role by name in current page
					for i := range retryListResp.List {
						if retryListResp.List[i].RoleName == roleName {
							return retryListResp.List[i].RoleID, false, nil
						}
					}

					// Check if there are more pages
					// Stop if current page has fewer results than pageSize
					if len(retryListResp.List) < retryPageSize {
						// No more pages
						break
					}

					// Also check Total to avoid infinite loops
					if retryListResp.Total > 0 && retryPage*retryPageSize >= retryListResp.Total {
						// Reached the total number of roles
						break
					}

					// Continue to next page
					retryPage++
				}
				// If ListRoles still fails, we can't find the role, but we know it exists
				// Return a more user-friendly error message
				return 0, false, fmt.Errorf("role '%s' already exists but could not be retrieved", roleName)
			}
		}
		return 0, false, fmt.Errorf("failed to create role: %w", err)
	}

	return createResp.RoleID, true, nil
}

// UpdateTableRole updates an existing role with table privileges.
//
// It updates the role's object-level privileges (table privileges) while preserving
// or updating global privileges. The comment and globalPrivs parameters have special
// semantics:
//   - If comment is empty, the existing comment will be preserved
//   - If globalPrivs is nil, existing global privileges will be preserved
//   - If globalPrivs is an empty slice, all global privileges will be removed
//
// The tablePrivs parameter can use either AuthorityCodeList (recommended, supports rules)
// or PrivCodes (deprecated, for backward compatibility). If both are provided,
// AuthorityCodeList takes precedence.
//
// Example:
//
//	// Update table privileges, preserve comment and global privileges
//	err := sdkClient.UpdateTableRole(ctx, 456, "", []sdk.TablePrivInfo{
//		{
//			TableID: 123,
//			AuthorityCodeList: []*sdk.AuthorityCodeAndRule{
//				{Code: "DT8", RuleList: nil},
//			},
//		},
//	}, nil)
//
//	// Update comment and global privileges
//	err := sdkClient.UpdateTableRole(ctx, 456, "New description", []sdk.TablePrivInfo{
//		// ... table privileges
//	}, []string{"U1", "R1"})
//
//	// Remove all global privileges
//	err := sdkClient.UpdateTableRole(ctx, 456, "", []sdk.TablePrivInfo{
//		// ... table privileges
//	}, []string{})
func (c *SDKClient) UpdateTableRole(ctx context.Context, roleID RoleID, comment string, tablePrivs []TablePrivInfo, globalPrivs []string) error {
	if roleID == 0 {
		return fmt.Errorf("role_id is required")
	}

	// Step 1: Get current role info if needed (to preserve comment or global privileges)
	var currentComment string
	var privList []string
	if comment == "" || globalPrivs == nil {
		roleInfo, err := c.raw.GetRole(ctx, &RoleInfoRequest{RoleID: roleID})
		if err != nil {
			return fmt.Errorf("failed to get role info: %w", err)
		}

		// Preserve comment if not provided
		if comment == "" {
			currentComment = roleInfo.Comment
		} else {
			currentComment = comment
		}

		// Preserve global privileges if not provided
		if globalPrivs == nil {
			// Extract global privilege codes from AuthorityList
			privList = make([]string, 0, len(roleInfo.AuthorityList))
			for _, priv := range roleInfo.AuthorityList {
				privList = append(privList, priv.PrivCode)
			}
		} else {
			// Use provided global privileges
			privList = globalPrivs
		}
	} else {
		// Both comment and globalPrivs are provided, no need to fetch role info
		currentComment = comment
		privList = globalPrivs
	}

	// Step 3: Convert table privilege info to ObjPrivResponse
	objPrivList := make([]ObjPrivResponse, 0, len(tablePrivs))
	for _, tablePriv := range tablePrivs {
		var authorityCodeList []*AuthorityCodeAndRule

		// Use AuthorityCodeList if provided, otherwise fall back to PrivCodes for backward compatibility
		if len(tablePriv.AuthorityCodeList) > 0 {
			// Use the provided AuthorityCodeList with rules
			authorityCodeList = tablePriv.AuthorityCodeList
		} else if len(tablePriv.PrivCodes) > 0 {
			// Convert PrivCode slice to AuthorityCodeAndRule slice (backward compatibility)
			authorityCodeList = make([]*AuthorityCodeAndRule, 0, len(tablePriv.PrivCodes))
			for _, privCode := range tablePriv.PrivCodes {
				authorityCodeList = append(authorityCodeList, &AuthorityCodeAndRule{
					Code:     string(privCode),
					RuleList: nil, // No rules by default
				})
			}
		} else {
			// Skip if neither is provided
			continue
		}

		objPrivList = append(objPrivList, ObjPrivResponse{
			ObjID:             fmt.Sprintf("%d", tablePriv.TableID),
			ObjType:           ObjTypeTable.String(), // "table"
			ObjName:           "",                    // Table name is optional, can be left empty
			AuthorityCodeList: authorityCodeList,
		})
	}

	// Step 4: Update role
	updateReq := &RoleUpdateInfoRequest{
		RoleID:      roleID,
		PrivList:    privList,
		ObjPrivList: objPrivList,
		Comment:     currentComment,
	}

	_, err := c.raw.UpdateRoleInfo(ctx, updateReq)
	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	return nil
}

// ImportLocalFileToTable imports a local file (already uploaded via UploadLocalFile) to a table.
// This is a high-level convenience method that simplifies the process of importing a file to a table.
// The method automatically determines whether to create a new table or import to an existing table
// based on the tableConfig.NewTable field.
//
// Parameters:
//   - tableConfig: the table configuration built from FilePreview results (required)
//     The tableConfig must have:
//   - ConnFileIDs: at least one file ID from UploadLocalFile (required)
//   - NewTable: true to create a new table, false to import to an existing table
//   - TableID: the target table ID (required when NewTable = false)
//   - ExistedTable: the mapping between file columns and table columns (optional when NewTable = false, but recommended)
//
// Returns:
//   - *UploadFileResponse: the response from the upload operation
//   - error: any error that occurred
//
// Note: This method uses magic values for VolumeID ("123456") and constructs Meta from the first conn_file_id.
// The Files field in UploadFileRequest is set to empty, as the file is already uploaded and referenced by conn_file_id.
func (c *SDKClient) ImportLocalFileToTable(ctx context.Context, tableConfig *TableConfig) (*UploadFileResponse, error) {
	if tableConfig == nil {
		return nil, fmt.Errorf("table_config is required")
	}
	if len(tableConfig.ConnFileIDs) == 0 {
		return nil, fmt.Errorf("table_config.conn_file_ids is required and must contain at least one file ID")
	}

	// Validate based on NewTable value
	if !tableConfig.NewTable {
		// Importing to an existing table: TableID is required
		if tableConfig.TableID == 0 {
			return nil, fmt.Errorf("table_config.table_id is required when importing to an existing table (new_table = false)")
		}
		// Ensure existed_table is initialized (even if empty, to avoid nil)
		if tableConfig.ExistedTable == nil {
			tableConfig.ExistedTable = []FileAndTableColumnMapping{}
		}
	}
	// For new table, no additional validation needed

	// Get the first conn_file_id for metadata
	connFileID := tableConfig.ConnFileIDs[0]

	// Use magic value for VolumeID as per requirements
	volumeID := VolumeID("123456")

	// Use connFileID as filename and default path
	filename := connFileID
	path := "/"

	// Build Meta using magic value format: [{"filename":"<filename>","path":"<path>"}]
	meta := []FileMeta{
		{
			Filename: filename,
			Path:     path,
		},
	}

	// Build UploadFileRequest
	// Note: Files is set to empty slice as the file is already uploaded and referenced by conn_file_id
	// The backend should use the conn_file_id from tableConfig.ConnFileIDs
	uploadReq := &UploadFileRequest{
		VolumeID:    volumeID,
		Files:       []FileUploadItem{}, // Empty, as file is already uploaded
		Meta:        meta,
		TableConfig: tableConfig,
	}

	// Call the raw client's UploadConnectorFile method
	return c.raw.UploadConnectorFile(ctx, uploadReq)
}

// ImportLocalFileToVolume uploads a local unstructured file to a target volume.
// This is a high-level convenience method that uploads a local file to a volume
// with metadata and deduplication configuration.
//
// Parameters:
//   - filePath: the local file path to upload (required)
//   - volumeID: the target volume ID (required)
//   - meta: file metadata describing the file location in the target volume (required)
//     Format: {"filename":"研发过程安全分析 202504.docx","path":"研发过程安全分析 202504.docx"}
//   - dedup: deduplication configuration (optional)
//     Format: {"by":["name","md5"],"strategy":"skip"}
//
// Returns:
//   - *UploadFileResponse: the response from the upload operation
//   - error: any error that occurred
//
// Example:
//
//	resp, err := sdkClient.ImportLocalFileToVolume(ctx, "/path/to/file.docx", "123456", sdk.FileMeta{
//		Filename: "研发过程安全分析 202504.docx",
//		Path:     "研发过程安全分析 202504.docx",
//	}, &sdk.DedupConfig{
//		By:       []string{"name", "md5"},
//		Strategy: "skip",
//	})
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Uploaded file: %s\n", resp.FileID)
func (c *SDKClient) ImportLocalFileToVolume(ctx context.Context, filePath string, volumeID VolumeID, meta FileMeta, dedup *DedupConfig, opts ...CallOption) (*UploadFileResponse, error) {
	if strings.TrimSpace(filePath) == "" {
		return nil, fmt.Errorf("file_path is required")
	}
	if volumeID == "" {
		return nil, fmt.Errorf("volume_id is required")
	}
	if strings.TrimSpace(meta.Filename) == "" {
		return nil, fmt.Errorf("meta.filename is required")
	}

	// Open the local file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file %s: %w", filePath, err)
	}
	defer file.Close()

	// Extract filename from path
	fileName := filepath.Base(filePath)

	// Build UploadFileRequest
	// Wrap meta in an array as required by UploadConnectorFile
	uploadReq := &UploadFileRequest{
		VolumeID: volumeID,
		Files: []FileUploadItem{
			{
				File:     file,
				FileName: fileName,
			},
		},
		Meta:        []FileMeta{meta},
		DedupConfig: dedup,
	}

	// Call the raw client's UploadConnectorFile method
	return c.raw.UploadConnectorFile(ctx, uploadReq, opts...)
}

// ImportLocalFilesToVolume uploads multiple local unstructured files to a target volume.
// This is a high-level convenience method that uploads multiple local files to a volume
// with metadata and deduplication configuration.
//
// Parameters:
//   - filePaths: array of local file paths to upload (required, must not be empty)
//   - volumeID: the target volume ID (required)
//   - metas: array of file metadata describing the file locations in the target volume (optional)
//     If provided, must have the same length as filePaths.
//     If empty or nil, metadata will be auto-generated from file paths.
//     Format: [{"filename":"file1.docx","path":"file1.docx"}, {"filename":"file2.docx","path":"file2.docx"}]
//   - dedup: deduplication configuration (optional, applied to all files)
//     Format: {"by":["name","md5"],"strategy":"skip"}
//
// Returns:
//   - *UploadFileResponse: the response from the upload operation
//   - error: any error that occurred
//
// Example:
//
//	resp, err := sdkClient.ImportLocalFilesToVolume(ctx, []string{
//		"/path/to/file1.docx",
//		"/path/to/file2.docx",
//	}, "123456", []sdk.FileMeta{
//		{Filename: "file1.docx", Path: "file1.docx"},
//		{Filename: "file2.docx", Path: "file2.docx"},
//	}, &sdk.DedupConfig{
//		By:       []string{"name", "md5"},
//		Strategy: "skip",
//	})
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Uploaded files, task_id: %d\n", resp.TaskId)
func (c *SDKClient) ImportLocalFilesToVolume(ctx context.Context, filePaths []string, volumeID VolumeID, metas []FileMeta, dedup *DedupConfig, opts ...CallOption) (*UploadFileResponse, error) {
	if len(filePaths) == 0 {
		return nil, fmt.Errorf("at least one file path is required")
	}
	if volumeID == "" {
		return nil, fmt.Errorf("volume_id is required")
	}

	// Validate metas if provided
	if len(metas) > 0 && len(metas) != len(filePaths) {
		return nil, fmt.Errorf("metas array length (%d) must match filePaths length (%d)", len(metas), len(filePaths))
	}

	// Open all files and build file upload items
	files := make([]FileUploadItem, 0, len(filePaths))
	fileMetas := make([]FileMeta, 0, len(filePaths))
	fileHandles := make([]*os.File, 0, len(filePaths))

	// Cleanup function to close all opened files
	cleanup := func() {
		for _, f := range fileHandles {
			if f != nil {
				f.Close()
			}
		}
	}
	defer cleanup()

	for i, filePath := range filePaths {
		if strings.TrimSpace(filePath) == "" {
			return nil, fmt.Errorf("file_path[%d] is empty", i)
		}

		// Open the local file
		file, err := os.Open(filePath)
		if err != nil {
			// Close already opened files before returning error
			cleanup()
			return nil, fmt.Errorf("open file %s: %w", filePath, err)
		}
		fileHandles = append(fileHandles, file)

		// Extract filename from path
		fileName := filepath.Base(filePath)

		// Build file upload item
		files = append(files, FileUploadItem{
			File:     file,
			FileName: fileName,
		})

		// Build meta - use provided meta or auto-generate from file path
		if i < len(metas) && strings.TrimSpace(metas[i].Filename) != "" {
			// Use provided meta
			fileMetas = append(fileMetas, metas[i])
		} else {
			// Auto-generate meta from file path
			fileMetas = append(fileMetas, FileMeta{
				Filename: fileName,
				Path:     fileName,
			})
		}
	}

	// Build UploadFileRequest
	uploadReq := &UploadFileRequest{
		VolumeID:    volumeID,
		Files:       files,
		Meta:        fileMetas,
		DedupConfig: dedup,
	}

	// Call the raw client's UploadConnectorFile method
	// Note: We need to keep files open until the request completes, so we don't defer close here
	// The files will be closed by the defer function above after the method returns
	return c.raw.UploadConnectorFile(ctx, uploadReq, opts...)
}

// RunSQL executes a SQL statement using the NL2SQL RunSQL operation.
//
// The statement must reference tables using fully qualified names (database.table).
// This requirement allows the catalog service to route the query to the correct database.
func (c *SDKClient) RunSQL(ctx context.Context, statement string, opts ...CallOption) (*NL2SQLRunSQLResponse, error) {
	if strings.TrimSpace(statement) == "" {
		return nil, fmt.Errorf("statement is required")
	}
	return c.raw.RunNL2SQL(ctx, &NL2SQLRunSQLRequest{
		Operation: RunSQL,
		Statement: statement,
	}, opts...)
}

// CreateDocumentProcessingWorkflow creates a workflow for processing documents from a source volume to a target volume.
//
// This is a high-level convenience method that creates a complete document processing pipeline
// with the following nodes:
//   - RootNode: Reads files from the source volume
//   - DocumentParseNode: Parses various document formats
//   - ChunkNode: Splits documents into chunks
//   - EmbedNode: Generates embeddings for document chunks
//   - WriteNode: Writes processed results to the target volume
//
// The workflow is configured to trigger automatically when files are loaded into the source volume
// (ProcessMode.Interval = -1).
//
// Supported file types:
//   - Text files: TXT (1), Markdown (6), HTM (27), HTML (28)
//   - Office documents: PDF (2), PPT (4), DOCX (11), PPTX (12), XLS (24), XLSX (25)
//   - Spreadsheets: CSV (7)
//
// Parameters:
//   - targetVolumeID: the target volume ID where processed results will be written (required)
//   - sourceVolumeID: the source volume ID where source documents are located (required)
//   - workflowName: the name of the workflow (required)
//
// Returns:
//   - workflowID: the ID of the created workflow
//   - error: any error that occurred
//
// Example:
//
//	workflowID, err := sdkClient.CreateDocumentProcessingWorkflow(ctx, "target-vol-123", "source-vol-456", "My Workflow")
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Created workflow: %s\n", workflowID)
func (c *SDKClient) CreateDocumentProcessingWorkflow(ctx context.Context, workflowName string, sourceVolumeID VolumeID, targetVolumeID VolumeID, opts ...CallOption) (workflowID string, err error) {
	if strings.TrimSpace(string(targetVolumeID)) == "" {
		return "", fmt.Errorf("target_volume_id is required")
	}
	if strings.TrimSpace(string(sourceVolumeID)) == "" {
		return "", fmt.Errorf("source_volume_id is required")
	}
	if strings.TrimSpace(workflowName) == "" {
		return "", fmt.Errorf("workflow_name is required")
	}

	// Build the workflow metadata with a complete document processing pipeline
	req := &WorkflowMetadata{
		Name:            workflowName,
		SourceVolumeIDs: []string{string(sourceVolumeID)},
		TargetVolumeID:  string(targetVolumeID),
		// Supported file types: TXT, PDF, PPT, DOCX, Markdown, PPTX, CSV, XLS, XLSX, HTM, HTML
		FileTypes: []int{
			int(FileTypeTXT), int(FileTypePDF), int(FileTypePPT), int(FileTypeDOCX),
			int(FileTypeMarkdown), int(FileTypePPTX), int(FileTypeCSV),
			int(FileTypeXLS), int(FileTypeXLSX), int(FileTypeHTM), int(FileTypeHTML),
		},
		// ProcessMode with Interval = -1 means trigger on file load
		ProcessMode: &ProcessMode{
			Interval: -1, // -1 means trigger on file load
			Offset:   0,
		},
		// Complete document processing pipeline
		Workflow: &CatalogWorkflow{
			Nodes: []CatalogWorkflowNode{
				{
					ID:             "RootNode_1",
					Type:           "RootNode",
					InitParameters: map[string]map[string]interface{}{},
				},
				{
					ID:             "DocumentParseNode_2",
					Type:           "DocumentParseNode",
					InitParameters: map[string]map[string]interface{}{},
				},
				{
					ID:   "ChunkNode_4",
					Type: "ChunkNode",
					InitParameters: map[string]map[string]interface{}{
						"DocumentSplitter": {
							"enable_level_based_split": true,
						},
						"DocumentSplitter-ImageOCR": {
							"enable_level_based_split": true,
						},
						"DocumentSplitter-ImageCaption": {
							"enable_level_based_split": true,
						},
					},
				},
				{
					ID:             "EmbedNode_5",
					Type:           "EmbedNode",
					InitParameters: map[string]map[string]interface{}{},
				},
				{
					ID:             "WriteNode_6",
					Type:           "WriteNode",
					InitParameters: map[string]map[string]interface{}{},
				},
			},
			Connections: []CatalogWorkflowConnection{
				{
					Sender:   "RootNode_1",
					Receiver: "DocumentParseNode_2",
				},
				{
					Sender:   "DocumentParseNode_2",
					Receiver: "ChunkNode_4",
				},
				{
					Sender:   "ChunkNode_4",
					Receiver: "EmbedNode_5",
				},
				{
					Sender:   "EmbedNode_5",
					Receiver: "WriteNode_6",
				},
			},
		},
	}

	resp, err := c.raw.CreateWorkflow(ctx, req, opts...)
	if err != nil {
		return "", fmt.Errorf("failed to create workflow: %w", err)
	}

	if resp == nil || resp.ID == "" {
		return "", fmt.Errorf("workflow created but ID is empty")
	}

	return resp.ID, nil
}

// GetWorkflowJob retrieves a single workflow job by workflow ID and source file ID.
//
// This is a high-level convenience method that queries workflow jobs using ListWorkflowJobs
// with filters for workflow ID and source file ID, then returns the first matching job.
//
// Parameters:
//   - workflowID: the workflow ID (required)
//   - sourceFileID: the source file ID (required)
//
// Returns:
//   - *WorkflowJob: the matching workflow job, or nil if not found
//   - error: any error that occurred, including when no job is found
//
// Example:
//
//	job, err := sdkClient.GetWorkflowJob(ctx, "workflow-123", "file-456")
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Job ID: %s, Status: %s\n", job.JobID, job.Status)
func (c *SDKClient) GetWorkflowJob(ctx context.Context, workflowID string, sourceFileID string, opts ...CallOption) (*WorkflowJob, error) {
	if strings.TrimSpace(workflowID) == "" {
		return nil, fmt.Errorf("workflow_id is required")
	}
	if strings.TrimSpace(sourceFileID) == "" {
		return nil, fmt.Errorf("source_file_id is required")
	}

	// Query jobs with both filters
	resp, err := c.raw.ListWorkflowJobs(ctx, &WorkflowJobListRequest{
		WorkflowID:   workflowID,
		SourceFileID: sourceFileID,
		Page:         1,
		PageSize:     1, // We only need one result
	}, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to list workflow jobs: %w", err)
	}

	if resp == nil || len(resp.Jobs) == 0 {
		return nil, fmt.Errorf("workflow job not found for workflow_id=%s, source_file_id=%s", workflowID, sourceFileID)
	}

	// Return the first matching job
	return &resp.Jobs[0], nil
}

// WaitForWorkflowJob polls for a workflow job until it is found or the context times out.
//
// This method continuously queries for a workflow job matching the given workflow ID and source file ID
// until either:
//   - The job is found (returns the job immediately)
//   - The context is cancelled or times out (returns an error)
//
// The polling interval and timeout are controlled by the provided context. If the context has a deadline,
// the method will respect it. If no deadline is set, a default timeout of 60 seconds will be used.
//
// Parameters:
//   - ctx: context with optional timeout/deadline. If no deadline is set, defaults to 60 seconds.
//   - workflowID: the workflow ID (required)
//   - sourceFileID: the source file ID (required)
//   - pollInterval: the interval between polling attempts (default: 2 seconds if <= 0)
//
// Returns:
//   - *WorkflowJob: the matching workflow job
//   - error: any error that occurred, including context timeout or job not found
//
// Example:
//
//	// With explicit timeout
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//	job, err := sdkClient.WaitForWorkflowJob(ctx, "workflow-123", "file-456", 2*time.Second)
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Job found: %s, Status: %d\n", job.JobID, job.Status)
func (c *SDKClient) WaitForWorkflowJob(ctx context.Context, workflowID string, sourceFileID string, pollInterval time.Duration) (*WorkflowJob, error) {
	if strings.TrimSpace(workflowID) == "" {
		return nil, fmt.Errorf("workflow_id is required")
	}
	if strings.TrimSpace(sourceFileID) == "" {
		return nil, fmt.Errorf("source_file_id is required")
	}

	// Set default poll interval if not provided
	if pollInterval <= 0 {
		pollInterval = 2 * time.Second
	}

	// Ensure context has a deadline (default: 60 seconds)
	ctxWithDeadline := ctx
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctxWithDeadline, cancel = context.WithTimeout(ctx, 60*time.Second)
		defer cancel()
	}

	// Poll for the job
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	// Try once immediately
	job, err := c.GetWorkflowJob(ctxWithDeadline, workflowID, sourceFileID)
	if err == nil && job != nil {
		return job, nil
	}

	// Poll until found or context expires
	for {
		select {
		case <-ctxWithDeadline.Done():
			// Context expired or cancelled
			if ctxWithDeadline.Err() == context.DeadlineExceeded {
				return nil, fmt.Errorf("workflow job not found within timeout for workflow_id=%s, source_file_id=%s: %w", workflowID, sourceFileID, ctxWithDeadline.Err())
			}
			return nil, fmt.Errorf("context cancelled while waiting for workflow job: %w", ctxWithDeadline.Err())
		case <-ticker.C:
			// Poll again
			job, err := c.GetWorkflowJob(ctxWithDeadline, workflowID, sourceFileID)
			if err == nil && job != nil {
				return job, nil
			}
			// Continue polling on error (job not found yet)
		}
	}
}

// FindFilesByName searches for files by name within a specific volume.
//
// This is a high-level convenience method that uses ListFiles with filters
// to find files matching the given file name in the specified volume.
// The search is performed in the root directory (parent_id is empty).
//
// Parameters:
//   - ctx: context for the request
//   - fileName: the file name to search for (required)
//   - volumeID: the volume ID to search within (required)
//   - opts: optional call options
//
// Returns:
//   - *FileListResponse: the response containing matching files
//   - error: any error that occurred
//
// Example:
//
//	resp, err := sdkClient.FindFilesByName(ctx, "许继电气：关于召开2", "019b39fc-f4ee-7915-b701-66ae6a48d9fc")
//	if err != nil {
//		return err
//	}
//	for _, file := range resp.List {
//		fmt.Printf("Found file: %s (ID: %s)\n", file.Name, file.ID)
//	}
func (c *SDKClient) FindFilesByName(ctx context.Context, fileName string, volumeID VolumeID, opts ...CallOption) (*FileListResponse, error) {
	if strings.TrimSpace(fileName) == "" {
		return nil, fmt.Errorf("file_name is required")
	}
	if volumeID == "" {
		return nil, fmt.Errorf("volume_id is required")
	}

	// Build the request with filters matching the provided JSON example
	req := &FileListRequest{
		CommonCondition: CommonCondition{
			Page:     1,
			PageSize: 10,
			Order:    "desc",
			OrderBy:  "",
			Filters: []CommonFilter{
				{
					Name:   "volume_id",
					Values: []string{string(volumeID)},
					Fuzzy:  false,
				},
				{
					Name:   "parent_id",
					Values: []string{""},
					Fuzzy:  false,
				},
				{
					Name:   "file_name",
					Values: []string{fileName},
					Fuzzy:  false,
				},
			},
		},
		Keyword: "",
	}

	// Call the raw client's ListFiles method
	return c.raw.ListFiles(ctx, req, opts...)
}
