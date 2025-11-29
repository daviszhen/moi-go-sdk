package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// doLLMJSON issues a JSON request to LLM Proxy API and decodes the direct response (no envelope).
// LLM Proxy APIs return data directly or error in ErrorResponse format, not in envelope format.
func (c *RawClient) doLLMJSON(ctx context.Context, method, path string, body interface{}, respBody interface{}, opts ...CallOption) error {
	if c == nil {
		return fmt.Errorf("sdk client is nil")
	}
	callOpts := newCallOptions(opts...)

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		reader = bytes.NewReader(payload)
	}

	// Build full URL with /llm-proxy prefix
	fullPath := "/llm-proxy" + ensureLeadingSlash(path)
	resp, err := c.doRaw(ctx, method, fullPath, reader, callOpts, func(req *http.Request) {
		req.Header.Set(headerAccept, mimeJSON)
		if body != nil {
			req.Header.Set(headerContentType, mimeJSON)
		}
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read response body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	// Check for error response format
	if resp.StatusCode >= http.StatusBadRequest {
		var errResp struct {
			Error struct {
				Message string `json:"message"`
				Type    string `json:"type"`
				Code    string `json:"code"`
			} `json:"error"`
		}
		if err := json.Unmarshal(data, &errResp); err == nil && errResp.Error.Message != "" {
			return &APIError{
				Code:       errResp.Error.Code,
				Message:    errResp.Error.Message,
				HTTPStatus: resp.StatusCode,
			}
		}
		// If not in error format, return HTTP error
		return &HTTPError{StatusCode: resp.StatusCode, Body: data}
	}

	// Parse successful response
	if respBody != nil && len(data) > 0 && string(data) != "null" {
		if err := json.Unmarshal(data, respBody); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

// ============ Session Management APIs ============

// CreateLLMSession creates a new session in LLM Proxy.
//
// Example:
//
//	resp, err := client.CreateLLMSession(ctx, &sdk.LLMSessionCreateRequest{
//		Title:  "My Session",
//		Source: "my-app",
//		UserID: "user123",
//		Tags:   []string{"alpha", "beta"},
//	})
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Created session ID: %d\n", resp.ID)
func (c *RawClient) CreateLLMSession(ctx context.Context, req *LLMSessionCreateRequest, opts ...CallOption) (*LLMSession, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp LLMSession
	if err := c.doLLMJSON(ctx, http.MethodPost, "/api/sessions", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListLLMSessions lists sessions with optional filtering and pagination.
//
// Example:
//
//	resp, err := client.ListLLMSessions(ctx, &sdk.LLMSessionListRequest{
//		UserID:   "user123",
//		Source:   "my-app",
//		Page:     1,
//		PageSize: 20,
//	})
//	if err != nil {
//		return err
//	}
//	for _, session := range resp.Sessions {
//		fmt.Printf("Session: %s (ID: %d)\n", session.Title, session.ID)
//	}
func (c *RawClient) ListLLMSessions(ctx context.Context, req *LLMSessionListRequest, opts ...CallOption) (*LLMSessionListResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	// Build query parameters
	query := url.Values{}
	if req.UserID != "" {
		query.Set("user_id", req.UserID)
	}
	if req.Source != "" {
		query.Set("source", req.Source)
	}
	if req.Keyword != "" {
		query.Set("keyword", req.Keyword)
	}
	if len(req.Tags) > 0 {
		query.Set("tags", strings.Join(req.Tags, ","))
	}
	if req.Page > 0 {
		query.Set("page", strconv.Itoa(req.Page))
	}
	if req.PageSize > 0 {
		query.Set("page_size", strconv.Itoa(req.PageSize))
	}

	var resp LLMSessionListResponse
	path := "/api/sessions"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	if err := c.doLLMJSON(ctx, http.MethodGet, path, nil, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetLLMSession retrieves a single session by ID.
//
// Example:
//
//	session, err := client.GetLLMSession(ctx, 1)
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Session: %s\n", session.Title)
func (c *RawClient) GetLLMSession(ctx context.Context, sessionID int64, opts ...CallOption) (*LLMSession, error) {
	var resp LLMSession
	path := fmt.Sprintf("/api/sessions/%d", sessionID)
	if err := c.doLLMJSON(ctx, http.MethodGet, path, nil, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateLLMSession updates a session (supports partial updates).
//
// Example:
//
//	updated, err := client.UpdateLLMSession(ctx, 1, &sdk.LLMSessionUpdateRequest{
//		Title: stringPtr("Updated Title"),
//		Tags:  &[]string{"release"},
//	})
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Updated session: %s\n", updated.Title)
func (c *RawClient) UpdateLLMSession(ctx context.Context, sessionID int64, req *LLMSessionUpdateRequest, opts ...CallOption) (*LLMSession, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp LLMSession
	path := fmt.Sprintf("/api/sessions/%d", sessionID)
	if err := c.doLLMJSON(ctx, http.MethodPut, path, req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteLLMSession deletes a session.
//
// Example:
//
//	_, err := client.DeleteLLMSession(ctx, 1)
//	if err != nil {
//		return err
//	}
func (c *RawClient) DeleteLLMSession(ctx context.Context, sessionID int64, opts ...CallOption) (*LLMSessionDeleteResponse, error) {
	var resp LLMSessionDeleteResponse
	path := fmt.Sprintf("/api/sessions/%d", sessionID)
	if err := c.doLLMJSON(ctx, http.MethodDelete, path, nil, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListLLMSessionMessages lists messages for a specific session with optional filtering.
//
// Example:
//
//	messages, err := client.ListLLMSessionMessages(ctx, 1, &sdk.LLMSessionMessagesListRequest{
//		Role:   sdk.LLMMessageRoleUser,
//		Status: sdk.LLMMessageStatusSuccess,
//	})
//	if err != nil {
//		return err
//	}
//	for _, msg := range messages {
//		fmt.Printf("Message: %s\n", msg.Content)
//	}
func (c *RawClient) ListLLMSessionMessages(ctx context.Context, sessionID int64, req *LLMSessionMessagesListRequest, opts ...CallOption) ([]LLMChatMessage, error) {
	if req == nil {
		req = &LLMSessionMessagesListRequest{}
	}

	// Build query parameters
	query := url.Values{}
	if req.Source != "" {
		query.Set("source", req.Source)
	}
	if req.Role != "" {
		query.Set("role", string(req.Role))
	}
	if req.Status != "" {
		query.Set("status", string(req.Status))
	}
	if req.Model != "" {
		query.Set("model", req.Model)
	}

	var resp []LLMChatMessage
	path := fmt.Sprintf("/api/sessions/%d/messages", sessionID)
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	if err := c.doLLMJSON(ctx, http.MethodGet, path, nil, &resp, opts...); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetLLMSessionLatestCompletedMessage retrieves the latest completed message ID for a session.
//
// Example:
//
//	resp, err := client.GetLLMSessionLatestCompletedMessage(ctx, 1)
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Latest completed message ID: %d\n", resp.MessageID)
func (c *RawClient) GetLLMSessionLatestCompletedMessage(ctx context.Context, sessionID int64, opts ...CallOption) (*LLMLatestCompletedMessageResponse, error) {
	var resp LLMLatestCompletedMessageResponse
	path := fmt.Sprintf("/api/sessions/%d/messages/latest-completed", sessionID)
	if err := c.doLLMJSON(ctx, http.MethodGet, path, nil, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ============ Chat Message Management APIs ============

// CreateLLMChatMessage creates a new chat message record.
//
// Example:
//
//	msg, err := client.CreateLLMChatMessage(ctx, &sdk.LLMChatMessageCreateRequest{
//		UserID:   "user123",
//		Source:   "my-app",
//		Role:     sdk.LLMMessageRoleUser,
//		Content:  "Hello, world!",
//		Model:    "gpt-4",
//		Status:   sdk.LLMMessageStatusSuccess,
//		Tags:     []string{"tag1", "tag2"},
//	})
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Created message ID: %d\n", msg.ID)
func (c *RawClient) CreateLLMChatMessage(ctx context.Context, req *LLMChatMessageCreateRequest, opts ...CallOption) (*LLMChatMessage, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp LLMChatMessage
	if err := c.doLLMJSON(ctx, http.MethodPost, "/api/chat-messages", req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListLLMChatMessages lists chat messages with optional filtering and pagination.
//
// Example:
//
//	resp, err := client.ListLLMChatMessages(ctx, &sdk.LLMChatMessageListRequest{
//		UserID:   "user123",
//		SessionID: int64Ptr(1),
//		Page:     1,
//		PageSize: 20,
//	})
//	if err != nil {
//		return err
//	}
//	for _, msg := range resp.Messages {
//		fmt.Printf("Message: %s\n", msg.Content)
//	}
func (c *RawClient) ListLLMChatMessages(ctx context.Context, req *LLMChatMessageListRequest, opts ...CallOption) (*LLMChatMessageListResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	// Build query parameters
	query := url.Values{}
	if req.UserID != "" {
		query.Set("user_id", req.UserID)
	}
	if req.SessionID != nil {
		query.Set("session_id", strconv.FormatInt(*req.SessionID, 10))
	}
	if req.Source != "" {
		query.Set("source", req.Source)
	}
	if req.Role != "" {
		query.Set("role", string(req.Role))
	}
	if req.Status != "" {
		query.Set("status", string(req.Status))
	}
	if len(req.Tags) > 0 {
		query.Set("tags", strings.Join(req.Tags, ","))
	}
	if req.Page > 0 {
		query.Set("page", strconv.Itoa(req.Page))
	}
	if req.PageSize > 0 {
		query.Set("page_size", strconv.Itoa(req.PageSize))
	}

	var resp LLMChatMessageListResponse
	path := "/api/chat-messages"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	if err := c.doLLMJSON(ctx, http.MethodGet, path, nil, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetLLMChatMessage retrieves a single chat message by ID.
//
// Example:
//
//	msg, err := client.GetLLMChatMessage(ctx, 1)
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Message: %s\n", msg.Content)
func (c *RawClient) GetLLMChatMessage(ctx context.Context, messageID int64, opts ...CallOption) (*LLMChatMessage, error) {
	var resp LLMChatMessage
	path := fmt.Sprintf("/api/chat-messages/%d", messageID)
	if err := c.doLLMJSON(ctx, http.MethodGet, path, nil, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateLLMChatMessage updates a chat message.
//
// Example:
//
//	updated, err := client.UpdateLLMChatMessage(ctx, 1, &sdk.LLMChatMessageUpdateRequest{
//		Status:   statusPtr(sdk.LLMMessageStatusSuccess),
//		Response: stringPtr("Updated response"),
//	})
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Updated message: %s\n", updated.Content)
func (c *RawClient) UpdateLLMChatMessage(ctx context.Context, messageID int64, req *LLMChatMessageUpdateRequest, opts ...CallOption) (*LLMChatMessage, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp LLMChatMessage
	path := fmt.Sprintf("/api/chat-messages/%d", messageID)
	if err := c.doLLMJSON(ctx, http.MethodPut, path, req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteLLMChatMessage deletes a chat message.
//
// Example:
//
//	_, err := client.DeleteLLMChatMessage(ctx, 1)
//	if err != nil {
//		return err
//	}
func (c *RawClient) DeleteLLMChatMessage(ctx context.Context, messageID int64, opts ...CallOption) (*LLMChatMessageDeleteResponse, error) {
	var resp LLMChatMessageDeleteResponse
	path := fmt.Sprintf("/api/chat-messages/%d", messageID)
	if err := c.doLLMJSON(ctx, http.MethodDelete, path, nil, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateLLMChatMessageTags updates message tags (complete replacement).
//
// Example:
//
//	updated, err := client.UpdateLLMChatMessageTags(ctx, 1, &sdk.LLMChatMessageTagsUpdateRequest{
//		Tags: []string{"tag1", "tag2", "tag3"},
//	})
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Updated message with %d tags\n", len(updated.Tags))
func (c *RawClient) UpdateLLMChatMessageTags(ctx context.Context, messageID int64, req *LLMChatMessageTagsUpdateRequest, opts ...CallOption) (*LLMChatMessage, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	var resp LLMChatMessage
	path := fmt.Sprintf("/api/chat-messages/%d/tags", messageID)
	if err := c.doLLMJSON(ctx, http.MethodPut, path, req, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteLLMChatMessageTag deletes a single tag from a message.
//
// Example:
//
//	_, err := client.DeleteLLMChatMessageTag(ctx, 1, "my-app", "tag1")
//	if err != nil {
//		return err
//	}
func (c *RawClient) DeleteLLMChatMessageTag(ctx context.Context, messageID int64, source, name string, opts ...CallOption) (*LLMChatMessageTagDeleteResponse, error) {
	var resp LLMChatMessageTagDeleteResponse
	// URL encode source and name
	path := fmt.Sprintf("/api/chat-messages/%d/tags/%s/%s", messageID, url.PathEscape(source), url.PathEscape(name))
	if err := c.doLLMJSON(ctx, http.MethodDelete, path, nil, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Helper functions for pointer creation
// These are used in tests and example code to create pointer values for optional fields.
func stringPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}

func llmStatusPtr(s LLMMessageStatus) *LLMMessageStatus {
	return &s
}
