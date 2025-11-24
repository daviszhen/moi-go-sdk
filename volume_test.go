package sdk

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVolumeLiveFlow(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	catalogID, markCatalogDeleted := createTestCatalog(t, client)
	databaseID, markDatabaseDeleted := createTestDatabase(t, client, catalogID)

	volumeName := randomName("sdk-volume-")
	createResp, err := client.CreateVolume(ctx, &VolumeCreateRequest{
		Name:       volumeName,
		DatabaseID: databaseID,
		Comment:    "sdk volume",
	})
	require.NoError(t, err)
	volumeID := createResp.VolumeID

	volumeDeleted := false
	t.Cleanup(func() {
		if volumeDeleted {
			return
		}
		if _, err := client.DeleteVolume(ctx, &VolumeDeleteRequest{VolumeID: volumeID}); err != nil {
			t.Logf("cleanup delete volume failed: %v", err)
		}
	})

	infoResp, err := client.GetVolume(ctx, &VolumeInfoRequest{VolumeID: volumeID})
	require.NoError(t, err)
	require.Equal(t, volumeName, infoResp.VolumeName)

	_, err = client.UpdateVolume(ctx, &VolumeUpdateRequest{
		VolumeID: volumeID,
		Name:     randomName("sdk-volume-updated-"),
		Comment:  "updated",
	})
	require.NoError(t, err)

	refResp, err := client.GetVolumeRefList(ctx, &VolumeRefListRequest{VolumeID: volumeID})
	require.NoError(t, err)
	require.NotNil(t, refResp)

	fullPathResp, err := client.GetVolumeFullPath(ctx, &VolumeFullPathRequest{VolumeIDList: []VolumeID{volumeID}})
	require.NoError(t, err)
	require.NotNil(t, fullPathResp)

	_, err = client.AddVolumeWorkflowRef(ctx, &VolumeAddRefWorkflowRequest{VolumeID: volumeID})
	require.NoError(t, err)

	_, err = client.RemoveVolumeWorkflowRef(ctx, &VolumeRemoveRefWorkflowRequest{VolumeID: volumeID})
	require.NoError(t, err)

	_, err = client.DeleteVolume(ctx, &VolumeDeleteRequest{VolumeID: volumeID})
	require.NoError(t, err)
	volumeDeleted = true

	_, err = client.DeleteDatabase(ctx, &DatabaseDeleteRequest{DatabaseID: databaseID})
	require.NoError(t, err)
	markDatabaseDeleted()

	_, err = client.DeleteCatalog(ctx, &CatalogDeleteRequest{CatalogID: catalogID})
	require.NoError(t, err)
	markCatalogDeleted()
}

func TestVolumeNilRequestErrors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := &RawClient{}

	tests := []struct {
		name string
		call func() error
	}{
		{"Create", func() error { _, err := client.CreateVolume(ctx, nil); return err }},
		{"Delete", func() error { _, err := client.DeleteVolume(ctx, nil); return err }},
		{"Update", func() error { _, err := client.UpdateVolume(ctx, nil); return err }},
		{"Info", func() error { _, err := client.GetVolume(ctx, nil); return err }},
		{"RefList", func() error { _, err := client.GetVolumeRefList(ctx, nil); return err }},
		{"FullPath", func() error { _, err := client.GetVolumeFullPath(ctx, nil); return err }},
		{"AddRefWorkflow", func() error { _, err := client.AddVolumeWorkflowRef(ctx, nil); return err }},
		{"RemoveRefWorkflow", func() error { _, err := client.RemoveVolumeWorkflowRef(ctx, nil); return err }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.ErrorIs(t, tc.call(), ErrNilRequest)
		})
	}
}

func TestVolumeDatabaseIDNotExists(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	nonExistentDatabaseID := DatabaseID(999999999)

	// Try to create volume with non-existent database ID
	_, err := client.CreateVolume(ctx, &VolumeCreateRequest{
		DatabaseID: nonExistentDatabaseID,
		Name:       randomName("test-volume-"),
		Comment:    "test",
	})
	require.Error(t, err)
	t.Logf("Expected error for non-existent database ID: %v", err)
}

func TestVolumeInvalidName(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	catalogID, markCatalogDeleted := createTestCatalog(t, client)
	databaseID, markDatabaseDeleted := createTestDatabase(t, client, catalogID)

	defer func() {
		markDatabaseDeleted()
		markCatalogDeleted()
	}()

	tests := []struct {
		name      string
		volumeName string
	}{
		{"SpecialChars", "v\"o'l1"},
		{"Empty", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			createReq := &VolumeCreateRequest{
				DatabaseID: databaseID,
				Name:       tc.volumeName,
				Comment:    "test",
			}
			_, err := client.CreateVolume(ctx, createReq)
			require.Error(t, err, "should fail for invalid name: %s", tc.volumeName)
			t.Logf("Expected error for invalid name '%s': %v", tc.volumeName, err)
		})
	}
}

func TestVolumeNameExists(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	catalogID, markCatalogDeleted := createTestCatalog(t, client)
	databaseID, markDatabaseDeleted := createTestDatabase(t, client, catalogID)

	defer func() {
		markDatabaseDeleted()
		markCatalogDeleted()
	}()

	volumeName := randomName("sdk-volume-")
	createReq := &VolumeCreateRequest{
		Name:       volumeName,
		DatabaseID: databaseID,
		Comment:    "test volume",
	}
	createResp, err := client.CreateVolume(ctx, createReq)
	require.NoError(t, err)
	require.NotZero(t, createResp.VolumeID)

	// Cleanup
	defer func() {
		if _, err := client.DeleteVolume(ctx, &VolumeDeleteRequest{VolumeID: createResp.VolumeID}); err != nil {
			t.Logf("cleanup delete volume failed: %v", err)
		}
	}()

	// Try to create another volume with the same name in the same database
	_, err = client.CreateVolume(ctx, createReq)
	require.Error(t, err)
	t.Logf("Expected error for duplicate name: %v", err)
}

func TestVolumeIDNotExists(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	nonExistentID := VolumeID("999999999")

	// Try to get non-existent volume
	_, err := client.GetVolume(ctx, &VolumeInfoRequest{VolumeID: nonExistentID})
	require.Error(t, err)
	t.Logf("Expected error for non-existent volume ID: %v", err)

	// Try to update non-existent volume
	_, err = client.UpdateVolume(ctx, &VolumeUpdateRequest{
		VolumeID: nonExistentID,
		Name:     randomName("test-"),
		Comment:  "test",
	})
	require.Error(t, err)
	t.Logf("Expected error for updating non-existent volume: %v", err)

	// Try to delete non-existent volume
	_, err = client.DeleteVolume(ctx, &VolumeDeleteRequest{VolumeID: nonExistentID})
	require.Error(t, err)
	t.Logf("Expected error for deleting non-existent volume: %v", err)
}

func TestVolumeUpdateNameExists(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	catalogID, markCatalogDeleted := createTestCatalog(t, client)
	databaseID, markDatabaseDeleted := createTestDatabase(t, client, catalogID)

	defer func() {
		markDatabaseDeleted()
		markCatalogDeleted()
	}()

	// Create first volume
	volumeName1 := randomName("sdk-volume-1-")
	createResp1, err := client.CreateVolume(ctx, &VolumeCreateRequest{
		Name:       volumeName1,
		DatabaseID: databaseID,
		Comment:    "test volume 1",
	})
	require.NoError(t, err)
	require.NotZero(t, createResp1.VolumeID)

	// Create second volume
	volumeName2 := randomName("sdk-volume-2-")
	createResp2, err := client.CreateVolume(ctx, &VolumeCreateRequest{
		Name:       volumeName2,
		DatabaseID: databaseID,
		Comment:    "test volume 2",
	})
	require.NoError(t, err)
	require.NotZero(t, createResp2.VolumeID)

	// Cleanup
	defer func() {
		if _, err := client.DeleteVolume(ctx, &VolumeDeleteRequest{VolumeID: createResp1.VolumeID}); err != nil {
			t.Logf("cleanup delete volume 1 failed: %v", err)
		}
		if _, err := client.DeleteVolume(ctx, &VolumeDeleteRequest{VolumeID: createResp2.VolumeID}); err != nil {
			t.Logf("cleanup delete volume 2 failed: %v", err)
		}
	}()

	// Try to update first volume with the name of second volume
	_, err = client.UpdateVolume(ctx, &VolumeUpdateRequest{
		VolumeID: createResp1.VolumeID,
		Name:     volumeName2,
		Comment:  "updated comment",
	})
	require.Error(t, err)
	t.Logf("Expected error for duplicate name in update: %v", err)
}

func TestVolumeUpdateInvalidName(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	catalogID, markCatalogDeleted := createTestCatalog(t, client)
	databaseID, markDatabaseDeleted := createTestDatabase(t, client, catalogID)

	defer func() {
		markDatabaseDeleted()
		markCatalogDeleted()
	}()

	volumeName := randomName("sdk-volume-")
	createResp, err := client.CreateVolume(ctx, &VolumeCreateRequest{
		Name:       volumeName,
		DatabaseID: databaseID,
		Comment:    "test volume",
	})
	require.NoError(t, err)
	require.NotZero(t, createResp.VolumeID)

	// Cleanup
	defer func() {
		if _, err := client.DeleteVolume(ctx, &VolumeDeleteRequest{VolumeID: createResp.VolumeID}); err != nil {
			t.Logf("cleanup delete volume failed: %v", err)
		}
	}()

	// Try to update with invalid name
	_, err = client.UpdateVolume(ctx, &VolumeUpdateRequest{
		VolumeID: createResp.VolumeID,
		Name:     "v\"o'l2",
		Comment:  "test",
	})
	require.Error(t, err)
	t.Logf("Expected error for invalid name in update: %v", err)
}

func TestVolumeUpdateAndVerify(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	catalogID, markCatalogDeleted := createTestCatalog(t, client)
	databaseID, markDatabaseDeleted := createTestDatabase(t, client, catalogID)

	defer func() {
		markDatabaseDeleted()
		markCatalogDeleted()
	}()

	volumeName := randomName("sdk-volume-")
	createResp, err := client.CreateVolume(ctx, &VolumeCreateRequest{
		Name:       volumeName,
		DatabaseID: databaseID,
		Comment:    "test volume",
	})
	require.NoError(t, err)
	require.NotZero(t, createResp.VolumeID)

	// Cleanup
	defer func() {
		if _, err := client.DeleteVolume(ctx, &VolumeDeleteRequest{VolumeID: createResp.VolumeID}); err != nil {
			t.Logf("cleanup delete volume failed: %v", err)
		}
	}()

	// Update volume
	updatedName := randomName("sdk-volume-updated-")
	updatedComment := "updated comment"
	_, err = client.UpdateVolume(ctx, &VolumeUpdateRequest{
		VolumeID: createResp.VolumeID,
		Name:     updatedName,
		Comment:  updatedComment,
	})
	require.NoError(t, err)

	// Verify the update
	infoResp, err := client.GetVolume(ctx, &VolumeInfoRequest{VolumeID: createResp.VolumeID})
	require.NoError(t, err)
	require.Equal(t, updatedName, infoResp.VolumeName)
	require.Equal(t, updatedComment, infoResp.Comment)
}

func TestVolumeFullPath(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	catalogID, markCatalogDeleted := createTestCatalog(t, client)
	databaseID, markDatabaseDeleted := createTestDatabase(t, client, catalogID)

	defer func() {
		markDatabaseDeleted()
		markCatalogDeleted()
	}()

	volumeName := randomName("sdk-volume-")
	createResp, err := client.CreateVolume(ctx, &VolumeCreateRequest{
		Name:       volumeName,
		DatabaseID: databaseID,
		Comment:    "test volume",
	})
	require.NoError(t, err)
	require.NotZero(t, createResp.VolumeID)

	// Cleanup
	defer func() {
		if _, err := client.DeleteVolume(ctx, &VolumeDeleteRequest{VolumeID: createResp.VolumeID}); err != nil {
			t.Logf("cleanup delete volume failed: %v", err)
		}
	}()

	// Get full path
	fullPathResp, err := client.GetVolumeFullPath(ctx, &VolumeFullPathRequest{
		VolumeIDList: []VolumeID{createResp.VolumeID},
	})
	require.NoError(t, err)
	require.NotNil(t, fullPathResp)
	require.NotEmpty(t, fullPathResp.VolumeFullPath, "should have at least one path")

	// Verify path structure
	path := fullPathResp.VolumeFullPath[0]
	require.NotEmpty(t, path.NameList, "name list should not be empty")
	require.NotEmpty(t, path.IDList, "id list should not be empty")
	require.Equal(t, len(path.NameList), len(path.IDList), "name list and id list should have same length")

	t.Logf("Volume full path: Names=%v, IDs=%v", path.NameList, path.IDList)
}
