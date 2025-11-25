package sdk

import (
	"context"
)

// RunNL2SQL executes a natural language to SQL query.
//
// This method takes a natural language question and converts it to SQL,
// then executes the SQL query and returns the results.
//
// Example:
//
//	resp, err := client.RunNL2SQL(ctx, &sdk.NL2SQLRunSQLRequest{
//		Question: "Show me all users created in the last month",
//		DatabaseID: 123,
//	})
//	if err != nil {
//		return err
//	}
//	fmt.Printf("SQL: %s\n", resp.SQL)
//	fmt.Printf("Results: %v\n", resp.Results)
func (c *RawClient) RunNL2SQL(ctx context.Context, req *NL2SQLRunSQLRequest, opts ...CallOption) (*NL2SQLRunSQLResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp NL2SQLRunSQLResponse
	if err := c.postJSON(ctx, "/catalog/nl2sql/run_sql", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}
