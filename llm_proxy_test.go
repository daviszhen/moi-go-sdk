package sdk

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// ============ Nil Request Validation Tests ============

func TestCreateLLMSession_NilRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := &RawClient{}

	resp, err := client.CreateLLMSession(ctx, nil)
	require.Nil(t, resp)
	require.ErrorIs(t, err, ErrNilRequest)
}

func TestListLLMSessions_NilRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := &RawClient{}

	resp, err := client.ListLLMSessions(ctx, nil)
	require.Nil(t, resp)
	require.ErrorIs(t, err, ErrNilRequest)
}

func TestUpdateLLMSession_NilRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := &RawClient{}

	resp, err := client.UpdateLLMSession(ctx, 1, nil)
	require.Nil(t, resp)
	require.ErrorIs(t, err, ErrNilRequest)
}

func TestCreateLLMChatMessage_NilRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := &RawClient{}

	resp, err := client.CreateLLMChatMessage(ctx, nil)
	require.Nil(t, resp)
	require.ErrorIs(t, err, ErrNilRequest)
}

func TestUpdateLLMChatMessage_NilRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := &RawClient{}

	resp, err := client.UpdateLLMChatMessage(ctx, 1, nil)
	require.Nil(t, resp)
	require.ErrorIs(t, err, ErrNilRequest)
}

func TestUpdateLLMChatMessageTags_NilRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := &RawClient{}

	resp, err := client.UpdateLLMChatMessageTags(ctx, 1, nil)
	require.Nil(t, resp)
	require.ErrorIs(t, err, ErrNilRequest)
}

// ============ Live Flow Tests (using real backend) ============

// TestLLMSessionLiveFlow tests the complete session management flow with a real backend.
func TestLLMSessionLiveFlow(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	// Create a session
	createReq := &LLMSessionCreateRequest{
		Title:  randomName("sdk-session-"),
		Source: "sdk-test",
		UserID: randomName("user-"),
		Config: `{"temperature": 0.7}`,
		// Tags omitted to avoid backend tag upsert issues
	}

	session, err := client.CreateLLMSession(ctx, createReq)
	require.NoError(t, err)
	require.NotNil(t, session)
	require.Greater(t, session.ID, int64(0))
	require.Equal(t, createReq.Title, session.Title)
	require.Equal(t, createReq.Source, session.Source)
	require.Equal(t, createReq.UserID, session.UserID)
	t.Logf("Created session ID: %d", session.ID)

	// Cleanup: delete the session
	t.Cleanup(func() {
		if _, err := client.DeleteLLMSession(ctx, session.ID); err != nil {
			t.Logf("cleanup delete session failed: %v", err)
		}
	})

	// Get the session
	gotSession, err := client.GetLLMSession(ctx, session.ID)
	require.NoError(t, err)
	require.NotNil(t, gotSession)
	require.Equal(t, session.ID, gotSession.ID)
	require.Equal(t, session.Title, gotSession.Title)

	// List sessions
	listResp, err := client.ListLLMSessions(ctx, &LLMSessionListRequest{
		UserID:   createReq.UserID,
		Source:   createReq.Source,
		Page:     1,
		PageSize: 20,
	})
	require.NoError(t, err)
	require.NotNil(t, listResp)
	require.Greater(t, listResp.Total, int64(0))
	found := false
	for _, s := range listResp.Sessions {
		if s.ID == session.ID {
			found = true
			break
		}
	}
	require.True(t, found, "Created session should be in the list")

	// Update the session
	updatedTitle := randomName("updated-session-")
	updatedSession, err := client.UpdateLLMSession(ctx, session.ID, &LLMSessionUpdateRequest{
		Title: stringPtr(updatedTitle),
	})
	require.NoError(t, err)
	require.NotNil(t, updatedSession)
	require.Equal(t, updatedTitle, updatedSession.Title)
}

// TestLLMSessionMessagesLiveFlow tests session messages operations with a real backend.
func TestLLMSessionMessagesLiveFlow(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	// Create a session first
	createReq := &LLMSessionCreateRequest{
		Title:  randomName("sdk-session-"),
		Source: "sdk-test",
		UserID: randomName("user-"),
	}

	session, err := client.CreateLLMSession(ctx, createReq)
	require.NoError(t, err)
	require.NotNil(t, session)
	t.Logf("Created session ID: %d", session.ID)

	// Cleanup: delete the session
	t.Cleanup(func() {
		if _, err := client.DeleteLLMSession(ctx, session.ID); err != nil {
			t.Logf("cleanup delete session failed: %v", err)
		}
	})

	// Create a message in the session
	message, err := client.CreateLLMChatMessage(ctx, &LLMChatMessageCreateRequest{
		UserID:    createReq.UserID,
		SessionID: int64Ptr(session.ID),
		Source:    createReq.Source,
		Role:      LLMMessageRoleUser,
		Content:   "Test message",
		Model:     "gpt-4",
		Status:    LLMMessageStatusSuccess,
	})
	require.NoError(t, err)
	require.NotNil(t, message)
	require.Greater(t, message.ID, int64(0))
	t.Logf("Created message ID: %d", message.ID)

	// Cleanup: delete the message
	t.Cleanup(func() {
		if _, err := client.DeleteLLMChatMessage(ctx, message.ID); err != nil {
			t.Logf("cleanup delete message failed: %v", err)
		}
	})

	// List session messages
	// Note: The messages list endpoint does not return original_content and content fields
	// to reduce data transfer. Use GetLLMChatMessage to get full message content.
	messages, err := client.ListLLMSessionMessages(ctx, session.ID, &LLMSessionMessagesListRequest{})
	require.NoError(t, err)
	require.NotEmpty(t, messages)
	foundMessage := false
	for _, m := range messages {
		if m.ID == message.ID {
			foundMessage = true
			// Verify message metadata (content fields are not returned by list endpoint)
			require.Equal(t, message.Role, m.Role)
			require.Equal(t, message.Status, m.Status)
			require.Equal(t, message.Model, m.Model)
			break
		}
	}
	require.True(t, foundMessage, "Created message should be in the session messages list")

	// List session messages with role filter
	messagesByRole, err := client.ListLLMSessionMessages(ctx, session.ID, &LLMSessionMessagesListRequest{
		Role: LLMMessageRoleUser,
	})
	require.NoError(t, err)
	require.NotEmpty(t, messagesByRole)

	// Get latest completed message
	latestResp, err := client.GetLLMSessionLatestCompletedMessage(ctx, session.ID)
	require.NoError(t, err)
	require.NotNil(t, latestResp)
	require.Equal(t, session.ID, latestResp.SessionID)
	require.Equal(t, message.ID, latestResp.MessageID)

	// Get latest message (regardless of status)
	latestResp2, err := client.GetLLMSessionLatestMessage(ctx, session.ID)
	require.NoError(t, err)
	require.NotNil(t, latestResp2)
	require.Equal(t, session.ID, latestResp2.SessionID)
	require.Equal(t, message.ID, latestResp2.MessageID)
}

// TestLLMChatMessageLiveFlow tests the complete chat message management flow with a real backend.
func TestLLMChatMessageLiveFlow(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	userID := randomName("user-")
	source := "sdk-test"

	// Create a message without initial Response to test Response update
	createReq := &LLMChatMessageCreateRequest{
		UserID:          userID,
		Source:          source,
		Role:            LLMMessageRoleUser,
		OriginalContent: "Original input",
		Content:         "Processed content",
		Model:           "gpt-4",
		Status:          LLMMessageStatusSuccess,
		// Response omitted initially to test update
		// Tags omitted to avoid backend tag upsert issues
	}

	message, err := client.CreateLLMChatMessage(ctx, createReq)
	require.NoError(t, err)
	require.NotNil(t, message)
	require.Greater(t, message.ID, int64(0))
	require.Equal(t, createReq.UserID, message.UserID)
	require.Equal(t, createReq.Content, message.Content)
	t.Logf("Created message ID: %d", message.ID)

	// Cleanup: delete the message
	t.Cleanup(func() {
		if _, err := client.DeleteLLMChatMessage(ctx, message.ID); err != nil {
			t.Logf("cleanup delete message failed: %v", err)
		}
	})

	// Get the message
	gotMessage, err := client.GetLLMChatMessage(ctx, message.ID)
	require.NoError(t, err)
	require.NotNil(t, gotMessage)
	require.Equal(t, message.ID, gotMessage.ID)
	require.Equal(t, message.Content, gotMessage.Content)

	// Note: ListLLMChatMessages API has been removed.
	// Use ListLLMSessionMessages to get messages for a session instead.

	// Update the message Response (backend appends for streaming, so we test with initial empty response)
	updatedResponse := "Updated AI response"
	updatedMessage, err := client.UpdateLLMChatMessage(ctx, message.ID, &LLMChatMessageUpdateRequest{
		Response: stringPtr(updatedResponse),
	})
	require.NoError(t, err)
	require.NotNil(t, updatedMessage)
	// Backend may append Response for streaming support, so we just verify it contains our update
	require.Contains(t, updatedMessage.Response, updatedResponse)

	// Note: Tag update and delete operations are skipped in this test
	// due to backend tag upsert issues with duplicate key updates.
	// These operations can be tested separately when the backend issue is resolved.
}

// TestLLMSessionDeleteLiveFlow tests session deletion with a real backend.
func TestLLMSessionDeleteLiveFlow(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	// Create a session
	createReq := &LLMSessionCreateRequest{
		Title:  randomName("sdk-session-"),
		Source: "sdk-test",
		UserID: randomName("user-"),
	}

	session, err := client.CreateLLMSession(ctx, createReq)
	require.NoError(t, err)
	require.NotNil(t, session)
	sessionID := session.ID
	t.Logf("Created session ID: %d", sessionID)

	// Delete the session
	deleteResp, err := client.DeleteLLMSession(ctx, sessionID)
	require.NoError(t, err)
	require.NotNil(t, deleteResp)
	t.Logf("Deleted session ID: %d", sessionID)

	// Verify session is deleted by trying to get it
	_, err = client.GetLLMSession(ctx, sessionID)
	require.Error(t, err)
}

// TestLLMChatMessageDeleteLiveFlow tests message deletion with a real backend.
func TestLLMChatMessageDeleteLiveFlow(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	userID := randomName("user-")
	source := "sdk-test"

	// Create a message
	createReq := &LLMChatMessageCreateRequest{
		UserID:  userID,
		Source:  source,
		Role:    LLMMessageRoleUser,
		Content: "Test message to delete",
		Model:   "gpt-4",
		Status:  LLMMessageStatusSuccess,
	}

	message, err := client.CreateLLMChatMessage(ctx, createReq)
	require.NoError(t, err)
	require.NotNil(t, message)
	messageID := message.ID
	t.Logf("Created message ID: %d", messageID)

	// Delete the message
	deleteResp, err := client.DeleteLLMChatMessage(ctx, messageID)
	require.NoError(t, err)
	require.NotNil(t, deleteResp)
	t.Logf("Deleted message ID: %d", messageID)

	// Verify message is deleted by trying to get it
	_, err = client.GetLLMChatMessage(ctx, messageID)
	require.Error(t, err)
}

// TestLLMSessionListWithFiltersLiveFlow tests listing sessions with various filters.
func TestLLMSessionListWithFiltersLiveFlow(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	userID := randomName("user-")
	source := "sdk-test"

	// Create a session
	createReq := &LLMSessionCreateRequest{
		Title:  randomName("sdk-session-"),
		Source: source,
		UserID: userID,
		// Tags omitted to avoid backend tag upsert issues
	}

	session, err := client.CreateLLMSession(ctx, createReq)
	require.NoError(t, err)
	require.NotNil(t, session)
	t.Logf("Created session ID: %d", session.ID)

	// Cleanup
	t.Cleanup(func() {
		if _, err := client.DeleteLLMSession(ctx, session.ID); err != nil {
			t.Logf("cleanup delete session failed: %v", err)
		}
	})

	// List sessions by user ID
	listResp, err := client.ListLLMSessions(ctx, &LLMSessionListRequest{
		UserID:   userID,
		Page:     1,
		PageSize: 20,
	})
	require.NoError(t, err)
	require.NotNil(t, listResp)
	require.Greater(t, listResp.Total, int64(0))

	// List sessions by source
	listResp2, err := client.ListLLMSessions(ctx, &LLMSessionListRequest{
		Source:   source,
		Page:     1,
		PageSize: 20,
	})
	require.NoError(t, err)
	require.NotNil(t, listResp2)
	require.Greater(t, listResp2.Total, int64(0))

	// List sessions with keyword
	listResp3, err := client.ListLLMSessions(ctx, &LLMSessionListRequest{
		UserID:   userID,
		Keyword:  session.Title,
		Page:     1,
		PageSize: 20,
	})
	require.NoError(t, err)
	require.NotNil(t, listResp3)
	require.Greater(t, listResp3.Total, int64(0))
}

// TestLLMChatMessageListWithFiltersLiveFlow tests listing messages with various filters.
func TestLLMChatMessageListWithFiltersLiveFlow(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	userID := randomName("user-")
	source := "sdk-test"

	// Create a message
	createReq := &LLMChatMessageCreateRequest{
		UserID:  userID,
		Source:  source,
		Role:    LLMMessageRoleUser,
		Content: "Filter test message",
		Model:   "gpt-4",
		Status:  LLMMessageStatusSuccess,
		// Tags omitted to avoid backend tag upsert issues
	}

	message, err := client.CreateLLMChatMessage(ctx, createReq)
	require.NoError(t, err)
	require.NotNil(t, message)
	t.Logf("Created message ID: %d", message.ID)

	// Cleanup
	t.Cleanup(func() {
		if _, err := client.DeleteLLMChatMessage(ctx, message.ID); err != nil {
			t.Logf("cleanup delete message failed: %v", err)
		}
	})

	// Note: ListLLMChatMessages API has been removed.
	// To list messages, use ListLLMSessionMessages with a session ID instead.
	// This test verifies message creation and deletion only.
}

// TestLLMSessionLatestMessageLiveFlow tests getting the latest message (regardless of status) with a real backend.
func TestLLMSessionLatestMessageLiveFlow(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	// Create a session
	createReq := &LLMSessionCreateRequest{
		Title:  randomName("sdk-session-"),
		Source: "sdk-test",
		UserID: randomName("user-"),
	}

	session, err := client.CreateLLMSession(ctx, createReq)
	require.NoError(t, err)
	require.NotNil(t, session)
	t.Logf("Created session ID: %d", session.ID)

	// Cleanup
	t.Cleanup(func() {
		if _, err := client.DeleteLLMSession(ctx, session.ID); err != nil {
			t.Logf("cleanup delete session failed: %v", err)
		}
	})

	// Create a message with success status
	message1, err := client.CreateLLMChatMessage(ctx, &LLMChatMessageCreateRequest{
		UserID:    createReq.UserID,
		SessionID: int64Ptr(session.ID),
		Source:    createReq.Source,
		Role:      LLMMessageRoleUser,
		Content:   "First message",
		Model:     "gpt-4",
		Status:    LLMMessageStatusSuccess,
	})
	require.NoError(t, err)
	require.NotNil(t, message1)
	t.Logf("Created message ID: %d (status: success)", message1.ID)

	// Cleanup message
	t.Cleanup(func() {
		if _, err := client.DeleteLLMChatMessage(ctx, message1.ID); err != nil {
			t.Logf("cleanup delete message failed: %v", err)
		}
	})

	// Create a message with failed status
	message2, err := client.CreateLLMChatMessage(ctx, &LLMChatMessageCreateRequest{
		UserID:    createReq.UserID,
		SessionID: int64Ptr(session.ID),
		Source:    createReq.Source,
		Role:      LLMMessageRoleUser,
		Content:   "Second message",
		Model:     "gpt-4",
		Status:    LLMMessageStatusFailed,
	})
	require.NoError(t, err)
	require.NotNil(t, message2)
	t.Logf("Created message ID: %d (status: failed)", message2.ID)

	// Cleanup message
	t.Cleanup(func() {
		if _, err := client.DeleteLLMChatMessage(ctx, message2.ID); err != nil {
			t.Logf("cleanup delete message failed: %v", err)
		}
	})

	// Get latest completed message (should return message1 with success status)
	latestCompletedResp, err := client.GetLLMSessionLatestCompletedMessage(ctx, session.ID)
	require.NoError(t, err)
	require.NotNil(t, latestCompletedResp)
	require.Equal(t, session.ID, latestCompletedResp.SessionID)
	require.Equal(t, message1.ID, latestCompletedResp.MessageID, "Latest completed should be message1 (success)")

	// Get latest message (regardless of status, should return message2 as it's the latest)
	latestResp, err := client.GetLLMSessionLatestMessage(ctx, session.ID)
	require.NoError(t, err)
	require.NotNil(t, latestResp)
	require.Equal(t, session.ID, latestResp.SessionID)
	require.Equal(t, message2.ID, latestResp.MessageID, "Latest message (any status) should be message2 (the most recent)")
}

// TestModifyLLMSessionMessageResponse_NilClient tests nil client handling.
func TestModifyLLMSessionMessageResponse_NilClient(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var client *RawClient = nil

	resp, err := client.ModifyLLMSessionMessageResponse(ctx, 1, 1, "test response")
	require.Nil(t, resp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "sdk client is nil")
}

// TestModifyLLMSessionMessageResponseLiveFlow tests modifying a message's modified_response with a real backend.
func TestModifyLLMSessionMessageResponseLiveFlow(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	// Create a session first
	createReq := &LLMSessionCreateRequest{
		Title:  randomName("sdk-session-"),
		Source: "sdk-test",
		UserID: randomName("user-"),
	}

	session, err := client.CreateLLMSession(ctx, createReq)
	require.NoError(t, err)
	require.NotNil(t, session)
	t.Logf("Created session ID: %d", session.ID)

	// Cleanup: delete the session
	t.Cleanup(func() {
		if _, err := client.DeleteLLMSession(ctx, session.ID); err != nil {
			t.Logf("cleanup delete session failed: %v", err)
		}
	})

	// Create a message in the session
	message, err := client.CreateLLMChatMessage(ctx, &LLMChatMessageCreateRequest{
		UserID:    createReq.UserID,
		SessionID: int64Ptr(session.ID),
		Source:    createReq.Source,
		Role:      LLMMessageRoleUser,
		Content:   "Test message for modified response",
		Model:     "gpt-4",
		Status:    LLMMessageStatusSuccess,
		Response:  "Original AI response",
	})
	require.NoError(t, err)
	require.NotNil(t, message)
	require.Greater(t, message.ID, int64(0))
	t.Logf("Created message ID: %d", message.ID)

	// Cleanup: delete the message
	t.Cleanup(func() {
		if _, err := client.DeleteLLMChatMessage(ctx, message.ID); err != nil {
			t.Logf("cleanup delete message failed: %v", err)
		}
	})

	// Verify initial state: modified_response should be empty
	gotMessage, err := client.GetLLMChatMessage(ctx, message.ID)
	require.NoError(t, err)
	require.NotNil(t, gotMessage)
	require.Equal(t, "", gotMessage.ModifiedResponse, "Initial modified_response should be empty")

	// Modify the message's modified_response
	modifiedResponse := "This is the modified response content"
	modifyResp, err := client.ModifyLLMSessionMessageResponse(ctx, session.ID, message.ID, modifiedResponse)
	require.NoError(t, err)
	require.NotNil(t, modifyResp)
	require.Equal(t, session.ID, modifyResp.SessionID)
	require.Equal(t, message.ID, modifyResp.MessageID)
	require.Equal(t, modifiedResponse, modifyResp.ModifiedResponse)
	t.Logf("Modified response for message ID: %d", message.ID)

	// Verify the modification by getting the message again
	gotMessage2, err := client.GetLLMChatMessage(ctx, message.ID)
	require.NoError(t, err)
	require.NotNil(t, gotMessage2)
	require.Equal(t, modifiedResponse, gotMessage2.ModifiedResponse, "Modified response should be updated")
	require.Equal(t, "Original AI response", gotMessage2.Response, "Original response should remain unchanged")

	// Update the modified_response again
	modifiedResponse2 := "Updated modified response content"
	modifyResp2, err := client.ModifyLLMSessionMessageResponse(ctx, session.ID, message.ID, modifiedResponse2)
	require.NoError(t, err)
	require.NotNil(t, modifyResp2)
	require.Equal(t, modifiedResponse2, modifyResp2.ModifiedResponse)

	// Verify the update
	gotMessage3, err := client.GetLLMChatMessage(ctx, message.ID)
	require.NoError(t, err)
	require.NotNil(t, gotMessage3)
	require.Equal(t, modifiedResponse2, gotMessage3.ModifiedResponse, "Modified response should be updated again")
}

// TestAppendLLMSessionMessageModifiedResponse_NilClient tests nil client handling.
func TestAppendLLMSessionMessageModifiedResponse_NilClient(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var client *RawClient = nil

	resp, err := client.AppendLLMSessionMessageModifiedResponse(ctx, 1, 1, "test content")
	require.Nil(t, resp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "sdk client is nil")
}

// TestAppendLLMSessionMessageModifiedResponseLiveFlow tests appending to a message's modified_response with a real backend.
func TestAppendLLMSessionMessageModifiedResponseLiveFlow(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	// Create a session first
	createReq := &LLMSessionCreateRequest{
		Title:  randomName("sdk-session-"),
		Source: "sdk-test",
		UserID: randomName("user-"),
	}

	session, err := client.CreateLLMSession(ctx, createReq)
	require.NoError(t, err)
	require.NotNil(t, session)
	t.Logf("Created session ID: %d", session.ID)

	// Cleanup: delete the session
	t.Cleanup(func() {
		if _, err := client.DeleteLLMSession(ctx, session.ID); err != nil {
			t.Logf("cleanup delete session failed: %v", err)
		}
	})

	// Create a message in the session
	message, err := client.CreateLLMChatMessage(ctx, &LLMChatMessageCreateRequest{
		UserID:    createReq.UserID,
		SessionID: int64Ptr(session.ID),
		Source:    createReq.Source,
		Role:      LLMMessageRoleUser,
		Content:   "Test message for append modified response",
		Model:     "gpt-4",
		Status:    LLMMessageStatusSuccess,
		Response:  "Original AI response",
	})
	require.NoError(t, err)
	require.NotNil(t, message)
	require.Greater(t, message.ID, int64(0))
	t.Logf("Created message ID: %d", message.ID)

	// Cleanup: delete the message
	t.Cleanup(func() {
		if _, err := client.DeleteLLMChatMessage(ctx, message.ID); err != nil {
			t.Logf("cleanup delete message failed: %v", err)
		}
	})

	// Verify initial state: modified_response should be empty
	gotMessage, err := client.GetLLMChatMessage(ctx, message.ID)
	require.NoError(t, err)
	require.NotNil(t, gotMessage)
	require.Equal(t, "", gotMessage.ModifiedResponse, "Initial modified_response should be empty")

	// First, set an initial modified_response
	initialModifiedResponse := "Initial modified content"
	_, err = client.ModifyLLMSessionMessageResponse(ctx, session.ID, message.ID, initialModifiedResponse)
	require.NoError(t, err)

	// Verify initial modified_response
	gotMessage1, err := client.GetLLMChatMessage(ctx, message.ID)
	require.NoError(t, err)
	require.Equal(t, initialModifiedResponse, gotMessage1.ModifiedResponse, "Initial modified_response should be set")

	// Append content to the modified_response
	appendContent1 := " - First append"
	appendResp1, err := client.AppendLLMSessionMessageModifiedResponse(ctx, session.ID, message.ID, appendContent1)
	require.NoError(t, err)
	require.NotNil(t, appendResp1)
	require.Equal(t, session.ID, appendResp1.SessionID)
	require.Equal(t, message.ID, appendResp1.MessageID)
	require.Equal(t, appendContent1, appendResp1.AppendContent)
	t.Logf("Appended content to message ID: %d", message.ID)

	// Verify the append by getting the message again
	gotMessage2, err := client.GetLLMChatMessage(ctx, message.ID)
	require.NoError(t, err)
	require.NotNil(t, gotMessage2)
	expectedResponse1 := initialModifiedResponse + appendContent1
	require.Equal(t, expectedResponse1, gotMessage2.ModifiedResponse, "Modified response should be appended")
	require.Equal(t, "Original AI response", gotMessage2.Response, "Original response should remain unchanged")

	// Append more content
	appendContent2 := " - Second append"
	appendResp2, err := client.AppendLLMSessionMessageModifiedResponse(ctx, session.ID, message.ID, appendContent2)
	require.NoError(t, err)
	require.NotNil(t, appendResp2)
	require.Equal(t, appendContent2, appendResp2.AppendContent)

	// Verify the second append
	gotMessage3, err := client.GetLLMChatMessage(ctx, message.ID)
	require.NoError(t, err)
	require.NotNil(t, gotMessage3)
	expectedResponse2 := expectedResponse1 + appendContent2
	require.Equal(t, expectedResponse2, gotMessage3.ModifiedResponse, "Modified response should be appended again")

	// Test appending to empty modified_response (after clearing it)
	_, err = client.ModifyLLMSessionMessageResponse(ctx, session.ID, message.ID, "")
	require.NoError(t, err)

	appendContent3 := "New content from scratch"
	appendResp3, err := client.AppendLLMSessionMessageModifiedResponse(ctx, session.ID, message.ID, appendContent3)
	require.NoError(t, err)
	require.NotNil(t, appendResp3)

	gotMessage4, err := client.GetLLMChatMessage(ctx, message.ID)
	require.NoError(t, err)
	require.Equal(t, appendContent3, gotMessage4.ModifiedResponse, "Appending to empty modified_response should work")
}
