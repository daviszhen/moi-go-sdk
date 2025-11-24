package sdk

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileLiveFlow(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	catalogID, markCatalogDeleted := createTestCatalog(t, client)
	databaseID, markDatabaseDeleted := createTestDatabase(t, client, catalogID)
	volumeID, markVolumeDeleted := createTestVolume(t, client, databaseID)

	folderResp, err := client.CreateFolder(ctx, &FolderCreateRequest{
		Name:     randomName("sdk-folder-"),
		VolumeID: volumeID,
	})
	require.NoError(t, err)
	folderID := folderResp.FolderID
	folderDeleted := false
	t.Cleanup(func() {
		if folderDeleted {
			return
		}
		if _, err := client.DeleteFolder(ctx, &FolderDeleteRequest{FolderID: folderID}); err != nil {
			t.Logf("cleanup delete folder failed: %v", err)
		}
	})

	createResp, err := client.CreateFile(ctx, &FileCreateRequest{
		Name:     randomName("sdk-file-"),
		VolumeID: volumeID,
		ParentID: folderID,
		ShowType: "normal",
		Size:     1,
		SavePath: "/tmp",
	})
	require.NoError(t, err)
	fileID := createResp.FileID

	_, err = client.UpdateFile(ctx, &FileUpdateRequest{
		FileID: fileID,
		Name:   randomName("sdk-file-updated-"),
	})
	require.NoError(t, err)

	infoResp, err := client.GetFile(ctx, &FileInfoRequest{FileID: fileID})
	require.NoError(t, err)
	require.Equal(t, fileID, infoResp.ID)

	listReq := &FileListRequest{}
	listReq.Filters = append(listReq.Filters, CommonFilter{
		Name:   "volume_id",
		Values: []string{string(volumeID)},
	})
	listResp, err := client.ListFiles(ctx, listReq)
	require.NoError(t, err)
	require.NotNil(t, listResp)

	_, err = client.DeleteFile(ctx, &FileDeleteRequest{FileID: fileID})
	require.NoError(t, err)

	folderDeleted = true
	_, err = client.DeleteFolder(ctx, &FolderDeleteRequest{FolderID: folderID})
	require.NoError(t, err)

	_, err = client.DeleteVolume(ctx, &VolumeDeleteRequest{VolumeID: volumeID})
	require.NoError(t, err)
	markVolumeDeleted()

	_, err = client.DeleteDatabase(ctx, &DatabaseDeleteRequest{DatabaseID: databaseID})
	require.NoError(t, err)
	markDatabaseDeleted()

	_, err = client.DeleteCatalog(ctx, &CatalogDeleteRequest{CatalogID: catalogID})
	require.NoError(t, err)
	markCatalogDeleted()
}

func TestFileNilRequestErrors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := &RawClient{}

	tests := []struct {
		name string
		call func() error
	}{
		{"Create", func() error { _, err := client.CreateFile(ctx, nil); return err }},
		{"Update", func() error { _, err := client.UpdateFile(ctx, nil); return err }},
		{"Delete", func() error { _, err := client.DeleteFile(ctx, nil); return err }},
		{"DeleteRef", func() error { _, err := client.DeleteFileRef(ctx, nil); return err }},
		{"Info", func() error { _, err := client.GetFile(ctx, nil); return err }},
		{"List", func() error { _, err := client.ListFiles(ctx, nil); return err }},
		{"Upload", func() error { _, err := client.UploadFile(ctx, nil); return err }},
		{"Download", func() error { _, err := client.GetFileDownloadLink(ctx, nil); return err }},
		{"PreviewLink", func() error { _, err := client.GetFilePreviewLink(ctx, nil); return err }},
		{"PreviewStream", func() error { _, err := client.GetFilePreviewStream(ctx, nil); return err }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.ErrorIs(t, tc.call(), ErrNilRequest)
		})
	}
}

func TestFileVolumeIDNotExists(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	nonExistentVolumeID := VolumeID("aaaaaa")

	// Try to create file with non-existent volume ID
	_, err := client.CreateFile(ctx, &FileCreateRequest{
		Name:     "test_file",
		VolumeID: nonExistentVolumeID,
		ParentID: "",
		Size:     10,
		ShowType: "normal",
	})
	require.Error(t, err)
	t.Logf("Expected error for non-existent volume ID: %v", err)
}

func TestFileParentIDNotExists(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	catalogID, markCatalogDeleted := createTestCatalog(t, client)
	databaseID, markDatabaseDeleted := createTestDatabase(t, client, catalogID)
	volumeID, markVolumeDeleted := createTestVolume(t, client, databaseID)

	defer func() {
		markVolumeDeleted()
		markDatabaseDeleted()
		markCatalogDeleted()
	}()

	nonExistentParentID := FileID("11111")

	// Try to create file with non-existent parent ID
	_, err := client.CreateFile(ctx, &FileCreateRequest{
		Name:     "test_file",
		VolumeID: volumeID,
		ParentID: nonExistentParentID,
		Size:     10,
		ShowType: "normal",
	})
	require.Error(t, err)
	t.Logf("Expected error for non-existent parent ID: %v", err)
}

func TestFileDuplicateNameHandling(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	catalogID, markCatalogDeleted := createTestCatalog(t, client)
	databaseID, markDatabaseDeleted := createTestDatabase(t, client, catalogID)
	volumeID, markVolumeDeleted := createTestVolume(t, client, databaseID)

	defer func() {
		markVolumeDeleted()
		markDatabaseDeleted()
		markCatalogDeleted()
	}()

	fileName := "file1.txt"

	// Create first file
	createResp1, err := client.CreateFile(ctx, &FileCreateRequest{
		Name:     fileName,
		VolumeID: volumeID,
		ParentID: "",
		Size:     10,
		ShowType: "normal",
	})
	require.NoError(t, err)
	require.NotZero(t, createResp1.FileID)

	// Cleanup
	defer func() {
		if _, err := client.DeleteFile(ctx, &FileDeleteRequest{FileID: createResp1.FileID}); err != nil {
			t.Logf("cleanup delete file 1 failed: %v", err)
		}
	}()

	// Create second file with same name - should be renamed
	createResp2, err := client.CreateFile(ctx, &FileCreateRequest{
		Name:     fileName,
		VolumeID: volumeID,
		ParentID: "",
		Size:     10,
		ShowType: "normal",
	})
	require.NoError(t, err)
	require.NotZero(t, createResp2.FileID)
	require.Contains(t, createResp2.Name, "file1")
	t.Logf("Second file name: %s", createResp2.Name)

	// Cleanup
	defer func() {
		if _, err := client.DeleteFile(ctx, &FileDeleteRequest{FileID: createResp2.FileID}); err != nil {
			t.Logf("cleanup delete file 2 failed: %v", err)
		}
	}()

	// Create third file with same name - should be renamed again
	createResp3, err := client.CreateFile(ctx, &FileCreateRequest{
		Name:     fileName,
		VolumeID: volumeID,
		ParentID: "",
		Size:     10,
		ShowType: "normal",
	})
	require.NoError(t, err)
	require.NotZero(t, createResp3.FileID)
	require.Contains(t, createResp3.Name, "file1")
	t.Logf("Third file name: %s", createResp3.Name)

	// Cleanup
	defer func() {
		if _, err := client.DeleteFile(ctx, &FileDeleteRequest{FileID: createResp3.FileID}); err != nil {
			t.Logf("cleanup delete file 3 failed: %v", err)
		}
	}()
}

func TestFileWithRefFileID(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	catalogID, markCatalogDeleted := createTestCatalog(t, client)
	databaseID, markDatabaseDeleted := createTestDatabase(t, client, catalogID)
	volumeID, markVolumeDeleted := createTestVolume(t, client, databaseID)

	defer func() {
		markVolumeDeleted()
		markDatabaseDeleted()
		markCatalogDeleted()
	}()

	refFileID := randomName("ref_test_")
	fileName := randomName("file1-")

	// Create file with RefFileID
	createResp, err := client.CreateFile(ctx, &FileCreateRequest{
		Name:      fileName + ".txt",
		VolumeID:  volumeID,
		ParentID:  "",
		Size:      10,
		ShowType:  "normal",
		RefFileID: refFileID,
	})
	require.NoError(t, err)
	require.NotZero(t, createResp.FileID)
	t.Logf("Created file with RefFileID: %s, FileID: %s", refFileID, createResp.FileID)

	// Cleanup
	defer func() {
		if _, err := client.DeleteFileRef(ctx, &FileDeleteRefRequest{RefFileID: refFileID}); err != nil {
			t.Logf("cleanup delete file by ref failed: %v", err)
		}
	}()

	// List files with ref_file_id filter (before deletion)
	listReq := &FileListRequest{
		CommonCondition: CommonCondition{
			Page:     1,
			PageSize: 10,
			Filters: []CommonFilter{
				{
					Name:   "volume_id",
					Values: []string{string(volumeID)},
					Fuzzy:  false,
				},
				{
					Name:   "ref_file_id",
					Values: []string{refFileID},
					Fuzzy:  false,
				},
			},
		},
	}
	listResp, err := client.ListFiles(ctx, listReq)
	require.NoError(t, err)
	require.NotNil(t, listResp)
	require.GreaterOrEqual(t, listResp.Total, 1, "should find at least one file with ref_file_id")
	require.GreaterOrEqual(t, len(listResp.List), 1, "list should contain at least one file")
	if len(listResp.List) > 0 {
		require.Equal(t, refFileID, listResp.List[0].RefFileID)
	}

	// Delete file by RefFileID
	deleteRefResp, err := client.DeleteFileRef(ctx, &FileDeleteRefRequest{RefFileID: refFileID})
	require.NoError(t, err)
	require.Equal(t, createResp.FileID, deleteRefResp.FileID)

	// Verify file is deleted
	_, err = client.GetFile(ctx, &FileInfoRequest{FileID: createResp.FileID})
	require.Error(t, err, "file should be deleted after DeleteFileRef")
	t.Logf("File deleted successfully: %v", err)
}

func TestFileUpdateNameExists(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	catalogID, markCatalogDeleted := createTestCatalog(t, client)
	databaseID, markDatabaseDeleted := createTestDatabase(t, client, catalogID)
	volumeID, markVolumeDeleted := createTestVolume(t, client, databaseID)

	defer func() {
		markVolumeDeleted()
		markDatabaseDeleted()
		markCatalogDeleted()
	}()

	// Create first file
	createResp1, err := client.CreateFile(ctx, &FileCreateRequest{
		Name:     "file1",
		VolumeID: volumeID,
		ParentID: "",
		Size:     10,
		ShowType: "normal",
	})
	require.NoError(t, err)
	require.NotZero(t, createResp1.FileID)

	// Create second file
	createResp2, err := client.CreateFile(ctx, &FileCreateRequest{
		Name:     "file2",
		VolumeID: volumeID,
		ParentID: "",
		Size:     10,
		ShowType: "normal",
	})
	require.NoError(t, err)
	require.NotZero(t, createResp2.FileID)

	// Cleanup
	defer func() {
		if _, err := client.DeleteFile(ctx, &FileDeleteRequest{FileID: createResp1.FileID}); err != nil {
			t.Logf("cleanup delete file 1 failed: %v", err)
		}
		if _, err := client.DeleteFile(ctx, &FileDeleteRequest{FileID: createResp2.FileID}); err != nil {
			t.Logf("cleanup delete file 2 failed: %v", err)
		}
	}()

	// Try to update first file with the name of second file
	_, err = client.UpdateFile(ctx, &FileUpdateRequest{
		FileID: createResp1.FileID,
		Name:   "file2",
	})
	require.Error(t, err)
	t.Logf("Expected error for duplicate name in update: %v", err)
}

func TestFileListWithFilters(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	catalogID, markCatalogDeleted := createTestCatalog(t, client)
	databaseID, markDatabaseDeleted := createTestDatabase(t, client, catalogID)
	volumeID, markVolumeDeleted := createTestVolume(t, client, databaseID)

	defer func() {
		markVolumeDeleted()
		markDatabaseDeleted()
		markCatalogDeleted()
	}()

	// Create a folder
	folderResp, err := client.CreateFolder(ctx, &FolderCreateRequest{
		Name:     randomName("sdk-folder-"),
		VolumeID: volumeID,
	})
	require.NoError(t, err)
	folderID := folderResp.FolderID

	// Create files in root and in folder
	file1, err := client.CreateFile(ctx, &FileCreateRequest{
		Name:     "root_file.txt",
		VolumeID: volumeID,
		ParentID: "",
		Size:     10,
		ShowType: "normal",
	})
	require.NoError(t, err)

	file2, err := client.CreateFile(ctx, &FileCreateRequest{
		Name:     "folder_file.txt",
		VolumeID: volumeID,
		ParentID: folderID,
		Size:     10,
		ShowType: "normal",
	})
	require.NoError(t, err)

	// Cleanup
	defer func() {
		if _, err := client.DeleteFile(ctx, &FileDeleteRequest{FileID: file1.FileID}); err != nil {
			t.Logf("cleanup delete file 1 failed: %v", err)
		}
		if _, err := client.DeleteFile(ctx, &FileDeleteRequest{FileID: file2.FileID}); err != nil {
			t.Logf("cleanup delete file 2 failed: %v", err)
		}
		if _, err := client.DeleteFolder(ctx, &FolderDeleteRequest{FolderID: folderID}); err != nil {
			t.Logf("cleanup delete folder failed: %v", err)
		}
	}()

	// List files in root (parent_id is empty)
	listReq := &FileListRequest{
		CommonCondition: CommonCondition{
			Page:     1,
			PageSize: 10,
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
			},
		},
	}
	listResp, err := client.ListFiles(ctx, listReq)
	require.NoError(t, err)
	require.NotNil(t, listResp)
	require.GreaterOrEqual(t, listResp.Total, 1)

	// List files in folder
	listReq2 := &FileListRequest{
		CommonCondition: CommonCondition{
			Page:     1,
			PageSize: 10,
			Filters: []CommonFilter{
				{
					Name:   "volume_id",
					Values: []string{string(volumeID)},
					Fuzzy:  false,
				},
				{
					Name:   "parent_id",
					Values: []string{string(folderID)},
					Fuzzy:  false,
				},
			},
		},
	}
	listResp2, err := client.ListFiles(ctx, listReq2)
	require.NoError(t, err)
	require.NotNil(t, listResp2)
	require.GreaterOrEqual(t, listResp2.Total, 1)
}

func TestFileDownloadAndPreview(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	catalogID, markCatalogDeleted := createTestCatalog(t, client)
	databaseID, markDatabaseDeleted := createTestDatabase(t, client, catalogID)
	volumeID, markVolumeDeleted := createTestVolume(t, client, databaseID)

	defer func() {
		markVolumeDeleted()
		markDatabaseDeleted()
		markCatalogDeleted()
	}()

	// Create a file
	createResp, err := client.CreateFile(ctx, &FileCreateRequest{
		Name:     "test_file.txt",
		VolumeID: volumeID,
		ParentID: "",
		Size:     10,
		ShowType: "normal",
	})
	require.NoError(t, err)
	require.NotZero(t, createResp.FileID)

	// Cleanup
	defer func() {
		if _, err := client.DeleteFile(ctx, &FileDeleteRequest{FileID: createResp.FileID}); err != nil {
			t.Logf("cleanup delete file failed: %v", err)
		}
	}()

	// Get download link
	downloadResp, err := client.GetFileDownloadLink(ctx, &FileDownloadRequest{FileID: createResp.FileID})
	require.NoError(t, err)
	require.NotEmpty(t, downloadResp.Url)
	t.Logf("Download URL: %s", downloadResp.Url)

	// Get preview link
	previewLinkResp, err := client.GetFilePreviewLink(ctx, &FilePreviewLinkRequest{FileID: createResp.FileID})
	require.NoError(t, err)
	require.NotEmpty(t, previewLinkResp.Url)
	t.Logf("Preview Link URL: %s", previewLinkResp.Url)

	// Get preview stream
	previewStreamResp, err := client.GetFilePreviewStream(ctx, &FilePreviewStreamRequest{FileID: createResp.FileID})
	require.NoError(t, err)
	require.NotEmpty(t, previewStreamResp.Url)
	t.Logf("Preview Stream URL: %s", previewStreamResp.Url)
}

func TestFileIDNotExists(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	nonExistentID := FileID("aaaaaa")

	// Try to get non-existent file
	_, err := client.GetFile(ctx, &FileInfoRequest{FileID: nonExistentID})
	require.Error(t, err)
	t.Logf("Expected error for non-existent file ID: %v", err)

	// Try to update non-existent file
	_, err = client.UpdateFile(ctx, &FileUpdateRequest{
		FileID: nonExistentID,
		Name:   "test",
	})
	require.Error(t, err)
	t.Logf("Expected error for updating non-existent file: %v", err)

	// Try to delete non-existent file
	_, err = client.DeleteFile(ctx, &FileDeleteRequest{FileID: nonExistentID})
	require.Error(t, err)
	t.Logf("Expected error for deleting non-existent file: %v", err)
}

func TestFileUpdateAndVerify(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	catalogID, markCatalogDeleted := createTestCatalog(t, client)
	databaseID, markDatabaseDeleted := createTestDatabase(t, client, catalogID)
	volumeID, markVolumeDeleted := createTestVolume(t, client, databaseID)

	defer func() {
		markVolumeDeleted()
		markDatabaseDeleted()
		markCatalogDeleted()
	}()

	// Create a file
	createResp, err := client.CreateFile(ctx, &FileCreateRequest{
		Name:     "file1",
		VolumeID: volumeID,
		ParentID: "",
		Size:     10,
		ShowType: "normal",
	})
	require.NoError(t, err)
	require.NotZero(t, createResp.FileID)

	// Cleanup
	defer func() {
		if _, err := client.DeleteFile(ctx, &FileDeleteRequest{FileID: createResp.FileID}); err != nil {
			t.Logf("cleanup delete file failed: %v", err)
		}
	}()

	// Update file name
	updatedName := "file1_new"
	_, err = client.UpdateFile(ctx, &FileUpdateRequest{
		FileID: createResp.FileID,
		Name:   updatedName,
	})
	require.NoError(t, err)

	// Verify the update
	infoResp, err := client.GetFile(ctx, &FileInfoRequest{FileID: createResp.FileID})
	require.NoError(t, err)
	require.Equal(t, updatedName, infoResp.Name)
}

func TestFileListWithOrdering(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	catalogID, markCatalogDeleted := createTestCatalog(t, client)
	databaseID, markDatabaseDeleted := createTestDatabase(t, client, catalogID)
	volumeID, markVolumeDeleted := createTestVolume(t, client, databaseID)

	defer func() {
		markVolumeDeleted()
		markDatabaseDeleted()
		markCatalogDeleted()
	}()

	// Create multiple files
	file1, err := client.CreateFile(ctx, &FileCreateRequest{
		Name:     "file_a.txt",
		VolumeID: volumeID,
		ParentID: "",
		Size:     10,
		ShowType: "normal",
	})
	require.NoError(t, err)

	file2, err := client.CreateFile(ctx, &FileCreateRequest{
		Name:     "file_b.txt",
		VolumeID: volumeID,
		ParentID: "",
		Size:     10,
		ShowType: "normal",
	})
	require.NoError(t, err)

	file3, err := client.CreateFile(ctx, &FileCreateRequest{
		Name:     "file_c.txt",
		VolumeID: volumeID,
		ParentID: "",
		Size:     10,
		ShowType: "normal",
	})
	require.NoError(t, err)

	// Cleanup
	defer func() {
		if _, err := client.DeleteFile(ctx, &FileDeleteRequest{FileID: file1.FileID}); err != nil {
			t.Logf("cleanup delete file 1 failed: %v", err)
		}
		if _, err := client.DeleteFile(ctx, &FileDeleteRequest{FileID: file2.FileID}); err != nil {
			t.Logf("cleanup delete file 2 failed: %v", err)
		}
		if _, err := client.DeleteFile(ctx, &FileDeleteRequest{FileID: file3.FileID}); err != nil {
			t.Logf("cleanup delete file 3 failed: %v", err)
		}
	}()

	// List files with ordering
	listReq := &FileListRequest{
		CommonCondition: CommonCondition{
			Page:     1,
			PageSize: 10,
			Order:    "DESC",
			OrderBy:  "file_type",
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
			},
		},
	}
	listResp, err := client.ListFiles(ctx, listReq)
	require.NoError(t, err)
	require.NotNil(t, listResp)
	require.GreaterOrEqual(t, listResp.Total, 3)
}

func TestFileDeleteAndRecreate(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	catalogID, markCatalogDeleted := createTestCatalog(t, client)
	databaseID, markDatabaseDeleted := createTestDatabase(t, client, catalogID)
	volumeID, markVolumeDeleted := createTestVolume(t, client, databaseID)

	defer func() {
		markVolumeDeleted()
		markDatabaseDeleted()
		markCatalogDeleted()
	}()

	fileName := "file1.txt"

	// Create first file
	createResp1, err := client.CreateFile(ctx, &FileCreateRequest{
		Name:      fileName,
		VolumeID:  volumeID,
		ParentID:  "",
		Size:      10,
		ShowType:  "normal",
		RefFileID: "ref_test_123",
	})
	require.NoError(t, err)
	require.NotZero(t, createResp1.FileID)
	require.Equal(t, "file1", createResp1.Name)

	// Delete the file
	_, err = client.DeleteFile(ctx, &FileDeleteRequest{FileID: createResp1.FileID})
	require.NoError(t, err)

	// Recreate file with same name and RefFileID - should not be renamed
	createResp2, err := client.CreateFile(ctx, &FileCreateRequest{
		Name:      fileName,
		VolumeID:  volumeID,
		ParentID:  "",
		Size:      10,
		ShowType:  "normal",
		RefFileID: "ref_test_123",
	})
	require.NoError(t, err)
	require.NotZero(t, createResp2.FileID)
	require.Equal(t, "file1", createResp2.Name, "should not be renamed when RefFileID is provided")

	// Cleanup
	defer func() {
		if _, err := client.DeleteFileRef(ctx, &FileDeleteRefRequest{RefFileID: "ref_test_123"}); err != nil {
			t.Logf("cleanup delete file by ref failed: %v", err)
		}
	}()
}

func TestFileListWithRefFileIDFilter(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	catalogID, markCatalogDeleted := createTestCatalog(t, client)
	databaseID, markDatabaseDeleted := createTestDatabase(t, client, catalogID)
	volumeID, markVolumeDeleted := createTestVolume(t, client, databaseID)

	defer func() {
		markVolumeDeleted()
		markDatabaseDeleted()
		markCatalogDeleted()
	}()

	refFileID := "ref_test_123"

	// Create file with RefFileID
	createResp, err := client.CreateFile(ctx, &FileCreateRequest{
		Name:      "file1.txt",
		VolumeID:  volumeID,
		ParentID:  "",
		Size:      10,
		ShowType:  "normal",
		RefFileID: refFileID,
	})
	require.NoError(t, err)
	require.NotZero(t, createResp.FileID)

	// Cleanup
	defer func() {
		if _, err := client.DeleteFileRef(ctx, &FileDeleteRefRequest{RefFileID: refFileID}); err != nil {
			t.Logf("cleanup delete file by ref failed: %v", err)
		}
	}()

	// List files with ref_file_id filter
	listReq := &FileListRequest{
		CommonCondition: CommonCondition{
			Page:     1,
			PageSize: 10,
			Filters: []CommonFilter{
				{
					Name:   "volume_id",
					Values: []string{string(volumeID)},
					Fuzzy:  false,
				},
				{
					Name:   "ref_file_id",
					Values: []string{refFileID},
					Fuzzy:  false,
				},
			},
		},
	}
	listResp, err := client.ListFiles(ctx, listReq)
	require.NoError(t, err)
	require.NotNil(t, listResp)
	require.Equal(t, 1, listResp.Total)
	require.Equal(t, 1, len(listResp.List))
	require.Equal(t, "file1", listResp.List[0].Name)
	require.Equal(t, refFileID, listResp.List[0].RefFileID)
}

func TestFileDownloadLinkFormat(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	catalogID, markCatalogDeleted := createTestCatalog(t, client)
	databaseID, markDatabaseDeleted := createTestDatabase(t, client, catalogID)
	volumeID, markVolumeDeleted := createTestVolume(t, client, databaseID)

	defer func() {
		markVolumeDeleted()
		markDatabaseDeleted()
		markCatalogDeleted()
	}()

	// Create a file
	createResp, err := client.CreateFile(ctx, &FileCreateRequest{
		Name:     "test_file.txt",
		VolumeID: volumeID,
		ParentID: "",
		Size:     10,
		ShowType: "normal",
	})
	require.NoError(t, err)
	require.NotZero(t, createResp.FileID)

	// Cleanup
	defer func() {
		if _, err := client.DeleteFile(ctx, &FileDeleteRequest{FileID: createResp.FileID}); err != nil {
			t.Logf("cleanup delete file failed: %v", err)
		}
	}()

	// Get download link
	downloadResp, err := client.GetFileDownloadLink(ctx, &FileDownloadRequest{FileID: createResp.FileID})
	require.NoError(t, err)
	require.NotEmpty(t, downloadResp.Url)
	// URL should contain signature and expires parameters
	require.Contains(t, downloadResp.Url, "Expires=")
	require.Contains(t, downloadResp.Url, "Signature=")
	t.Logf("Download URL format verified: %s", downloadResp.Url)
}

func TestFilePreviewLinkFormat(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	catalogID, markCatalogDeleted := createTestCatalog(t, client)
	databaseID, markDatabaseDeleted := createTestDatabase(t, client, catalogID)
	volumeID, markVolumeDeleted := createTestVolume(t, client, databaseID)

	defer func() {
		markVolumeDeleted()
		markDatabaseDeleted()
		markCatalogDeleted()
	}()

	// Create a file
	createResp, err := client.CreateFile(ctx, &FileCreateRequest{
		Name:     "test_file.txt",
		VolumeID: volumeID,
		ParentID: "",
		Size:     10,
		ShowType: "normal",
	})
	require.NoError(t, err)
	require.NotZero(t, createResp.FileID)

	// Cleanup
	defer func() {
		if _, err := client.DeleteFile(ctx, &FileDeleteRequest{FileID: createResp.FileID}); err != nil {
			t.Logf("cleanup delete file failed: %v", err)
		}
	}()

	// Get preview link
	previewLinkResp, err := client.GetFilePreviewLink(ctx, &FilePreviewLinkRequest{FileID: createResp.FileID})
	require.NoError(t, err)
	require.NotEmpty(t, previewLinkResp.Url)
	// URL should contain signature and expires parameters
	require.Contains(t, previewLinkResp.Url, "Expires=")
	require.Contains(t, previewLinkResp.Url, "Signature=")
	t.Logf("Preview Link URL format verified: %s", previewLinkResp.Url)
}

func TestFilePreviewStreamFormat(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	catalogID, markCatalogDeleted := createTestCatalog(t, client)
	databaseID, markDatabaseDeleted := createTestDatabase(t, client, catalogID)
	volumeID, markVolumeDeleted := createTestVolume(t, client, databaseID)

	defer func() {
		markVolumeDeleted()
		markDatabaseDeleted()
		markCatalogDeleted()
	}()

	// Create a file
	createResp, err := client.CreateFile(ctx, &FileCreateRequest{
		Name:     "test_file.txt",
		VolumeID: volumeID,
		ParentID: "",
		Size:     10,
		ShowType: "normal",
	})
	require.NoError(t, err)
	require.NotZero(t, createResp.FileID)

	// Cleanup
	defer func() {
		if _, err := client.DeleteFile(ctx, &FileDeleteRequest{FileID: createResp.FileID}); err != nil {
			t.Logf("cleanup delete file failed: %v", err)
		}
	}()

	// Get preview stream
	previewStreamResp, err := client.GetFilePreviewStream(ctx, &FilePreviewStreamRequest{FileID: createResp.FileID})
	require.NoError(t, err)
	require.NotEmpty(t, previewStreamResp.Url)
	// URL should contain signature and expires parameters
	require.Contains(t, previewStreamResp.Url, "Expires=")
	require.Contains(t, previewStreamResp.Url, "Signature=")
	t.Logf("Preview Stream URL format verified: %s", previewStreamResp.Url)
}
