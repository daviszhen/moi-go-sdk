package sdk

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCatalogLiveCRUD(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	createReq := &CatalogCreateRequest{
		CatalogName: randomName("sdk-catalog-"),
	}
	createResp, err := client.CreateCatalog(ctx, createReq)
	require.NoError(t, err)
	require.NotZero(t, createResp.CatalogID)

	catalogID := createResp.CatalogID
	cleanupDone := false
	t.Cleanup(func() {
		if cleanupDone {
			return
		}
		if _, err := client.DeleteCatalog(ctx, &CatalogDeleteRequest{CatalogID: catalogID}); err != nil {
			t.Logf("cleanup delete catalog failed: %v", err)
		}
	})

	infoResp, err := client.GetCatalog(ctx, &CatalogInfoRequest{CatalogID: catalogID})
	require.NoError(t, err)
	require.Equal(t, createReq.CatalogName, infoResp.CatalogName)

	updatedName := randomName("sdk-catalog-updated-")
	_, err = client.UpdateCatalog(ctx, &CatalogUpdateRequest{
		CatalogID:   catalogID,
		CatalogName: updatedName,
	})
	require.NoError(t, err)

	infoResp, err = client.GetCatalog(ctx, &CatalogInfoRequest{CatalogID: catalogID})
	require.NoError(t, err)
	require.Equal(t, updatedName, infoResp.CatalogName)

	listResp, err := client.ListCatalogs(ctx)
	require.NoError(t, err)
	require.NotNil(t, listResp)

	treeResp, err := client.GetCatalogTree(ctx)
	require.NoError(t, err)
	require.NotNil(t, treeResp)

	refResp, err := client.GetCatalogRefList(ctx, &CatalogRefListRequest{CatalogID: catalogID})
	require.NoError(t, err)
	require.NotNil(t, refResp)

	_, err = client.DeleteCatalog(ctx, &CatalogDeleteRequest{CatalogID: catalogID})
	require.NoError(t, err)
	cleanupDone = true
}

func TestCatalogNilRequestErrors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := &RawClient{}

	tests := []struct {
		name string
		call func() error
	}{
		{"Create", func() error { _, err := client.CreateCatalog(ctx, nil); return err }},
		{"Delete", func() error { _, err := client.DeleteCatalog(ctx, nil); return err }},
		{"Update", func() error { _, err := client.UpdateCatalog(ctx, nil); return err }},
		{"Get", func() error { _, err := client.GetCatalog(ctx, nil); return err }},
		{"RefList", func() error { _, err := client.GetCatalogRefList(ctx, nil); return err }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.ErrorIs(t, tc.call(), ErrNilRequest)
		})
	}
}

func TestCatalogNameExists(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	catalogName := randomName("sdk-catalog-")
	createReq := &CatalogCreateRequest{
		CatalogName: catalogName,
		Comment:     "test catalog",
	}
	createResp, err := client.CreateCatalog(ctx, createReq)
	require.NoError(t, err)
	require.NotZero(t, createResp.CatalogID)

	// Cleanup
	defer func() {
		if _, err := client.DeleteCatalog(ctx, &CatalogDeleteRequest{CatalogID: createResp.CatalogID}); err != nil {
			t.Logf("cleanup delete catalog failed: %v", err)
		}
	}()

	// Try to create another catalog with the same name
	_, err = client.CreateCatalog(ctx, createReq)
	require.Error(t, err)
	// Service should return an error about name already existing
	t.Logf("Expected error for duplicate name: %v", err)
}

func TestCatalogInvalidName(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	tests := []struct {
		name        string
		catalogName string
	}{
		{"TooLong", string(make([]byte, 300))}, // Name too long
		{"SpecialChars", "\"aa'"},              // Special characters
		{"Empty", ""},                           // Empty name
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			createReq := &CatalogCreateRequest{
				CatalogName: tc.catalogName,
				Comment:     "test",
			}
			_, err := client.CreateCatalog(ctx, createReq)
			require.Error(t, err, "should fail for invalid name: %s", tc.catalogName)
			t.Logf("Expected error for invalid name '%s': %v", tc.catalogName, err)
		})
	}
}

func TestCatalogIDNotExists(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	nonExistentID := CatalogID(999999999)

	// Try to get non-existent catalog
	_, err := client.GetCatalog(ctx, &CatalogInfoRequest{CatalogID: nonExistentID})
	require.Error(t, err)
	t.Logf("Expected error for non-existent catalog ID: %v", err)

	// Try to update non-existent catalog
	_, err = client.UpdateCatalog(ctx, &CatalogUpdateRequest{
		CatalogID:   nonExistentID,
		CatalogName: randomName("test-"),
	})
	require.Error(t, err)
	t.Logf("Expected error for updating non-existent catalog: %v", err)

	// Try to delete non-existent catalog
	_, err = client.DeleteCatalog(ctx, &CatalogDeleteRequest{CatalogID: nonExistentID})
	require.Error(t, err)
	t.Logf("Expected error for deleting non-existent catalog: %v", err)
}

func TestCatalogTreeWithDatabaseAndTable(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	// Create catalog
	catalogID, markCatalogDeleted := createTestCatalog(t, client)

	// Create database
	databaseID, markDatabaseDeleted := createTestDatabase(t, client, catalogID)

	// Create table
	tableID, markTableDeleted := createTestTable(t, client, databaseID)

	// Cleanup
	defer func() {
		markTableDeleted()
		markDatabaseDeleted()
		markCatalogDeleted()
	}()

	// Get catalog tree
	treeResp, err := client.GetCatalogTree(ctx)
	require.NoError(t, err)
	require.NotNil(t, treeResp)
	require.NotEmpty(t, treeResp.Tree, "tree should not be empty")

	// Find our catalog in the tree
	var foundCatalog bool
	var foundDatabase bool
	var foundTable bool

	catalogIDStr := fmt.Sprintf("%d", catalogID)
	databaseIDStr := fmt.Sprintf("%d", databaseID)
	tableIDStr := fmt.Sprintf("%d", tableID)

	for _, catalogNode := range treeResp.Tree {
		if catalogNode.ID == catalogIDStr {
			foundCatalog = true
			t.Logf("Found catalog in tree: %s (ID: %s)", catalogNode.Name, catalogNode.ID)

			// Check for database
			for _, dbNode := range catalogNode.NodeList {
				if dbNode.ID == databaseIDStr {
					foundDatabase = true
					t.Logf("Found database in tree: %s (ID: %s)", dbNode.Name, dbNode.ID)

					// Check for table
					for _, tableNode := range dbNode.NodeList {
						if tableNode.ID == tableIDStr {
							foundTable = true
							t.Logf("Found table in tree: %s (ID: %s)", tableNode.Name, tableNode.ID)
							break
						}
					}
					break
				}
			}
			break
		}
	}

	require.True(t, foundCatalog, "catalog should be found in tree")
	require.True(t, foundDatabase, "database should be found in tree")
	require.True(t, foundTable, "table should be found in tree")
}
