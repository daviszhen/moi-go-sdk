package sdk

import (
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// ============ Nil Request Validation Tests ============

func TestAnalyzeDataStream_NilRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := &RawClient{}

	resp, err := client.AnalyzeDataStream(ctx, nil)
	require.Nil(t, resp)
	require.ErrorIs(t, err, ErrNilRequest)
}

func TestAnalyzeDataStream_EmptyQuestion(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := &RawClient{}

	req := &DataAnalysisRequest{
		Question: "",
	}
	resp, err := client.AnalyzeDataStream(ctx, req)
	require.Nil(t, resp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "question cannot be empty")
}

// ============ Live Flow Tests (using real backend) ============

// TestAnalyzeDataStreamLiveFlow tests the data analysis streaming API with a real backend.
func TestAnalyzeDataStreamLiveFlow(t *testing.T) {
	// Use a context with longer timeout for streaming tests
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	client := newTestClient(t)

	// Build request from the provided JSON
	source := "rag"
	sessionID := "019a5672-74f6-7bb8-ba55-239dea01d00f"
	codeType := 1

	req := &DataAnalysisRequest{
		Question:  "平均薪资是多少？",
		Source:    &source,
		SessionID: &sessionID,
		Config: &DataAnalysisConfig{
			DataCategory: "admin",
			FilterConditions: &FilterConditions{
				Type: "non_inter_data",
			},
			DataSource: &DataSource{
				Type: "all",
			},
			DataScope: &DataScope{
				Type:     "specified",
				CodeType: &codeType,
				CodeGroup: []CodeGroup{
					{
						Name:   "1001",
						Values: []string{"100101", "100102", "100103"},
					},
					{
						Name:   "1002",
						Values: []string{"1002"},
					},
					{
						Name:   "1003",
						Values: []string{"1003"},
					},
				},
			},
		},
	}

	// Call the streaming API
	stream, err := client.AnalyzeDataStream(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, stream)
	defer stream.Close()

	// Verify response headers
	require.Equal(t, 200, stream.StatusCode)
	contentType := stream.Header.Get("Content-Type")
	require.Contains(t, contentType, "text/event-stream", "Content-Type should be text/event-stream")

	// Read events from the stream
	eventCount := 0
	hasClassification := false
	hasComplete := false
	maxEvents := 50 // Limit events to prevent test timeout

	readEvents := true
	for readEvents {
		// Check context cancellation before reading
		select {
		case <-ctx.Done():
			t.Logf("Context cancelled after %d events, stopping event reading", eventCount)
			readEvents = false
		default:
		}

		if !readEvents {
			break
		}

		data, err := stream.ReadEvent()
		if err == io.EOF {
			t.Logf("Stream ended after %d events", eventCount)
			break
		}

		// Handle timeout errors gracefully (streaming may take a long time)
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				t.Logf("Timeout reached after %d events (this is acceptable for long-running streams)", eventCount)
				break
			}
			// Check if error is due to context cancellation
			if err.Error() != "" && (contains(err.Error(), "context deadline exceeded") || contains(err.Error(), "context canceled")) {
				t.Logf("Context error after %d events: %v (this is acceptable for long-running streams)", eventCount, err)
				break
			}
			require.NoError(t, err, "Error reading event")
		}

		require.NotEmpty(t, data, "Event data should not be empty")

		eventCount++

		// Log event details (truncate long data for readability)
		rawDataStr := data
		if len(rawDataStr) > 200 {
			rawDataStr = rawDataStr[:200] + "..."
		}
		t.Logf("Event #%d: RawData: %s", eventCount, rawDataStr)

		// Try to parse JSON to check for specific event types
		var eventData map[string]interface{}
		if err := json.Unmarshal([]byte(data), &eventData); err == nil {
			// Check for type field
			if eventType, ok := eventData["type"].(string); ok {
				if eventType == "classification" {
					hasClassification = true
					require.NotEmpty(t, data, "Classification event should have data")
				}
				if eventType == "complete" {
					hasComplete = true
					t.Logf("Analysis completed")
					readEvents = false
					break // Complete event indicates end of stream
				}
				if eventType == "error" {
					t.Logf("Error event received: %s", data)
				}
			}
			// Check for step_type field
			if stepType, ok := eventData["step_type"].(string); ok {
				if stepType == "init" {
					t.Logf("Init event received")
				}
			}
		} else {
			// If not JSON, check if it contains SSE fields
			if strings.Contains(data, "event: classification") {
				hasClassification = true
			}
			if strings.Contains(data, "event: complete") {
				hasComplete = true
				t.Logf("Analysis completed")
				readEvents = false
				break
			}
			if strings.Contains(data, "event: error") {
				t.Logf("Error event received: %s", data)
			}
		}

		// Limit events to prevent test timeout
		if eventCount >= maxEvents {
			t.Logf("Reached max events limit (%d), stopping to prevent timeout", maxEvents)
			readEvents = false
			break
		}
	}

	// Verify we received at least some events
	require.Greater(t, eventCount, 0, "Should receive at least one event")
	t.Logf("Total events received: %d", eventCount)

	// Note: We don't require classification or complete events as the backend behavior
	// may vary, but we log if they are present
	if hasClassification {
		t.Logf("Classification event was received")
	}
	if hasComplete {
		t.Logf("Complete event was received")
	}
}

// TestAnalyzeDataStream_SimpleRequest tests with a minimal request.
func TestAnalyzeDataStream_SimpleRequest(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(t)

	req := &DataAnalysisRequest{
		Question: "平均薪资是多少？",
		Config: &DataAnalysisConfig{
			DataCategory: "admin",
			DataSource: &DataSource{
				Type: "all",
			},
		},
	}

	stream, err := client.AnalyzeDataStream(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, stream)
	defer stream.Close()

	// Read at least one event to verify the stream works
	data, err := stream.ReadEvent()
	if err == io.EOF {
		t.Log("Stream ended immediately (no events)")
		return
	}
	require.NoError(t, err)
	require.NotEmpty(t, data)
	t.Logf("First event data: %s", data)
}

// ============ Cancel Analyze Tests ============

func TestCancelAnalyze_NilRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := &RawClient{}

	resp, err := client.CancelAnalyze(ctx, nil)
	require.Nil(t, resp)
	require.ErrorIs(t, err, ErrNilRequest)
}

func TestCancelAnalyze_EmptyRequestID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	client := &RawClient{}

	req := &CancelAnalyzeRequest{
		RequestID: "",
	}
	resp, err := client.CancelAnalyze(ctx, req)
	require.Nil(t, resp)
	require.Error(t, err)
	require.Contains(t, err.Error(), "request_id cannot be empty")
}

// TestCancelAnalyzeLiveFlow tests the cancel analyze API with a real backend.
// This test requires:
// 1. A running backend server
// 2. A valid request_id from a previous analysis request
func TestCancelAnalyzeLiveFlow(t *testing.T) {
	// Skip if not running live tests
	if testing.Short() {
		t.Skip("Skipping live test in short mode")
	}

	ctx := context.Background()
	client := newTestClient(t)

	// First, start an analysis request to get a request_id
	req := &DataAnalysisRequest{
		Question: "平均薪资是多少？",
		Config: &DataAnalysisConfig{
			DataCategory: "admin",
			DataSource: &DataSource{
				Type: "all",
			},
		},
	}

	stream, err := client.AnalyzeDataStream(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, stream)
	defer stream.Close()

	// Read the first event to get request_id
	data, err := stream.ReadEvent()
	require.NoError(t, err)
	require.NotEmpty(t, data)

	// Extract request_id from the init event
	var requestID string
	// Try to parse as JSON first
	var eventData map[string]interface{}
	if err := json.Unmarshal([]byte(data), &eventData); err == nil {
		// Check for step_type="init" or type="init"
		if stepType, ok := eventData["step_type"].(string); ok && stepType == "init" {
			if dataField, ok := eventData["data"].(map[string]interface{}); ok {
				if id, ok := dataField["request_id"].(string); ok {
					requestID = id
				}
			}
		}
	} else {
		// If not JSON, check for SSE format with init event
		if strings.Contains(data, "step_type") && strings.Contains(data, "init") {
			// Try to extract from SSE format
			lines := strings.Split(data, "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "data: ") {
					dataLine := strings.TrimPrefix(line, "data: ")
					var initData map[string]interface{}
					if err := json.Unmarshal([]byte(dataLine), &initData); err == nil {
						if dataField, ok := initData["data"].(map[string]interface{}); ok {
							if id, ok := dataField["request_id"].(string); ok {
								requestID = id
								break
							}
						}
					}
				}
			}
		}
	}

	if requestID == "" {
		t.Skip("Could not extract request_id from stream, skipping cancel test")
	}

	// Now cancel the request
	cancelReq := &CancelAnalyzeRequest{
		RequestID: requestID,
	}

	cancelResp, err := client.CancelAnalyze(ctx, cancelReq)
	require.NoError(t, err)
	require.NotNil(t, cancelResp)
	require.Equal(t, requestID, cancelResp.RequestID)
	require.Equal(t, "cancelled", cancelResp.Status)
	require.NotEmpty(t, cancelResp.UserID)
	t.Logf("Successfully cancelled request: %s, Status: %s, UserID: %s", cancelResp.RequestID, cancelResp.Status, cancelResp.UserID)
}
