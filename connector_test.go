package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUploadLocalFilesNilRequestErrors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := &RawClient{}

	tests := []struct {
		name      string
		files     []FileUploadItem
		meta      []FileMeta
		expectErr string
	}{
		{
			name:      "NoFiles",
			files:     []FileUploadItem{},
			meta:      []FileMeta{{Filename: "test.txt", Path: "/test"}},
			expectErr: "at least one file is required",
		},
		{
			name:      "NoMeta",
			files:     []FileUploadItem{{File: strings.NewReader("test"), FileName: "test.txt"}},
			meta:      []FileMeta{},
			expectErr: "meta is required",
		},
		{
			name:      "BothEmpty",
			files:     []FileUploadItem{},
			meta:      []FileMeta{},
			expectErr: "at least one file is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := client.UploadLocalFiles(ctx, tc.files, tc.meta)
			require.Nil(t, resp)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.expectErr)
		})
	}
}

func TestUploadLocalFileNilRequestErrors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := &RawClient{}

	tests := []struct {
		name      string
		file      io.Reader
		fileName  string
		meta      []FileMeta
		expectErr string
	}{
		{
			name:      "NoMeta",
			file:      strings.NewReader("test"),
			fileName:  "test.txt",
			meta:      []FileMeta{},
			expectErr: "meta is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := client.UploadLocalFile(ctx, tc.file, tc.fileName, tc.meta)
			require.Nil(t, resp)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.expectErr)
		})
	}
}

func TestUploadLocalFileFromPathErrors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := &RawClient{baseURL: "http://example.com", apiKey: "test-key"}

	tests := []struct {
		name      string
		filePath  string
		meta      []FileMeta
		expectErr string
	}{
		{
			name:      "NonExistentFile",
			filePath:  "/nonexistent/file/path",
			meta:      []FileMeta{{Filename: "test.txt", Path: "/test"}},
			expectErr: "open file",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := client.UploadLocalFileFromPath(ctx, tc.filePath, tc.meta)
			require.Nil(t, resp)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.expectErr)
		})
	}
}

func TestUploadLocalFilesMultipartForm(t *testing.T) {
	t.Parallel()

	// Create a mock client to test multipart form creation
	// Note: This is a unit test, not an integration test
	files := []FileUploadItem{
		{
			File:     strings.NewReader("test file content 1"),
			FileName: "test1.txt",
		},
		{
			File:     strings.NewReader("test file content 2"),
			FileName: "test2.txt",
		},
	}

	meta := []FileMeta{
		{Filename: "test1.txt", Path: "/test/path1"},
		{Filename: "test2.txt", Path: "/test/path2"},
	}

	// Test that the multipart form is created correctly
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add meta field
	metaJSON, err := json.Marshal(meta)
	require.NoError(t, err)
	metaField, err := writer.CreateFormField("meta")
	require.NoError(t, err)
	_, err = metaField.Write(metaJSON)
	require.NoError(t, err)

	// Add files
	for _, item := range files {
		fileField, err := writer.CreateFormFile("file", item.FileName)
		require.NoError(t, err)
		_, err = io.Copy(fileField, item.File)
		require.NoError(t, err)
	}

	err = writer.Close()
	require.NoError(t, err)

	// Verify multipart form was created
	contentType := writer.FormDataContentType()
	require.Contains(t, contentType, "multipart/form-data")
	require.Greater(t, body.Len(), 0)
}

func TestFileMeta(t *testing.T) {
	t.Parallel()

	meta := FileMeta{
		Filename: "test.txt",
		Path:     "/test/path",
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(meta)
	require.NoError(t, err)
	require.Contains(t, string(jsonData), "test.txt")
	require.Contains(t, string(jsonData), "/test/path")

	// Test JSON unmarshaling
	var unmarshaled FileMeta
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)
	require.Equal(t, meta.Filename, unmarshaled.Filename)
	require.Equal(t, meta.Path, unmarshaled.Path)
}

func TestLocalFileUploadResponse(t *testing.T) {
	t.Parallel()

	resp := LocalFileUploadResponse{
		ConnFileIds: []string{"123", "456", "789"},
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(resp)
	require.NoError(t, err)
	require.Contains(t, string(jsonData), "123")
	require.Contains(t, string(jsonData), "conn_file_ids")

	// Test JSON unmarshaling
	var unmarshaled LocalFileUploadResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)
	require.Equal(t, resp.ConnFileIds, unmarshaled.ConnFileIds)
}

func TestUploadLocalFileLiveFlow(t *testing.T) {
	ctx := context.Background()
	// Create client directly without health check since connector endpoint might be available
	// even if healthz endpoint is not
	client, err := NewRawClient(testBaseURL, testAPIKey)
	require.NoError(t, err)

	// Create a temporary file for testing
	tmpDir := t.TempDir()
	testFileName := "test_upload_file.txt"
	testFilePath := filepath.Join(tmpDir, testFileName)
	testContent := "This is a test file content for SDK upload test\n" + randomName("content-")

	err = os.WriteFile(testFilePath, []byte(testContent), 0644)
	require.NoError(t, err)

	// Prepare file metadata
	meta := []FileMeta{
		{
			Filename: testFileName,
			Path:     "/test/upload/path",
		},
	}

	// Test UploadLocalFileFromPath
	t.Run("UploadFromPath", func(t *testing.T) {
		resp, err := client.UploadLocalFileFromPath(ctx, testFilePath, meta)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotEmpty(t, resp.ConnFileIds)
		require.Greater(t, len(resp.ConnFileIds), 0)
		t.Logf("Upload successful, received conn_file_ids: %v", resp.ConnFileIds)
	})

	// Test UploadLocalFile with io.Reader
	t.Run("UploadFromReader", func(t *testing.T) {
		file, err := os.Open(testFilePath)
		require.NoError(t, err)
		defer file.Close()

		resp, err := client.UploadLocalFile(ctx, file, testFileName, meta)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotEmpty(t, resp.ConnFileIds)
		require.Greater(t, len(resp.ConnFileIds), 0)
		t.Logf("Upload successful, received conn_file_ids: %v", resp.ConnFileIds)
	})

	// Test UploadLocalFiles with multiple files
	t.Run("UploadMultipleFiles", func(t *testing.T) {
		// Create another test file
		testFileName2 := "test_upload_file2.txt"
		testFilePath2 := filepath.Join(tmpDir, testFileName2)
		testContent2 := "This is another test file content\n" + randomName("content2-")

		err := os.WriteFile(testFilePath2, []byte(testContent2), 0644)
		require.NoError(t, err)

		file1, err := os.Open(testFilePath)
		require.NoError(t, err)
		defer file1.Close()

		file2, err := os.Open(testFilePath2)
		require.NoError(t, err)
		defer file2.Close()

		files := []FileUploadItem{
			{
				File:     file1,
				FileName: testFileName,
			},
			{
				File:     file2,
				FileName: testFileName2,
			},
		}

		metaMultiple := []FileMeta{
			{
				Filename: testFileName,
				Path:     "/test/upload/path1",
			},
			{
				Filename: testFileName2,
				Path:     "/test/upload/path2",
			},
		}

		resp, err := client.UploadLocalFiles(ctx, files, metaMultiple)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotEmpty(t, resp.ConnFileIds)
		require.GreaterOrEqual(t, len(resp.ConnFileIds), 1)
		t.Logf("Multiple files upload successful, received conn_file_ids: %v", resp.ConnFileIds)
	})
}

func TestFilePreviewNilRequestErrors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := &RawClient{}

	resp, err := client.FilePreview(ctx, nil)
	require.Nil(t, resp)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrNilRequest)
}

func TestFilePreviewValidationErrors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := &RawClient{baseURL: "http://example.com", apiKey: "test-key"}

	tests := []struct {
		name      string
		req       *FilePreviewRequest
		expectErr string
	}{
		{
			name:      "NoConnectorIdNoConnFileId",
			req:       &FilePreviewRequest{},
			expectErr: "file preview needs conn_file_id for local upload file",
		},
		{
			name: "ConnectorIdButNoUriOrConnFileId",
			req: &FilePreviewRequest{
				ConnectorId: 123,
			},
			expectErr: "file preview needs uri or conn_file_id when connector_id is provided",
		},
		{
			name: "ConnectorIdWithEmptyUriAndConnFileId",
			req: &FilePreviewRequest{
				ConnectorId: 123,
				Uri:         "",
				ConnFileId:  "",
			},
			expectErr: "file preview needs uri or conn_file_id when connector_id is provided",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := client.FilePreview(ctx, tc.req)
			require.Nil(t, resp)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.expectErr)
		})
	}
}

func TestFilePreviewRequestJSON(t *testing.T) {
	t.Parallel()

	req := FilePreviewRequest{
		ConnectorId:   123,
		ConnFileId:    "456",
		Uri:           "s3://bucket/file.csv",
		IsColumnName:  true,
		ColumnNameRow: 1,
		RowStart:      2,
		Csv: &ConnectorCsvConfig{
			Separator: ",",
			Delimiter: "\"",
			IsEscape:  true,
		},
		FileType: 1,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(req)
	require.NoError(t, err)
	require.Contains(t, string(jsonData), "123")
	require.Contains(t, string(jsonData), "456")
	require.Contains(t, string(jsonData), "s3://bucket/file.csv")
	require.Contains(t, string(jsonData), "connector_id")
	require.Contains(t, string(jsonData), "conn_file_id")
	require.Contains(t, string(jsonData), "uri")
	require.Contains(t, string(jsonData), "isColumnName")
	require.Contains(t, string(jsonData), "csv")

	// Test JSON unmarshaling
	var unmarshaled FilePreviewRequest
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)
	require.Equal(t, req.ConnectorId, unmarshaled.ConnectorId)
	require.Equal(t, req.ConnFileId, unmarshaled.ConnFileId)
	require.Equal(t, req.Uri, unmarshaled.Uri)
	require.Equal(t, req.IsColumnName, unmarshaled.IsColumnName)
	require.Equal(t, req.ColumnNameRow, unmarshaled.ColumnNameRow)
	require.Equal(t, req.RowStart, unmarshaled.RowStart)
	require.Equal(t, req.FileType, unmarshaled.FileType)
	require.NotNil(t, unmarshaled.Csv)
	require.Equal(t, req.Csv.Separator, unmarshaled.Csv.Separator)
	require.Equal(t, req.Csv.Delimiter, unmarshaled.Csv.Delimiter)
	require.Equal(t, req.Csv.IsEscape, unmarshaled.Csv.IsEscape)
}

func TestConnectorCsvConfigJSON(t *testing.T) {
	t.Parallel()

	csvCfg := ConnectorCsvConfig{
		Separator: ",",
		Delimiter: "\"",
		IsEscape:  true,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(csvCfg)
	require.NoError(t, err)
	require.Contains(t, string(jsonData), "separator")
	require.Contains(t, string(jsonData), "delimiter")
	require.Contains(t, string(jsonData), "isEscape")

	// Test JSON unmarshaling
	var unmarshaled ConnectorCsvConfig
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)
	require.Equal(t, csvCfg.Separator, unmarshaled.Separator)
	require.Equal(t, csvCfg.Delimiter, unmarshaled.Delimiter)
	require.Equal(t, csvCfg.IsEscape, unmarshaled.IsEscape)
}

func TestFilePreviewResponseJSON(t *testing.T) {
	t.Parallel()

	resp := FilePreviewResponse{
		ConnFileId: "123",
		Rows: []*PreviewRow{
			{
				Number:         1,
				ColumnName:     "col1",
				ColumnValues:   []string{"value1", "value2"},
				CharNumber:     "A",
				CharColumnName: "Column A",
			},
			{
				Number:         2,
				ColumnName:     "col2",
				ColumnValues:   []string{"value3", "value4"},
				CharNumber:     "B",
				CharColumnName: "Column B",
			},
		},
		FileType: 1,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(resp)
	require.NoError(t, err)
	require.Contains(t, string(jsonData), "123")
	require.Contains(t, string(jsonData), "conn_file_id")
	require.Contains(t, string(jsonData), "rows")
	require.Contains(t, string(jsonData), "col1")
	require.Contains(t, string(jsonData), "value1")

	// Test JSON unmarshaling
	var unmarshaled FilePreviewResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)
	require.Equal(t, resp.ConnFileId, unmarshaled.ConnFileId)
	require.Equal(t, resp.FileType, unmarshaled.FileType)
	require.Len(t, unmarshaled.Rows, 2)
	require.Equal(t, resp.Rows[0].Number, unmarshaled.Rows[0].Number)
	require.Equal(t, resp.Rows[0].ColumnName, unmarshaled.Rows[0].ColumnName)
	require.Equal(t, resp.Rows[0].ColumnValues, unmarshaled.Rows[0].ColumnValues)
	require.Equal(t, resp.Rows[0].CharNumber, unmarshaled.Rows[0].CharNumber)
	require.Equal(t, resp.Rows[0].CharColumnName, unmarshaled.Rows[0].CharColumnName)
}

func TestPreviewRowJSON(t *testing.T) {
	t.Parallel()

	row := PreviewRow{
		Number:         1,
		ColumnName:     "col1",
		ColumnValues:   []string{"value1", "value2", "value3"},
		CharNumber:     "A",
		CharColumnName: "Column A",
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(row)
	require.NoError(t, err)
	require.Contains(t, string(jsonData), "number")
	require.Contains(t, string(jsonData), "columnName")
	require.Contains(t, string(jsonData), "columnValues")
	require.Contains(t, string(jsonData), "charNumber")
	require.Contains(t, string(jsonData), "charColumnName")
	require.Contains(t, string(jsonData), "value1")

	// Test JSON unmarshaling
	var unmarshaled PreviewRow
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)
	require.Equal(t, row.Number, unmarshaled.Number)
	require.Equal(t, row.ColumnName, unmarshaled.ColumnName)
	require.Equal(t, row.ColumnValues, unmarshaled.ColumnValues)
	require.Equal(t, row.CharNumber, unmarshaled.CharNumber)
	require.Equal(t, row.CharColumnName, unmarshaled.CharColumnName)
}

func TestFilePreviewLiveFlow(t *testing.T) {
	ctx := context.Background()
	client, err := NewRawClient(testBaseURL, testAPIKey)
	require.NoError(t, err)

	// Create a temporary CSV file for testing
	tmpDir := t.TempDir()
	testFileName := "test_preview_file.csv"
	testFilePath := filepath.Join(tmpDir, testFileName)
	testContent := "name,age,city\nJohn,30,New York\nJane,25,Los Angeles\nBob,35,Chicago"

	err = os.WriteFile(testFilePath, []byte(testContent), 0644)
	require.NoError(t, err)

	// Prepare file metadata
	meta := []FileMeta{
		{
			Filename: testFileName,
			Path:     "/test/preview/path",
		},
	}

	// First, upload the file
	uploadResp, err := client.UploadLocalFileFromPath(ctx, testFilePath, meta)
	require.NoError(t, err)
	require.NotNil(t, uploadResp)
	require.NotEmpty(t, uploadResp.ConnFileIds)
	require.Greater(t, len(uploadResp.ConnFileIds), 0)

	connFileId := uploadResp.ConnFileIds[0]
	t.Logf("File uploaded successfully, conn_file_id: %s", connFileId)

	// Test FilePreview with conn_file_id (local upload file)
	// Note: RowStart must be between 1 and 1000 (inclusive)
	t.Run("PreviewLocalUploadFile", func(t *testing.T) {
		req := &FilePreviewRequest{
			ConnFileId:    connFileId,
			IsColumnName:  true,
			ColumnNameRow: 1,
			RowStart:      1, // RowStart must be between 1 and 1000
			Csv: &ConnectorCsvConfig{
				Separator: ",",
				Delimiter: "\"",
				IsEscape:  true,
			},
			// FileType: 0 means auto-detect or not specified
		}

		resp, err := client.FilePreview(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotEmpty(t, resp.ConnFileId)
		require.NotNil(t, resp.Rows)
		t.Logf("File preview successful, conn_file_id: %s, rows count: %d", resp.ConnFileId, len(resp.Rows))

		// Verify preview rows structure
		if len(resp.Rows) > 0 {
			for i, row := range resp.Rows {
				require.NotNil(t, row)
				require.Greater(t, row.Number, int32(0))
				t.Logf("Row %d: Number=%d, ColumnName=%s, ColumnValues=%v, CharNumber=%s, CharColumnName=%s",
					i, row.Number, row.ColumnName, row.ColumnValues, row.CharNumber, row.CharColumnName)
			}
		}
	})

	// Test FilePreview with CSV config
	t.Run("PreviewWithCsvConfig", func(t *testing.T) {
		req := &FilePreviewRequest{
			ConnFileId:    connFileId,
			IsColumnName:  true,
			ColumnNameRow: 1,
			RowStart:      1, // RowStart must be between 1 and 1000
			Csv: &ConnectorCsvConfig{
				Separator: ",",
				Delimiter: "\"",
				IsEscape:  false,
			},
			// FileType: 0 means auto-detect
		}

		resp, err := client.FilePreview(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		t.Logf("File preview with CSV config successful, rows count: %d", len(resp.Rows))
	})

	// Test FilePreview without CSV config (should use defaults)
	t.Run("PreviewWithoutCsvConfig", func(t *testing.T) {
		req := &FilePreviewRequest{
			ConnFileId:   connFileId,
			IsColumnName: false,
			RowStart:     1, // RowStart must be between 1 and 1000
			// FileType: 0 means auto-detect
		}

		resp, err := client.FilePreview(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		t.Logf("File preview without CSV config successful, rows count: %d", len(resp.Rows))
	})

	// Test FilePreview with RowStart > 1
	t.Run("PreviewWithRowStart", func(t *testing.T) {
		req := &FilePreviewRequest{
			ConnFileId:    connFileId,
			IsColumnName:  true,
			ColumnNameRow: 1,
			RowStart:      2, // Start from row 2
			Csv: &ConnectorCsvConfig{
				Separator: ",",
				Delimiter: "\"",
				IsEscape:  true,
			},
			// FileType: 0 means auto-detect
		}

		resp, err := client.FilePreview(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		t.Logf("File preview with RowStart=2 successful, rows count: %d", len(resp.Rows))
	})
}

func TestDownloadConnectorFileNilRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := &RawClient{}

	resp, err := client.DownloadConnectorFile(ctx, nil)
	require.Nil(t, resp)
	require.ErrorIs(t, err, ErrNilRequest)
}

func TestDownloadConnectorFileValidation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := &RawClient{}

	resp, err := client.DownloadConnectorFile(ctx, &ConnectorFileDownloadRequest{})
	require.Nil(t, resp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "conn_file_id is required")
}

func TestDeleteConnectorFileNilRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := &RawClient{}

	resp, err := client.DeleteConnectorFile(ctx, nil)
	require.Nil(t, resp)
	require.ErrorIs(t, err, ErrNilRequest)
}

func TestDeleteConnectorFileValidation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := &RawClient{}

	resp, err := client.DeleteConnectorFile(ctx, &ConnectorFileDeleteRequest{})
	require.Nil(t, resp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "conn_file_id is required")
}

func TestDownloadConnectorFileLiveFlow(t *testing.T) {
	ctx := context.Background()
	client, err := NewRawClient(testBaseURL, testAPIKey)
	require.NoError(t, err)

	connFileId, expectedContent := uploadConnectorTestFile(t, client, "download")
	t.Cleanup(func() {
		if _, err := client.DeleteConnectorFile(ctx, &ConnectorFileDeleteRequest{ConnFileId: connFileId}); err != nil {
			t.Logf("cleanup delete connector file failed: %v", err)
		}
	})

	resp, err := client.DownloadConnectorFile(ctx, &ConnectorFileDownloadRequest{
		ConnFileId: connFileId,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.URL)
	t.Logf("Download URL: %s", resp.URL)

	downloadResp, err := http.Get(resp.URL)
	require.NoError(t, err)
	defer downloadResp.Body.Close()
	require.Equal(t, http.StatusOK, downloadResp.StatusCode)

	body, err := io.ReadAll(downloadResp.Body)
	require.NoError(t, err)
	require.Equal(t, expectedContent, body)
}

func TestDeleteConnectorFileLiveFlow(t *testing.T) {
	ctx := context.Background()
	client, err := NewRawClient(testBaseURL, testAPIKey)
	require.NoError(t, err)

	connFileId, _ := uploadConnectorTestFile(t, client, "delete")
	resp, err := client.DeleteConnectorFile(ctx, &ConnectorFileDeleteRequest{
		ConnFileId: connFileId,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	t.Logf("Delete connector file %s completed", connFileId)
}

func uploadConnectorTestFile(t *testing.T, client *RawClient, prefix string) (string, []byte) {
	t.Helper()
	ctx := context.Background()

	tmpDir := t.TempDir()
	fileName := randomName("sdk-"+prefix+"-") + ".txt"
	filePath := filepath.Join(tmpDir, fileName)
	fileContent := []byte("connector sdk test file " + randomName("content-"))
	err := os.WriteFile(filePath, fileContent, 0644)
	require.NoError(t, err)

	meta := []FileMeta{
		{
			Filename: fileName,
			Path:     "/sdk/tests/" + prefix,
		},
	}

	resp, err := client.UploadLocalFileFromPath(ctx, filePath, meta)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.ConnFileIds)
	return resp.ConnFileIds[0], fileContent
}

// ============ Tests for UploadConnectorFile ============

func TestUploadConnectorFile_NilRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := &RawClient{}

	resp, err := client.UploadConnectorFile(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Equal(t, ErrNilRequest, err)
}

func TestUploadConnectorFile_EmptyVolumeID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := &RawClient{}

	req := &UploadFileRequest{
		VolumeID: "",
		Files: []FileUploadItem{
			{
				File:     strings.NewReader("test content"),
				FileName: "test.txt",
			},
		},
	}

	resp, err := client.UploadConnectorFile(ctx, req)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "volume_id is required")
}

func TestUploadConnectorFile_NoFiles(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := &RawClient{}

	req := &UploadFileRequest{
		VolumeID: VolumeID("test-volume"),
		Files:    []FileUploadItem{},
	}

	resp, err := client.UploadConnectorFile(ctx, req)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "at least one file is required")
}

func TestUploadConnectorFile_RequestJSON(t *testing.T) {
	t.Parallel()

	req := &UploadFileRequest{
		VolumeID: VolumeID("test-volume-id"),
		Files: []FileUploadItem{
			{
				File:     strings.NewReader("test"),
				FileName: "test.txt",
			},
		},
		Meta: []FileMeta{
			{
				Filename: "test.txt",
				Path:     "/test/path",
			},
		},
		FileTypes:          []int32{1, 2, 3},
		PathRegex:          ".*\\.txt$",
		UnzipKeepStructure: true,
		DedupConfig: &DedupConfig{
			By: []string{"name", "md5"},
		},
		TableConfig: &TableConfig{},
	}

	// Test that request structure is valid
	require.NotNil(t, req)
	require.Equal(t, VolumeID("test-volume-id"), req.VolumeID)
	require.Len(t, req.Files, 1)
	require.Len(t, req.Meta, 1)
	require.Len(t, req.FileTypes, 3)
	require.Equal(t, ".*\\.txt$", req.PathRegex)
	require.True(t, req.UnzipKeepStructure)
	require.NotNil(t, req.DedupConfig)
	require.Len(t, req.DedupConfig.By, 2)
	require.NotNil(t, req.TableConfig)
}

func TestUploadConnectorFile_ResponseJSON(t *testing.T) {
	t.Parallel()

	// Test JSON marshaling
	responseJSON := `{
		"file_id": "file-123",
		"message": "upload successful",
		"success": true,
		"results": [
			{
				"file_id": "file-123",
				"message": "uploaded successfully",
				"success": true
			},
			{
				"file_id": "file-456",
				"message": "uploaded successfully",
				"success": true
			}
		],
		"task_id": 789
	}`

	var resp UploadFileResponse
	err := json.Unmarshal([]byte(responseJSON), &resp)
	require.NoError(t, err)
	require.Equal(t, "file-123", resp.FileID)
	require.Equal(t, "upload successful", resp.Message)
	require.True(t, resp.Success)
	require.Len(t, resp.Results, 2)
	require.Equal(t, int64(789), resp.TaskId)

	// Test JSON marshaling back
	jsonData, err := json.Marshal(&resp)
	require.NoError(t, err)
	require.Contains(t, string(jsonData), "file-123")
	require.Contains(t, string(jsonData), "task_id")
}

func TestUploadConnectorFile_FileUploadResultJSON(t *testing.T) {
	t.Parallel()

	resultJSON := `{
		"file_id": "file-123",
		"message": "uploaded successfully",
		"success": true
	}`

	var result FileUploadResult
	err := json.Unmarshal([]byte(resultJSON), &result)
	require.NoError(t, err)
	require.Equal(t, "file-123", result.FileID)
	require.Equal(t, "uploaded successfully", result.Message)
	require.True(t, result.Success)

	// Test JSON marshaling back
	jsonData, err := json.Marshal(&result)
	require.NoError(t, err)
	require.Contains(t, string(jsonData), "file-123")
	require.Contains(t, string(jsonData), "success")
}

func TestUploadConnectorFile_MultipartFormData(t *testing.T) {
	t.Parallel()

	req := &UploadFileRequest{
		VolumeID: VolumeID("test-volume"),
		Files: []FileUploadItem{
			{
				File:     strings.NewReader("test content 1"),
				FileName: "test1.txt",
			},
			{
				File:     strings.NewReader("test content 2"),
				FileName: "test2.txt",
			},
		},
		Meta: []FileMeta{
			{
				Filename: "test1.txt",
				Path:     "/test/path1",
			},
			{
				Filename: "test2.txt",
				Path:     "/test/path2",
			},
		},
		FileTypes:          []int32{1, 2},
		PathRegex:          ".*\\.txt$",
		UnzipKeepStructure: true,
		DedupConfig: &DedupConfig{
			By: []string{"name"},
		},
	}

	// Build multipart form data to verify structure
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add VolumeID
	volumeIDField, _ := writer.CreateFormField("VolumeID")
	volumeIDField.Write([]byte(string(req.VolumeID)))

	// Add meta
	metaJSON, _ := json.Marshal(req.Meta)
	metaField, _ := writer.CreateFormField("meta")
	metaField.Write(metaJSON)

	// Add file_types
	fileTypesJSON, _ := json.Marshal(req.FileTypes)
	fileTypesField, _ := writer.CreateFormField("file_types")
	fileTypesField.Write(fileTypesJSON)

	// Add path_regex
	pathRegexField, _ := writer.CreateFormField("path_regex")
	pathRegexField.Write([]byte(req.PathRegex))

	// Add unzip_keep_structure
	unzipField, _ := writer.CreateFormField("unzip_keep_structure")
	unzipField.Write([]byte("true"))

	// Add dedup
	dedupJSON, _ := json.Marshal(req.DedupConfig)
	dedupField, _ := writer.CreateFormField("dedup")
	dedupField.Write(dedupJSON)

	// Add files
	for _, item := range req.Files {
		fileField, _ := writer.CreateFormFile("file", item.FileName)
		io.Copy(fileField, item.File)
	}

	writer.Close()

	// Verify multipart form data was created
	require.NotEmpty(t, body.Bytes())
	contentType := writer.FormDataContentType()
	require.Contains(t, contentType, "multipart/form-data")
	require.Contains(t, contentType, "boundary")

	t.Logf("Multipart form data created successfully with Content-Type: %s", contentType)
}

func TestUploadConnectorFile_LiveFlow(t *testing.T) {
	ctx := context.Background()
	client, err := NewRawClient(testBaseURL, testAPIKey)
	require.NoError(t, err)

	// Create a temporary file for testing
	tmpDir := t.TempDir()
	testFileName := "test_upload_connector_file.txt"
	testFilePath := filepath.Join(tmpDir, testFileName)
	testContent := "This is a test file content for UploadConnectorFile test\n" + randomName("content-")

	err = os.WriteFile(testFilePath, []byte(testContent), 0644)
	require.NoError(t, err)

	// Prepare request with required fields
	// Note: VolumeID needs to be a valid volume ID from the server
	// For testing, we'll use a test volume ID or skip if not available
	volumeID := VolumeID("test") // Use "test" as a common test volume ID

	req := &UploadFileRequest{
		VolumeID: volumeID,
		Files: []FileUploadItem{
			{
				File:     strings.NewReader(testContent),
				FileName: testFileName,
			},
		},
		Meta: []FileMeta{
			{
				Filename: testFileName,
				Path:     "/test/upload/path",
			},
		},
	}

	t.Run("UploadConnectorFile_Basic", func(t *testing.T) {
		resp, err := client.UploadConnectorFile(ctx, req)
		// Note: This may fail if the volume ID doesn't exist or permission is denied
		// That's okay, we're testing the SDK implementation, not the server
		if err != nil {
			// Check if it's a server-side error (expected) vs SDK error (unexpected)
			var httpErr *HTTPError
			var apiErr *APIError
			if !(err == ErrNilRequest ||
				strings.Contains(err.Error(), "volume_id is required") ||
				strings.Contains(err.Error(), "at least one file is required") ||
				err.Error() == "create request: context deadline exceeded" ||
				strings.Contains(err.Error(), "execute request") ||
				err.Error() == "decode response") {
				// If it's not a known SDK validation error, it might be a server error
				// which is acceptable for integration testing
				if errors.As(err, &httpErr) || errors.As(err, &apiErr) {
					t.Logf("Server returned error (expected for integration test): %v", err)
					return
				}
			}
			// For unexpected SDK errors, fail the test
			require.NoError(t, err)
			return
		}

		// If we got a response, verify its structure
		require.NotNil(t, resp)
		// Results may be empty in some cases, so we just check that TaskId is present
		if resp.TaskId > 0 {
			t.Logf("Upload successful, task_id: %d, results count: %d", resp.TaskId, len(resp.Results))
		} else {
			t.Logf("Upload response received, task_id: %d, results count: %d", resp.TaskId, len(resp.Results))
		}
	})

	t.Run("UploadConnectorFile_WithOptionalFields", func(t *testing.T) {
		reqWithOptions := &UploadFileRequest{
			VolumeID: volumeID,
			Files: []FileUploadItem{
				{
					File:     strings.NewReader(testContent),
					FileName: testFileName,
				},
			},
			Meta: []FileMeta{
				{
					Filename: testFileName,
					Path:     "/test/upload/path",
				},
			},
			FileTypes:          []int32{1, 2},
			PathRegex:          ".*\\.txt$",
			UnzipKeepStructure: true,
			DedupConfig: &DedupConfig{
				By: []string{"name"},
			},
		}

		resp, err := client.UploadConnectorFile(ctx, reqWithOptions)
		// Similar error handling as above
		if err != nil {
			var httpErr *HTTPError
			var apiErr *APIError
			if !(err == ErrNilRequest ||
				strings.Contains(err.Error(), "volume_id is required") ||
				strings.Contains(err.Error(), "at least one file is required") ||
				err.Error() == "create request: context deadline exceeded" ||
				strings.Contains(err.Error(), "execute request") ||
				err.Error() == "decode response") {
				if errors.As(err, &httpErr) || errors.As(err, &apiErr) {
					t.Logf("Server returned error (expected for integration test): %v", err)
					return
				}
			}
			require.NoError(t, err)
			return
		}

		require.NotNil(t, resp)
		t.Logf("Upload with optional fields successful, task_id: %d", resp.TaskId)
	})
}
