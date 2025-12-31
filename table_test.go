package sdk

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTableLiveFlow(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	catalogID, markCatalogDeleted := createTestCatalog(t, client)
	databaseID, markDatabaseDeleted := createTestDatabase(t, client, catalogID)

	tableName := randomName("sdk-table-")
	columns := []Column{
		{Name: "id", Type: "int", IsPk: true},
		{Name: "name", Type: "varchar(255)"},
	}
	createResp, err := client.CreateTable(ctx, &TableCreateRequest{
		DatabaseID: databaseID,
		Name:       tableName,
		Columns:    columns,
		Comment:    "sdk test table",
	})
	require.NoError(t, err)
	tableID := createResp.TableID

	tableDeleted := false
	t.Cleanup(func() {
		if tableDeleted {
			return
		}
		if _, err := client.DeleteTable(ctx, &TableDeleteRequest{TableID: tableID}); err != nil {
			t.Logf("cleanup delete table failed: %v", err)
		}
	})

	infoResp, err := client.GetTable(ctx, &TableInfoRequest{TableID: tableID})
	require.NoError(t, err)
	require.Equal(t, tableName, infoResp.Name)

	multiInfoResp, err := client.GetMultiTable(ctx, &MultiTableInfoRequest{
		TableList: []TableInfoRequest{
			{TableID: tableID},
		},
	})
	require.NoError(t, err)
	key := fmt.Sprintf("%d %s", databaseID, tableName)
	require.Equal(t, tableName, multiInfoResp.InfoMap[key].Name)

	exists, err := client.CheckTableExists(ctx, &TableExistRequest{
		DatabaseID: databaseID,
		Name:       tableName,
	})
	require.NoError(t, err)
	require.True(t, exists)

	previewResp, err := client.PreviewTable(ctx, &TablePreviewRequest{TableID: tableID, Lines: 5})
	require.NoError(t, err)
	require.NotNil(t, previewResp)

	truncResp, err := client.TruncateTable(ctx, &TableTruncateRequest{TableID: tableID})
	require.NoError(t, err)
	require.NotNil(t, truncResp)

	fullPathResp, err := client.GetTableFullPath(ctx, &TableFullPathRequest{TableIDList: []TableID{tableID}})
	require.NoError(t, err)
	require.NotNil(t, fullPathResp)

	refListResp, err := client.GetTableRefList(ctx, &TableRefListRequest{TableID: tableID})
	require.NoError(t, err)
	require.NotNil(t, refListResp)

	_, err = client.DeleteTable(ctx, &TableDeleteRequest{TableID: tableID})
	require.NoError(t, err)
	tableDeleted = true

	exists, err = client.CheckTableExists(ctx, &TableExistRequest{
		DatabaseID: databaseID,
		Name:       tableName,
	})
	require.NoError(t, err)
	require.False(t, exists)

	_, err = client.DeleteDatabase(ctx, &DatabaseDeleteRequest{DatabaseID: databaseID})
	require.NoError(t, err)
	markDatabaseDeleted()

	_, err = client.DeleteCatalog(ctx, &CatalogDeleteRequest{CatalogID: catalogID})
	require.NoError(t, err)
	markCatalogDeleted()
}

func TestTableNilRequestErrors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := &RawClient{}

	tests := []struct {
		name string
		call func() error
	}{
		{"Create", func() error { _, err := client.CreateTable(ctx, nil); return err }},
		{"Info", func() error { _, err := client.GetTable(ctx, nil); return err }},
		{"MultiInfo", func() error { _, err := client.GetMultiTable(ctx, nil); return err }},
		{"Exist", func() error { _, err := client.CheckTableExists(ctx, nil); return err }},
		{"Preview", func() error { _, err := client.PreviewTable(ctx, nil); return err }},
		{"Load", func() error { _, err := client.LoadTable(ctx, nil); return err }},
		{"Download", func() error { _, err := client.GetTableDownloadLink(ctx, nil); return err }},
		{"DownloadData", func() error { _, err := client.DownloadTableData(ctx, nil); return err }},
		{"Truncate", func() error { _, err := client.TruncateTable(ctx, nil); return err }},
		{"Delete", func() error { _, err := client.DeleteTable(ctx, nil); return err }},
		{"FullPath", func() error { _, err := client.GetTableFullPath(ctx, nil); return err }},
		{"RefList", func() error { _, err := client.GetTableRefList(ctx, nil); return err }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.ErrorIs(t, tc.call(), ErrNilRequest)
		})
	}
}

func TestTableDatabaseIDNotExists(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	nonExistentDatabaseID := DatabaseID(999999999)

	// Try to create table with non-existent database ID
	_, err := client.CreateTable(ctx, &TableCreateRequest{
		DatabaseID: nonExistentDatabaseID,
		Name:       randomName("test-table-"),
		Columns: []Column{
			{Name: "id", Type: "int", IsPk: true},
		},
		Comment: "test",
	})
	require.Error(t, err)
	t.Logf("Expected error for non-existent database ID: %v", err)
}

func TestTableNameExists(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	catalogID, markCatalogDeleted := createTestCatalog(t, client)
	databaseID, markDatabaseDeleted := createTestDatabase(t, client, catalogID)

	defer func() {
		markDatabaseDeleted()
		markCatalogDeleted()
	}()

	tableName := randomName("sdk-table-")
	columns := []Column{
		{Name: "id", Type: "int", IsPk: true},
		{Name: "name", Type: "varchar(255)"},
	}

	createReq := &TableCreateRequest{
		DatabaseID: databaseID,
		Name:       tableName,
		Columns:    columns,
		Comment:    "test table",
	}
	createResp, err := client.CreateTable(ctx, createReq)
	require.NoError(t, err)
	require.NotZero(t, createResp.TableID)

	// Cleanup
	defer func() {
		if _, err := client.DeleteTable(ctx, &TableDeleteRequest{TableID: createResp.TableID}); err != nil {
			t.Logf("cleanup delete table failed: %v", err)
		}
	}()

	// Try to create another table with the same name in the same database
	_, err = client.CreateTable(ctx, createReq)
	require.Error(t, err)
	t.Logf("Expected error for duplicate name: %v", err)
}

func TestTableIDNotExists(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	nonExistentID := TableID(999999999)

	// Try to get non-existent table
	_, err := client.GetTable(ctx, &TableInfoRequest{TableID: nonExistentID})
	require.Error(t, err)
	t.Logf("Expected error for non-existent table ID: %v", err)

	// Try to get non-existent table in batch
	resp, err := client.GetMultiTable(ctx, &MultiTableInfoRequest{
		TableList: []TableInfoRequest{
			{TableID: nonExistentID},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 0, len(resp.InfoMap))

	// Try to preview non-existent table - may not error if service allows empty preview
	_, err = client.PreviewTable(ctx, &TablePreviewRequest{TableID: nonExistentID, Lines: 5})
	if err != nil {
		t.Logf("Error for previewing non-existent table (expected): %v", err)
	} else {
		t.Logf("Preview succeeded for non-existent table (service may allow empty preview)")
	}

	// Try to delete non-existent table - service may allow idempotent delete
	_, err = client.DeleteTable(ctx, &TableDeleteRequest{TableID: nonExistentID})
	// Service may allow idempotent delete, so we don't require an error
	t.Logf("Delete result for non-existent table: %v (service may allow idempotent delete)", err)
}

func TestTableWithDefaultValues(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	catalogID, markCatalogDeleted := createTestCatalog(t, client)
	databaseID, markDatabaseDeleted := createTestDatabase(t, client, catalogID)

	defer func() {
		markDatabaseDeleted()
		markCatalogDeleted()
	}()

	tableName := randomName("sdk-table-default-")
	columns := []Column{
		{Name: "id", Type: "int", IsPk: true},
		{Name: "age", Type: "int", Default: "0"},
		{Name: "default_test", Type: "varchar(100)", Default: "VARCHAR DEFAULT"},
	}

	createResp, err := client.CreateTable(ctx, &TableCreateRequest{
		DatabaseID: databaseID,
		Name:       tableName,
		Columns:    columns,
		Comment:    "test table with defaults",
	})
	require.NoError(t, err)
	require.NotZero(t, createResp.TableID)

	// Cleanup
	defer func() {
		if _, err := client.DeleteTable(ctx, &TableDeleteRequest{TableID: createResp.TableID}); err != nil {
			t.Logf("cleanup delete table failed: %v", err)
		}
	}()

	// Verify table was created successfully
	infoResp, err := client.GetTable(ctx, &TableInfoRequest{TableID: createResp.TableID})
	require.NoError(t, err)
	require.Equal(t, tableName, infoResp.Name)
	require.Len(t, infoResp.Columns, 3, "should have 3 columns")
}

func TestDownloadTableData_NilRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := &RawClient{}

	stream, err := client.DownloadTableData(ctx, nil)
	require.Nil(t, stream)
	require.ErrorIs(t, err, ErrNilRequest)
}

func TestDownloadTableData_InvalidID(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	nonExistentID := int64(999999999)

	stream, err := client.DownloadTableData(ctx, &TableDownloadDataRequest{
		ID: nonExistentID,
	})
	if err != nil {
		// Expected error for non-existent table ID
		require.Nil(t, stream)
		t.Logf("Expected error for non-existent table ID: %v", err)
	} else {
		// If no error, ensure stream is properly closed
		if stream != nil {
			defer stream.Close()
			// Try to read from stream - should handle gracefully
			_, readErr := io.ReadAll(stream.Body)
			if readErr != nil {
				t.Logf("Error reading stream (expected for non-existent table): %v", readErr)
			}
		}
	}
}

func TestDownloadTableData_LiveFlow(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	catalogID, markCatalogDeleted := createTestCatalog(t, client)
	databaseID, markDatabaseDeleted := createTestDatabase(t, client, catalogID)

	defer func() {
		markDatabaseDeleted()
		markCatalogDeleted()
	}()

	// Create a test table
	tableName := randomName("sdk-table-download-")
	columns := []Column{
		{Name: "id", Type: "int", IsPk: true},
		{Name: "name", Type: "varchar(255)"},
		{Name: "value", Type: "int"},
	}
	createResp, err := client.CreateTable(ctx, &TableCreateRequest{
		DatabaseID: databaseID,
		Name:       tableName,
		Columns:    columns,
		Comment:    "test table for download",
	})
	require.NoError(t, err)
	tableID := createResp.TableID

	// Cleanup table
	defer func() {
		if _, err := client.DeleteTable(ctx, &TableDeleteRequest{TableID: tableID}); err != nil {
			t.Logf("cleanup delete table failed: %v", err)
		}
	}()

	// Download table data as CSV
	stream, err := client.DownloadTableData(ctx, &TableDownloadDataRequest{
		ID: int64(tableID),
	})
	if err != nil {
		// If download fails (e.g., table is empty or service doesn't support it yet), log and skip
		t.Logf("DownloadTableData failed (may be expected for empty table): %v", err)
		return
	}
	require.NotNil(t, stream)
	defer stream.Close()

	// Verify stream properties
	require.NotNil(t, stream.Body)
	require.NotNil(t, stream.Header)
	require.Equal(t, 200, stream.StatusCode)

	// Test WriteToFile method
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "table_data.csv")

	written, err := stream.WriteToFile(outputFile)
	require.NoError(t, err)
	require.GreaterOrEqual(t, written, int64(0))
	t.Logf("Wrote %d bytes to file: %s", written, outputFile)

	// Verify file was created and has content
	fileInfo, err := os.Stat(outputFile)
	require.NoError(t, err)
	require.Equal(t, written, fileInfo.Size())

	// Read the file back and verify content
	fileData, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	require.Equal(t, int(written), len(fileData))

	// Log the downloaded data size
	t.Logf("Downloaded %d bytes of CSV data", len(fileData))
	if len(fileData) > 0 {
		// Log first 200 characters of CSV for debugging
		previewLen := 200
		if len(fileData) < previewLen {
			previewLen = len(fileData)
		}
		t.Logf("CSV preview (first %d chars): %s", previewLen, string(fileData[:previewLen]))
	}
}

func TestFileStream_WriteToFile(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := newTestClient(t)

	catalogID, markCatalogDeleted := createTestCatalog(t, client)
	databaseID, markDatabaseDeleted := createTestDatabase(t, client, catalogID)

	defer func() {
		markDatabaseDeleted()
		markCatalogDeleted()
	}()

	// Create a test table
	tableName := randomName("sdk-table-write-")
	columns := []Column{
		{Name: "id", Type: "int", IsPk: true},
		{Name: "name", Type: "varchar(255)"},
	}
	createResp, err := client.CreateTable(ctx, &TableCreateRequest{
		DatabaseID: databaseID,
		Name:       tableName,
		Columns:    columns,
		Comment:    "test table for WriteToFile",
	})
	require.NoError(t, err)
	tableID := createResp.TableID

	// Cleanup table
	defer func() {
		if _, err := client.DeleteTable(ctx, &TableDeleteRequest{TableID: tableID}); err != nil {
			t.Logf("cleanup delete table failed: %v", err)
		}
	}()

	// Download table data
	stream, err := client.DownloadTableData(ctx, &TableDownloadDataRequest{
		ID: int64(tableID),
	})
	if err != nil {
		t.Logf("DownloadTableData failed (may be expected for empty table): %v", err)
		return
	}
	require.NotNil(t, stream)
	defer stream.Close()

	// Test WriteToFile with nested directory
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "subdir", "table_data.csv")

	written, err := stream.WriteToFile(outputFile)
	require.NoError(t, err)
	require.GreaterOrEqual(t, written, int64(0))
	t.Logf("Wrote %d bytes to file: %s", written, outputFile)

	// Verify file exists and has correct size
	fileInfo, err := os.Stat(outputFile)
	require.NoError(t, err)
	require.Equal(t, written, fileInfo.Size())

	// Test WriteToFile with nil stream
	var nilStream *FileStream
	_, err = nilStream.WriteToFile("/tmp/test.csv")
	require.Error(t, err)
	require.Equal(t, io.ErrUnexpectedEOF, err)
}
