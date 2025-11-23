package sdk

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	testAPIKey  = "JeuCjV_8320G5ACwgPDHo0tJAdkESCrMPnxteLJri8IHb72dPySWkhN6uFWw41-W7qpdH4w3QCG8pJmf"
	testBaseURL = "https://freetier-01.cn-hangzhou.cluster.cn-dev.matrixone.tech"
)

func newTestClient(t *testing.T) *RawClient {
	t.Helper()
	client, err := NewRawClient(testBaseURL, testAPIKey)
	require.NoError(t, err)
	return client
}

func randomName(prefix string) string {
	return fmt.Sprintf("%s%d", prefix, time.Now().UnixNano())
}

func randomUserName() string {
	return fmt.Sprintf("sdkuser%d", time.Now().UnixNano())
}

func createTestCatalog(t *testing.T, client *RawClient) (CatalogID, func()) {
	t.Helper()
	ctx := context.Background()
	resp, err := client.CreateCatalog(ctx, &CatalogCreateRequest{
		CatalogName: randomName("sdk-cat-"),
	})
	require.NoError(t, err)
	deleted := false
	t.Cleanup(func() {
		if deleted {
			return
		}
		if _, err := client.DeleteCatalog(ctx, &CatalogDeleteRequest{CatalogID: resp.CatalogID}); err != nil {
			t.Logf("cleanup delete catalog failed: %v", err)
		}
	})
	return resp.CatalogID, func() { deleted = true }
}

func createTestDatabase(t *testing.T, client *RawClient, catalogID CatalogID) (DatabaseID, func()) {
	t.Helper()
	ctx := context.Background()
	resp, err := client.CreateDatabase(ctx, &DatabaseCreateRequest{
		DatabaseName: randomName("sdk-db-"),
		CatalogID:    catalogID,
	})
	require.NoError(t, err)
	deleted := false
	t.Cleanup(func() {
		if deleted {
			return
		}
		if _, err := client.DeleteDatabase(ctx, &DatabaseDeleteRequest{DatabaseID: resp.DatabaseID}); err != nil {
			t.Logf("cleanup delete database failed: %v", err)
		}
	})
	return resp.DatabaseID, func() { deleted = true }
}

func createTestVolume(t *testing.T, client *RawClient, databaseID DatabaseID) (VolumeID, func()) {
	t.Helper()
	ctx := context.Background()
	resp, err := client.CreateVolume(ctx, &VolumeCreateRequest{
		Name:       randomName("sdk-volume-"),
		DatabaseID: databaseID,
		Comment:    "sdk helper volume",
	})
	require.NoError(t, err)
	deleted := false
	t.Cleanup(func() {
		if deleted {
			return
		}
		if _, err := client.DeleteVolume(ctx, &VolumeDeleteRequest{VolumeID: resp.VolumeID}); err != nil {
			t.Logf("cleanup delete volume failed: %v", err)
		}
	})
	return resp.VolumeID, func() { deleted = true }
}

func createTestTable(t *testing.T, client *RawClient, databaseID DatabaseID) (TableID, func()) {
	t.Helper()
	ctx := context.Background()
	tableName := randomName("sdk-table-")
	columns := []Column{
		{Name: "id", Type: "int", IsPk: true},
		{Name: "name", Type: "varchar(255)"},
	}
	resp, err := client.CreateTable(ctx, &TableCreateRequest{
		DatabaseID: databaseID,
		Name:       tableName,
		Columns:    columns,
		Comment:    "sdk test table",
	})
	require.NoError(t, err)
	deleted := false
	t.Cleanup(func() {
		if deleted {
			return
		}
		if _, err := client.DeleteTable(ctx, &TableDeleteRequest{TableID: resp.TableID}); err != nil {
			t.Logf("cleanup delete table failed: %v", err)
		}
	})
	return resp.TableID, func() { deleted = true }
}

func createTestRole(t *testing.T, client *RawClient, privCodes []string) (RoleID, func()) {
	t.Helper()
	ctx := context.Background()
	// Use underscore instead of hyphen in role name since special characters are not allowed
	roleName := fmt.Sprintf("sdk_role_%d", time.Now().UnixNano())
	req := &RoleCreateRequest{
		RoleName:    roleName,
		PrivList:    privCodes,
		ObjPrivList: []ObjPrivResponse{}, // Use empty slice instead of nil
	}
	resp, err := client.CreateRole(ctx, req)
	require.NoError(t, err)
	deleted := false
	t.Cleanup(func() {
		if deleted {
			return
		}
		if _, err := client.DeleteRole(ctx, &RoleDeleteRequest{RoleID: resp.RoleID}); err != nil {
			t.Logf("cleanup delete role failed: %v", err)
		}
	})
	return resp.RoleID, func() { deleted = true }
}
