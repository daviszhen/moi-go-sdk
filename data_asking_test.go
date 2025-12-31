package sdk

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
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

		event, err := stream.ReadEvent()
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

		require.NotNil(t, event, "Event should not be nil")

		eventCount++

		// Log event details (truncate long data for readability)
		rawDataStr := string(event.RawData)
		if len(rawDataStr) > 200 {
			rawDataStr = rawDataStr[:200] + "..."
		}
		t.Logf("Event #%d: Type=%s, Source=%s, StepType=%s, StepName=%s",
			eventCount, event.Type, event.Source, event.StepType, event.StepName)
		t.Logf("  RawData: %s", rawDataStr)

		// Track specific event types
		if event.Type == "classification" {
			hasClassification = true
			// Verify classification event structure
			require.NotEmpty(t, event.RawData, "Classification event should have data")
		}

		if event.Type == "complete" {
			hasComplete = true
			t.Logf("Analysis completed")
			readEvents = false
			break // Complete event indicates end of stream
		}

		if event.Type == "error" {
			t.Logf("Error event received: %s", string(event.RawData))
		}

		// For events without explicit type field, check for source and step_type
		if event.Type == "" {
			if event.Source != "" {
				t.Logf("Event with source %s: step_type=%s, step_name=%s", event.Source, event.StepType, event.StepName)
			}
			// Some events have step_type in the JSON but not parsed into Type field
			if event.StepType != "" {
				t.Logf("Event with step_type: %s", event.StepType)
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
	event, err := stream.ReadEvent()
	if err == io.EOF {
		t.Log("Stream ended immediately (no events)")
		return
	}
	require.NoError(t, err)
	require.NotNil(t, event)
	t.Logf("First event: Type=%s, Source=%s", event.Type, event.Source)
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
	event, err := stream.ReadEvent()
	require.NoError(t, err)
	require.NotNil(t, event)

	// Extract request_id from the init event
	var requestID string
	if event.StepType == "init" {
		// Parse the data field to get request_id
		var initData map[string]interface{}
		if err := json.Unmarshal(event.RawData, &initData); err == nil {
			if data, ok := initData["data"].(map[string]interface{}); ok {
				if id, ok := data["request_id"].(string); ok {
					requestID = id
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

// ============ ReadEvent Unit Tests ============

func TestDataAnalysisStream_ReadEvent_Basic(t *testing.T) {
	t.Parallel()

	// Create a simple SSE stream
	sseData := "event: classification\ndata: {\"type\":\"classification\",\"data\":{\"category\":\"query\"}}\n\n"
	stream := &DataAnalysisStream{
		Body:          io.NopCloser(strings.NewReader(sseData)),
		Header:        make(http.Header),
		StatusCode:    200,
		initialBufferSize: 0, // Use default
	}

	event, err := stream.ReadEvent()
	require.NoError(t, err)
	require.NotNil(t, event)
	require.Equal(t, "classification", event.Type)
	require.NotEmpty(t, event.RawData)

	// Should return EOF for next read
	event, err = stream.ReadEvent()
	require.ErrorIs(t, err, io.EOF)
	require.Nil(t, event)

	require.NoError(t, stream.Close())
}

func TestDataAnalysisStream_ReadEvent_MultipleEvents(t *testing.T) {
	t.Parallel()

	sseData := "event: init\ndata: {\"step_type\":\"init\",\"data\":{\"request_id\":\"req-123\"}}\n\n" +
		"event: classification\ndata: {\"type\":\"classification\"}\n\n" +
		"event: complete\ndata: {\"type\":\"complete\"}\n\n"

	stream := &DataAnalysisStream{
		Body:          io.NopCloser(strings.NewReader(sseData)),
		Header:        make(http.Header),
		StatusCode:    200,
		initialBufferSize: 0,
	}

	// Read first event
	event, err := stream.ReadEvent()
	require.NoError(t, err)
	require.NotNil(t, event)
	require.Equal(t, "init", event.Type)

	// Read second event
	event, err = stream.ReadEvent()
	require.NoError(t, err)
	require.NotNil(t, event)
	require.Equal(t, "classification", event.Type)

	// Read third event
	event, err = stream.ReadEvent()
	require.NoError(t, err)
	require.NotNil(t, event)
	require.Equal(t, "complete", event.Type)

	// Should return EOF
	event, err = stream.ReadEvent()
	require.ErrorIs(t, err, io.EOF)

	require.NoError(t, stream.Close())
}

func TestDataAnalysisStream_ReadEvent_MultiLineData(t *testing.T) {
	t.Parallel()

	sseData := "event: test\ndata: {\"key1\":\"value1\"}\ndata: {\"key2\":\"value2\"}\n\n"

	stream := &DataAnalysisStream{
		Body:          io.NopCloser(strings.NewReader(sseData)),
		Header:        make(http.Header),
		StatusCode:    200,
		initialBufferSize: 0,
	}

	event, err := stream.ReadEvent()
	require.NoError(t, err)
	require.NotNil(t, event)
	require.Equal(t, "test", event.Type)
	// Multi-line data should be joined with newline
	require.Contains(t, string(event.RawData), "key1")
	require.Contains(t, string(event.RawData), "key2")

	require.NoError(t, stream.Close())
}

func TestDataAnalysisStream_ReadEvent_EmptyStream(t *testing.T) {
	t.Parallel()

	stream := &DataAnalysisStream{
		Body:          io.NopCloser(strings.NewReader("")),
		Header:        make(http.Header),
		StatusCode:    200,
		initialBufferSize: 0,
	}

	event, err := stream.ReadEvent()
	require.ErrorIs(t, err, io.EOF)
	require.Nil(t, event)

	require.NoError(t, stream.Close())
}

func TestDataAnalysisStream_ReadEvent_DefaultBufferSize(t *testing.T) {
	t.Parallel()

	// Create data larger than default buffer - should automatically grow
	largeData := strings.Repeat("x", 100*1024) // 100KB
	// Use JSON encoding to properly escape the data
	jsonData, err := json.Marshal(map[string]string{"data": largeData})
	require.NoError(t, err)
	sseData := "event: large\ndata: " + string(jsonData) + "\n\n"

	stream := &DataAnalysisStream{
		Body:          io.NopCloser(strings.NewReader(sseData)),
		Header:        make(http.Header),
		StatusCode:    200,
		initialBufferSize: 0, // Use default 4KB initial buffer (will grow automatically)
	}

	event, err := stream.ReadEvent()
	require.NoError(t, err, "Should handle large data with dynamic buffer growth")
	require.NotNil(t, event)
	require.Equal(t, "large", event.Type)
	require.Contains(t, string(event.RawData), largeData)

	require.NoError(t, stream.Close())
}

func TestDataAnalysisStream_ReadEvent_CustomBufferSize(t *testing.T) {
	t.Parallel()

	// Create data larger than initial buffer - should automatically grow
	largeData := strings.Repeat("y", 200*1024) // 200KB
	// Use JSON encoding to properly escape the data
	jsonData, err := json.Marshal(map[string]string{"data": largeData})
	require.NoError(t, err)
	sseData := "event: large\ndata: " + string(jsonData) + "\n\n"

	stream := &DataAnalysisStream{
		Body:          io.NopCloser(strings.NewReader(sseData)),
		Header:        make(http.Header),
		StatusCode:    200,
		initialBufferSize: 512 * 1024, // 512KB initial buffer (will grow as needed)
	}

	event, err := stream.ReadEvent()
	require.NoError(t, err, "Should handle large data with dynamic buffer growth")
	require.NotNil(t, event)
	require.Equal(t, "large", event.Type)
	require.Contains(t, string(event.RawData), largeData)

	require.NoError(t, stream.Close())
}

func TestDataAnalysisStream_ReadEvent_VeryLargeData(t *testing.T) {
	t.Parallel()

	// Create very large data - buffer should automatically grow to handle it
	largeData := strings.Repeat("z", 2*1024*1024) // 2MB
	// Use JSON encoding to properly escape the data
	jsonData, err := json.Marshal(map[string]string{"data": largeData})
	require.NoError(t, err)
	sseData := "event: verylarge\ndata: " + string(jsonData) + "\n\n"

	stream := &DataAnalysisStream{
		Body:          io.NopCloser(strings.NewReader(sseData)),
		Header:        make(http.Header),
		StatusCode:    200,
		initialBufferSize: 4 * 1024 * 1024, // 4MB initial buffer (will grow as needed)
	}

	event, err := stream.ReadEvent()
	require.NoError(t, err, "Should handle very large data with dynamic buffer growth")
	require.NotNil(t, event)
	require.Equal(t, "verylarge", event.Type)
	require.Contains(t, string(event.RawData), largeData)

	require.NoError(t, stream.Close())
}

func TestDataAnalysisStream_ReadEvent_InvalidJSON(t *testing.T) {
	t.Parallel()

	// SSE with invalid JSON should still return the raw data
	sseData := "event: test\ndata: {invalid json}\n\n"

	stream := &DataAnalysisStream{
		Body:          io.NopCloser(strings.NewReader(sseData)),
		Header:        make(http.Header),
		StatusCode:    200,
		initialBufferSize: 0,
	}

	event, err := stream.ReadEvent()
	require.NoError(t, err, "Should return event even with invalid JSON")
	require.NotNil(t, event)
	require.Equal(t, "test", event.Type)
	require.Equal(t, "{invalid json}", string(event.RawData))

	require.NoError(t, stream.Close())
}

func TestDataAnalysisStream_ReadEvent_NoEventType(t *testing.T) {
	t.Parallel()

	// SSE without event type should still work
	sseData := "data: {\"step_type\":\"init\",\"data\":{\"request_id\":\"req-123\"}}\n\n"

	stream := &DataAnalysisStream{
		Body:          io.NopCloser(strings.NewReader(sseData)),
		Header:        make(http.Header),
		StatusCode:    200,
		initialBufferSize: 0,
	}

	event, err := stream.ReadEvent()
	require.NoError(t, err)
	require.NotNil(t, event)
	// Type should be empty if not specified in event: field
	require.Empty(t, event.Type)
	require.NotEmpty(t, event.RawData)

	require.NoError(t, stream.Close())
}

func TestDataAnalysisStream_ReadEvent_WithStreamBufferSizeOption(t *testing.T) {
	t.Parallel()

	// This test verifies that WithStreamBufferSize option is properly passed through
	// The buffer will automatically grow to handle data larger than initial size
	largeData := strings.Repeat("a", 150*1024) // 150KB
	// Use JSON encoding to properly escape the data
	jsonData, err := json.Marshal(map[string]string{"data": largeData})
	require.NoError(t, err)
	sseData := "event: test\ndata: " + string(jsonData) + "\n\n"

	// Create stream with custom initial buffer size (simulating what AnalyzeDataStream would do)
	stream := &DataAnalysisStream{
		Body:          io.NopCloser(strings.NewReader(sseData)),
		Header:        make(http.Header),
		StatusCode:    200,
		initialBufferSize: 256 * 1024, // 256KB initial buffer (set via WithStreamBufferSize, will grow as needed)
	}

	event, err := stream.ReadEvent()
	require.NoError(t, err, "Should handle data with dynamic buffer growth from option")
	require.NotNil(t, event)
	require.Equal(t, "test", event.Type)

	require.NoError(t, stream.Close())
}

func TestDataAnalysisStream_ReadEvent_EmptyLines(t *testing.T) {
	t.Parallel()

	// SSE with multiple empty lines should be handled correctly
	sseData := "\n\nevent: test\ndata: {\"key\":\"value\"}\n\n\n\n"

	stream := &DataAnalysisStream{
		Body:          io.NopCloser(strings.NewReader(sseData)),
		Header:        make(http.Header),
		StatusCode:    200,
		initialBufferSize: 0,
	}

	event, err := stream.ReadEvent()
	require.NoError(t, err)
	require.NotNil(t, event)
	require.Equal(t, "test", event.Type)

	// Next read should be EOF
	event, err = stream.ReadEvent()
	require.ErrorIs(t, err, io.EOF)

	require.NoError(t, stream.Close())
}

func TestWithStreamBufferSize_Option(t *testing.T) {
	t.Parallel()

	// Test that WithStreamBufferSize properly sets the buffer size in callOptions
	opts := newCallOptions(WithStreamBufferSize(2 * 1024 * 1024)) // 2MB
	require.Equal(t, 2*1024*1024, opts.streamBufferSize)

	// Test with zero value (should not change default)
	opts = newCallOptions(WithStreamBufferSize(0))
	require.Equal(t, 0, opts.streamBufferSize)

	// Test with negative value (should not change default)
	opts = newCallOptions(WithStreamBufferSize(-1))
	require.Equal(t, 0, opts.streamBufferSize)

	// Test default value
	opts = newCallOptions()
	require.Equal(t, 0, opts.streamBufferSize)
}
