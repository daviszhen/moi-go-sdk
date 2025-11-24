package sdk

import (
	"context"
)

// CreateKnowledge creates a new NL2SQL knowledge entry.
//
// Knowledge entries are used to provide context and examples for natural language
// to SQL conversion, improving the accuracy of NL2SQL queries.
//
// Example:
//
//	resp, err := client.CreateKnowledge(ctx, &sdk.NL2SQLKnowledgeCreateRequest{
//		Question: "What are the total sales?",
//		SQL:      "SELECT SUM(amount) FROM sales",
//		DatabaseID: 123,
//	})
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Created knowledge ID: %d\n", resp.KnowledgeID)
func (c *RawClient) CreateKnowledge(ctx context.Context, req *NL2SQLKnowledgeCreateRequest, opts ...CallOption) (*NL2SQLKnowledgeCreateResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp NL2SQLKnowledgeCreateResponse
	if err := c.postJSON(ctx, "/catalog/nl2sql_knowledge/create", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateKnowledge updates an existing NL2SQL knowledge entry.
//
// You can update the question, SQL, or other properties of the knowledge entry.
//
// Example:
//
//	resp, err := client.UpdateKnowledge(ctx, &sdk.NL2SQLKnowledgeUpdateRequest{
//		KnowledgeID: 456,
//		Question:    "Updated question",
//		SQL:         "SELECT * FROM updated_table",
//	})
func (c *RawClient) UpdateKnowledge(ctx context.Context, req *NL2SQLKnowledgeUpdateRequest, opts ...CallOption) (*NL2SQLKnowledgeUpdateResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp NL2SQLKnowledgeUpdateResponse
	if err := c.postJSON(ctx, "/catalog/nl2sql_knowledge/update", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteKnowledge deletes the specified NL2SQL knowledge entry.
//
// This operation permanently removes the knowledge entry.
//
// Example:
//
//	resp, err := client.DeleteKnowledge(ctx, &sdk.NL2SQLKnowledgeDeleteRequest{
//		KnowledgeID: 456,
//	})
func (c *RawClient) DeleteKnowledge(ctx context.Context, req *NL2SQLKnowledgeDeleteRequest, opts ...CallOption) (*NL2SQLKnowledgeDeleteResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp NL2SQLKnowledgeDeleteResponse
	if err := c.postJSON(ctx, "/catalog/nl2sql_knowledge/delete", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetKnowledge retrieves detailed information about the specified NL2SQL knowledge entry.
//
// The response includes the question, SQL, and other metadata.
//
// Example:
//
//	resp, err := client.GetKnowledge(ctx, &sdk.NL2SQLKnowledgeGetRequest{
//		KnowledgeID: 456,
//	})
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Question: %s\n", resp.Question)
//	fmt.Printf("SQL: %s\n", resp.SQL)
func (c *RawClient) GetKnowledge(ctx context.Context, req *NL2SQLKnowledgeGetRequest, opts ...CallOption) (*NL2SQLKnowledgeGetResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp NL2SQLKnowledgeGetResponse
	if err := c.postJSON(ctx, "/catalog/nl2sql_knowledge/get", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListKnowledge lists NL2SQL knowledge entries with optional filtering and pagination.
//
// Supports filtering by database ID and other criteria.
//
// Example:
//
//	resp, err := client.ListKnowledge(ctx, &sdk.NL2SQLKnowledgeListRequest{
//		DatabaseID: 123,
//		CommonCondition: sdk.CommonCondition{
//			Page:     1,
//			PageSize: 10,
//		},
//	})
//	if err != nil {
//		return err
//	}
//	for _, knowledge := range resp.List {
//		fmt.Printf("Knowledge: %s\n", knowledge.Question)
//	}
func (c *RawClient) ListKnowledge(ctx context.Context, req *NL2SQLKnowledgeListRequest, opts ...CallOption) (*NL2SQLKnowledgeListResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp NL2SQLKnowledgeListResponse
	if err := c.postJSON(ctx, "/catalog/nl2sql_knowledge/list", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SearchKnowledge searches NL2SQL knowledge entries by question or SQL.
//
// This is useful for finding similar knowledge entries that might help with
// a new NL2SQL query.
//
// Example:
//
//	resp, err := client.SearchKnowledge(ctx, &sdk.NL2SQLKnowledgeSearchRequest{
//		Query:      "total sales",
//		DatabaseID: 123,
//		Limit:      10,
//	})
//	if err != nil {
//		return err
//	}
//	for _, result := range resp.Results {
//		fmt.Printf("Match: %s (score: %f)\n", result.Question, result.Score)
//	}
func (c *RawClient) SearchKnowledge(ctx context.Context, req *NL2SQLKnowledgeSearchRequest, opts ...CallOption) (*NL2SQLKnowledgeSearchResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp NL2SQLKnowledgeSearchResponse
	if err := c.postJSON(ctx, "/catalog/nl2sql_knowledge/search", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}
