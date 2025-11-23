package sdk

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDatabaseLiveCRUD(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	catalogResp, err := client.CreateCatalog(ctx, &CatalogCreateRequest{
		CatalogName: randomName("sdk-db-catalog-"),
	})
	require.NoError(t, err)
	catalogID := catalogResp.CatalogID

	catalogDeleted := false
	t.Cleanup(func() {
		if catalogDeleted {
			return
		}
		if _, err := client.DeleteCatalog(ctx, &CatalogDeleteRequest{CatalogID: catalogID}); err != nil {
			t.Logf("cleanup delete catalog failed: %v", err)
		}
	})

	createResp, err := client.CreateDatabase(ctx, &DatabaseCreateRequest{
		DatabaseName: randomName("sdk-db-"),
		CatalogID:    catalogID,
	})
	require.NoError(t, err)
	require.NotZero(t, createResp.DatabaseID)
	dbID := createResp.DatabaseID
	dbDeleted := false
	t.Cleanup(func() {
		if dbDeleted {
			return
		}
		if _, err := client.DeleteDatabase(ctx, &DatabaseDeleteRequest{DatabaseID: dbID}); err != nil {
			t.Logf("cleanup delete database failed: %v", err)
		}
	})

	infoResp, err := client.GetDatabase(ctx, &DatabaseInfoRequest{DatabaseID: dbID})
	require.NoError(t, err)
	require.NotEmpty(t, infoResp.DatabaseName)

	_, err = client.UpdateDatabase(ctx, &DatabaseUpdateRequest{
		DatabaseID: dbID,
		Comment:    "updated from sdk tests",
	})
	require.NoError(t, err)

	listResp, err := client.ListDatabases(ctx, &DatabaseListRequest{CatalogID: catalogID})
	require.NoError(t, err)
	require.NotNil(t, listResp)

	childrenResp, err := client.GetDatabaseChildren(ctx, &DatabaseChildrenRequest{DatabaseID: dbID})
	require.NoError(t, err)
	require.NotNil(t, childrenResp)

	refResp, err := client.GetDatabaseRefList(ctx, &DatabaseRefListRequest{DatabaseID: dbID})
	require.NoError(t, err)
	require.NotNil(t, refResp)

	_, err = client.DeleteDatabase(ctx, &DatabaseDeleteRequest{DatabaseID: dbID})
	require.NoError(t, err)
	dbDeleted = true

	_, err = client.DeleteCatalog(ctx, &CatalogDeleteRequest{CatalogID: catalogID})
	require.NoError(t, err)
	catalogDeleted = true
}

func TestDatabaseNilRequestErrors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := &RawClient{}

	tests := []struct {
		name string
		call func() error
	}{
		{"Create", func() error { _, err := client.CreateDatabase(ctx, nil); return err }},
		{"Delete", func() error { _, err := client.DeleteDatabase(ctx, nil); return err }},
		{"Update", func() error { _, err := client.UpdateDatabase(ctx, nil); return err }},
		{"Info", func() error { _, err := client.GetDatabase(ctx, nil); return err }},
		{"List", func() error { _, err := client.ListDatabases(ctx, nil); return err }},
		{"Children", func() error { _, err := client.GetDatabaseChildren(ctx, nil); return err }},
		{"RefList", func() error { _, err := client.GetDatabaseRefList(ctx, nil); return err }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.ErrorIs(t, tc.call(), ErrNilRequest)
		})
	}
}

func TestDatabaseCatalogIDNotExists(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	nonExistentCatalogID := CatalogID(999999999)

	// Try to create database with non-existent catalog ID
	_, err := client.CreateDatabase(ctx, &DatabaseCreateRequest{
		CatalogID:    nonExistentCatalogID,
		DatabaseName: randomName("test-db-"),
		Comment:      "test",
	})
	require.Error(t, err)
	t.Logf("Expected error for non-existent catalog ID: %v", err)
}

func TestDatabaseNameExists(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	catalogID, markCatalogDeleted := createTestCatalog(t, client)
	defer markCatalogDeleted()

	databaseName := randomName("sdk-db-")
	createReq := &DatabaseCreateRequest{
		DatabaseName: databaseName,
		CatalogID:    catalogID,
		Comment:      "test database",
	}
	createResp, err := client.CreateDatabase(ctx, createReq)
	require.NoError(t, err)
	require.NotZero(t, createResp.DatabaseID)

	// Cleanup
	defer func() {
		if _, err := client.DeleteDatabase(ctx, &DatabaseDeleteRequest{DatabaseID: createResp.DatabaseID}); err != nil {
			t.Logf("cleanup delete database failed: %v", err)
		}
	}()

	// Try to create another database with the same name in the same catalog
	_, err = client.CreateDatabase(ctx, createReq)
	require.Error(t, err)
	t.Logf("Expected error for duplicate name: %v", err)
}

func TestDatabaseInvalidName(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	catalogID, markCatalogDeleted := createTestCatalog(t, client)
	defer markCatalogDeleted()

	tests := []struct {
		name         string
		databaseName string
	}{
		{"SpecialChars", "aa\"b'"},
		{"Empty", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			createReq := &DatabaseCreateRequest{
				DatabaseName: tc.databaseName,
				CatalogID:    catalogID,
				Comment:      "test",
			}
			_, err := client.CreateDatabase(ctx, createReq)
			require.Error(t, err, "should fail for invalid name: %s", tc.databaseName)
			t.Logf("Expected error for invalid name '%s': %v", tc.databaseName, err)
		})
	}
}

func TestDatabaseIDNotExists(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	nonExistentID := DatabaseID(999999999)

	// Try to get non-existent database
	_, err := client.GetDatabase(ctx, &DatabaseInfoRequest{DatabaseID: nonExistentID})
	require.Error(t, err)
	t.Logf("Expected error for non-existent database ID: %v", err)

	// Try to update non-existent database
	_, err = client.UpdateDatabase(ctx, &DatabaseUpdateRequest{
		DatabaseID: nonExistentID,
		Comment:    "test",
	})
	require.Error(t, err)
	t.Logf("Expected error for updating non-existent database: %v", err)

	// Try to delete non-existent database
	_, err = client.DeleteDatabase(ctx, &DatabaseDeleteRequest{DatabaseID: nonExistentID})
	require.Error(t, err)
	t.Logf("Expected error for deleting non-existent database: %v", err)
}
