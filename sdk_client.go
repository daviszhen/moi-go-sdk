package sdk

import (
	"context"
	"errors"
	"fmt"
	"strings"
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
