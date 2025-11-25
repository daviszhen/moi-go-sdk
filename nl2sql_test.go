package sdk

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNL2SQLRunSQL_NilRequest(t *testing.T) {
	client := newTestClient(t)
	_, err := client.RunNL2SQL(context.Background(), nil)
	require.ErrorIs(t, err, ErrNilRequest)
}

func TestNL2SQLRunSQL_ShowOperations(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	catalogName := randomName("sdk-nl2sql-cat-")
	catalogResp, err := client.CreateCatalog(ctx, &CatalogCreateRequest{
		CatalogName: catalogName,
		Comment:     "sdk nl2sql catalog",
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		if _, err := client.DeleteCatalog(ctx, &CatalogDeleteRequest{CatalogID: catalogResp.CatalogID}); err != nil {
			t.Logf("cleanup delete catalog failed: %v", err)
		}
	})

	databaseName := randomName("sdk_nl2sql_db_")
	dbResp, err := client.CreateDatabase(ctx, &DatabaseCreateRequest{
		CatalogID:    catalogResp.CatalogID,
		DatabaseName: databaseName,
		Comment:      "sdk nl2sql database",
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		if _, err := client.DeleteDatabase(ctx, &DatabaseDeleteRequest{DatabaseID: dbResp.DatabaseID}); err != nil {
			t.Logf("cleanup delete database failed: %v", err)
		}
	})

	tableName := randomName("sdk_nl2sql_table_")
	tableResp, err := client.CreateTable(ctx, &TableCreateRequest{
		DatabaseID: dbResp.DatabaseID,
		Name:       tableName,
		Comment:    "sdk nl2sql table",
		Columns: []Column{
			{Name: "id", Type: "INT", IsPk: true, Comment: "comment"},
			{Name: "name", Type: "VARCHAR(128)"},
			{Name: "age", Type: "INT", Default: "0"},
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		if _, err := client.DeleteTable(ctx, &TableDeleteRequest{TableID: tableResp.TableID}); err != nil {
			t.Logf("cleanup delete table failed: %v", err)
		}
	})

	tableInfo := []DbAndTablesInfo{
		{
			DbName:     databaseName,
			TableNames: []string{tableName},
		},
	}

	cases := []struct {
		name     string
		request  *NL2SQLRunSQLRequest
		validate func(t *testing.T, resp *NL2SQLRunSQLResponse)
	}{
		{
			name: "show_databases",
			request: &NL2SQLRunSQLRequest{
				Operation: ShowDatabases,
			},
			validate: func(t *testing.T, resp *NL2SQLRunSQLResponse) {
				require.NotEmpty(t, resp.Results)
				require.NotEmpty(t, resp.Results[0].Columns)
				require.Contains(t, resp.Results[0].Columns, "Database")
				requireRowContainsValue(t, resp.Results[0].Rows, databaseName)
			},
		},
		{
			name: "show_table",
			request: &NL2SQLRunSQLRequest{
				Operation: ShowTable,
				DbNames:   []string{databaseName},
			},
			validate: func(t *testing.T, resp *NL2SQLRunSQLResponse) {
				require.NotEmpty(t, resp.Results)
				require.NotEmpty(t, resp.Results[0].Columns)
				requireRowContainsValue(t, resp.Results[0].Rows, tableName)
			},
		},
		{
			name: "desc_table",
			request: &NL2SQLRunSQLRequest{
				Operation:  DescTable,
				TableNames: tableInfo,
			},
			validate: func(t *testing.T, resp *NL2SQLRunSQLResponse) {
				require.NotEmpty(t, resp.Results)
				require.Contains(t, resp.Results[0].Columns, "Field")
				require.Contains(t, resp.Results[0].Columns, "Type")
				requireRowContainsValue(t, resp.Results[0].Rows, "id")
				requireRowContainsValue(t, resp.Results[0].Rows, "comment")
			},
		},
		{
			name: "show_create_table",
			request: &NL2SQLRunSQLRequest{
				Operation:  ShowCreateTable,
				TableNames: tableInfo,
			},
			validate: func(t *testing.T, resp *NL2SQLRunSQLResponse) {
				require.NotEmpty(t, resp.Results)
				require.Contains(t, resp.Results[0].Columns, "Table")
				requireRowContainsValue(t, resp.Results[0].Rows, tableName)
			},
		},
		{
			name: "select_3",
			request: &NL2SQLRunSQLRequest{
				Operation:  Select_3,
				TableNames: tableInfo,
			},
			validate: func(t *testing.T, resp *NL2SQLRunSQLResponse) {
				require.NotEmpty(t, resp.Results)
				require.ElementsMatch(t, []string{"id", "name", "age"}, resp.Results[0].Columns)
				require.Empty(t, resp.Results[0].Rows)
			},
		},
		{
			name: "run_sql_select",
			request: &NL2SQLRunSQLRequest{
				Operation: RunSQL,
				Statement: fmt.Sprintf("select * from `%s`.`%s`", databaseName, tableName),
			},
			validate: func(t *testing.T, resp *NL2SQLRunSQLResponse) {
				require.NotEmpty(t, resp.Results)
				require.Equal(t, []string{"id", "name", "age"}, resp.Results[0].Columns)
				require.Empty(t, resp.Results[0].Rows)
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			resp, err := client.RunNL2SQL(ctx, tc.request)
			require.NoError(t, err)
			tc.validate(t, resp)
		})
	}
}

func TestNL2SQLRunSQL_InvalidStatement(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()
	_, err := client.RunNL2SQL(ctx, &NL2SQLRunSQLRequest{
		Operation: RunSQL,
		Statement: "drop table moi.catalog_database",
	})
	require.Error(t, err)
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	require.NotEmpty(t, apiErr.Code)
}

func requireRowContainsValue(t *testing.T, rows []NL2SQLRow, value string) {
	t.Helper()
	for _, row := range rows {
		for _, cell := range row {
			if cell == value {
				return
			}
		}
	}
	t.Fatalf("value %q not found in rows %v", value, rows)
}
