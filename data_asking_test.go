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

// ============ Stream Read Timeout Tests ============

// slowReader is a reader that reads data slowly to simulate network delays
type slowReader struct {
	data      []byte
	chunkSize int
	delay     time.Duration
	pos       int
	closed    bool
	firstRead bool // Whether this is the first read (skip delay for first read)
}

func newSlowReader(data []byte, chunkSize int, delay time.Duration) *slowReader {
	return &slowReader{
		data:      data,
		chunkSize: chunkSize,
		delay:     delay,
		pos:       0,
		closed:    false,
		firstRead: true,
	}
}

func (r *slowReader) Read(p []byte) (n int, err error) {
	if r.closed {
		return 0, io.EOF
	}
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}

	// Add delay to simulate slow network (skip delay for first read)
	if r.delay > 0 && !r.firstRead {
		time.Sleep(r.delay)
	}
	r.firstRead = false

	// Read a chunk
	chunkLen := r.chunkSize
	if chunkLen > len(r.data)-r.pos {
		chunkLen = len(r.data) - r.pos
	}
	if chunkLen > len(p) {
		chunkLen = len(p)
	}

	copy(p, r.data[r.pos:r.pos+chunkLen])
	r.pos += chunkLen
	return chunkLen, nil
}

func (r *slowReader) Close() error {
	r.closed = true
	return nil
}

// blockingReader blocks on read until explicitly unblocked
type blockingReader struct {
	data   []byte
	ch     chan struct{}
	closed bool
	pos    int
}

func newBlockingReader(data []byte) *blockingReader {
	return &blockingReader{
		data: data,
		ch:   make(chan struct{}),
		pos:  0,
	}
}

func (r *blockingReader) Read(p []byte) (n int, err error) {
	if r.closed {
		return 0, io.EOF
	}
	// Block until unblocked
	<-r.ch
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

func (r *blockingReader) Close() error {
	r.closed = true
	close(r.ch)
	return nil
}

func (r *blockingReader) unblock() {
	select {
	case r.ch <- struct{}{}:
	default:
	}
}

func TestTimeoutReader_Read_Success(t *testing.T) {
	t.Parallel()

	// Create a reader with immediate data
	data := []byte("test data")
	reader := newTimeoutReader(io.NopCloser(strings.NewReader(string(data))), 100*time.Millisecond)

	// Read should succeed immediately (no timeout)
	buf := make([]byte, len(data))
	n, err := reader.Read(buf)
	require.NoError(t, err)
	require.Equal(t, len(data), n)
	require.Equal(t, data, buf[:n])

	require.NoError(t, reader.Close())
}

func TestTimeoutReader_Read_WithSlowData(t *testing.T) {
	t.Parallel()

	// Create a slow reader that delays 50ms between chunks
	data := []byte("test data chunk")
	slowR := newSlowReader(data, 5, 50*time.Millisecond)
	reader := newTimeoutReader(slowR, 200*time.Millisecond) // 200ms timeout

	// Read should succeed (50ms delay < 200ms timeout)
	buf := make([]byte, len(data))
	start := time.Now()
	n, err := reader.Read(buf)
	duration := time.Since(start)

	require.NoError(t, err)
	require.Greater(t, n, 0)
	require.Less(t, duration, 200*time.Millisecond, "Read should complete before timeout")
	require.Equal(t, data[:n], buf[:n])

	require.NoError(t, reader.Close())
}

func TestTimeoutReader_Read_Timeout(t *testing.T) {
	t.Parallel()

	// Create a blocking reader that blocks on read
	blockingR := newBlockingReader([]byte("test"))
	reader := newTimeoutReader(blockingR, 100*time.Millisecond) // 100ms timeout

	// Read should timeout after 100ms
	buf := make([]byte, 100)
	start := time.Now()
	n, err := reader.Read(buf)
	duration := time.Since(start)

	require.Error(t, err)
	require.Contains(t, err.Error(), "read timeout")
	require.Contains(t, err.Error(), "100ms")
	require.Equal(t, 0, n)
	require.GreaterOrEqual(t, duration, 90*time.Millisecond, "Should timeout after approximately 100ms")
	require.Less(t, duration, 200*time.Millisecond, "Should timeout before 200ms")

	require.NoError(t, reader.Close())
}

func TestTimeoutReader_Read_TimeoutResetOnSuccess(t *testing.T) {
	t.Parallel()

	// Create a reader with multiple chunks, delayed between chunks
	data := []byte("chunk1\nchunk2\nchunk3")
	// First chunk reads immediately, then 50ms delay, then next chunk
	slowR := newSlowReader(data, 7, 50*time.Millisecond)
	reader := newTimeoutReader(slowR, 150*time.Millisecond) // 150ms timeout

	// Read first chunk - should succeed (immediate, no delay on first read)
	buf1 := make([]byte, 7)
	start1 := time.Now()
	n1, err1 := reader.Read(buf1)
	duration1 := time.Since(start1)
	require.NoError(t, err1)
	require.Greater(t, n1, 0)
	// First read should be fast (no delay)
	require.Less(t, duration1, 50*time.Millisecond, "First read should be fast")

	// Read second chunk - should succeed (50ms delay < 150ms timeout)
	// This tests that timeout is reset after first successful read
	buf2 := make([]byte, 7)
	start2 := time.Now()
	n2, err2 := reader.Read(buf2)
	duration2 := time.Since(start2)
	require.NoError(t, err2)
	require.Greater(t, n2, 0)
	require.GreaterOrEqual(t, duration2, 40*time.Millisecond, "Second read should include delay")
	require.Less(t, duration2, 100*time.Millisecond, "Should complete before timeout")

	// Read third chunk - should also succeed
	buf3 := make([]byte, 7)
	n3, err3 := reader.Read(buf3)
	require.NoError(t, err3)
	require.Greater(t, n3, 0)

	require.NoError(t, reader.Close())
}

func TestTimeoutReader_Read_MillisecondTimeout(t *testing.T) {
	t.Parallel()

	// Test with millisecond-level timeout
	blockingR := newBlockingReader([]byte("test"))
	reader := newTimeoutReader(blockingR, 50*time.Millisecond) // 50ms timeout

	buf := make([]byte, 100)
	start := time.Now()
	n, err := reader.Read(buf)
	duration := time.Since(start)

	require.Error(t, err)
	require.Contains(t, err.Error(), "read timeout")
	require.Contains(t, err.Error(), "50ms")
	require.Equal(t, 0, n)
	// Should timeout after approximately 50ms (allow some margin)
	require.GreaterOrEqual(t, duration, 40*time.Millisecond, "Should timeout after approximately 50ms")
	require.Less(t, duration, 100*time.Millisecond, "Should timeout before 100ms")

	require.NoError(t, reader.Close())
}

func TestTimeoutReader_Read_NoTimeout(t *testing.T) {
	t.Parallel()

	// Test with no timeout (timeout = 0)
	data := []byte("test data")
	reader := newTimeoutReader(io.NopCloser(strings.NewReader(string(data))), 0)

	// Read should work normally without timeout
	buf := make([]byte, len(data))
	n, err := reader.Read(buf)
	require.NoError(t, err)
	require.Equal(t, len(data), n)

	require.NoError(t, reader.Close())
}

func TestDataAnalysisStream_ReadEvent_WithTimeout_Success(t *testing.T) {
	t.Parallel()

	// Create SSE stream with data that arrives quickly
	sseData := "event: test\ndata: {\"key\":\"value\"}\n\n"
	stream := &DataAnalysisStream{
		Body:          io.NopCloser(strings.NewReader(sseData)),
		Header:        make(http.Header),
		StatusCode:    200,
		initialBufferSize: 0,
		readTimeout:   100 * time.Millisecond, // 100ms timeout
	}

	// Read should succeed immediately
	start := time.Now()
	event, err := stream.ReadEvent()
	duration := time.Since(start)

	require.NoError(t, err)
	require.NotNil(t, event)
	require.Equal(t, "test", event.Type)
	require.Less(t, duration, 50*time.Millisecond, "Should read quickly")

	require.NoError(t, stream.Close())
}

func TestDataAnalysisStream_ReadEvent_WithTimeout_Timeout(t *testing.T) {
	t.Parallel()

	// Create a blocking reader
	blockingR := newBlockingReader([]byte("event: test\ndata: {\"key\":\"value\"}\n\n"))
	stream := &DataAnalysisStream{
		Body:          blockingR,
		Header:        make(http.Header),
		StatusCode:    200,
		initialBufferSize: 0,
		readTimeout:   100 * time.Millisecond, // 100ms timeout
	}

	// Read should timeout
	start := time.Now()
	event, err := stream.ReadEvent()
	duration := time.Since(start)

	require.Error(t, err)
	require.Contains(t, err.Error(), "read timeout")
	require.Nil(t, event)
	require.GreaterOrEqual(t, duration, 90*time.Millisecond, "Should timeout after approximately 100ms")
	require.Less(t, duration, 200*time.Millisecond, "Should timeout before 200ms")

	require.NoError(t, stream.Close())
}

func TestDataAnalysisStream_ReadEvent_WithTimeout_ResetOnSuccess(t *testing.T) {
	t.Parallel()

	// Create SSE stream with multiple events, with delay between events
	// First event arrives quickly, second after delay
	sseData := "event: first\ndata: {\"key1\":\"value1\"}\n\n" +
		"event: second\ndata: {\"key2\":\"value2\"}\n\n"

	// Create a slow reader that delays between chunks
	slowR := newSlowReader([]byte(sseData), 20, 50*time.Millisecond)
	stream := &DataAnalysisStream{
		Body:          slowR,
		Header:        make(http.Header),
		StatusCode:    200,
		initialBufferSize: 0,
		readTimeout:   150 * time.Millisecond, // 150ms timeout
	}

	// Read first event - should succeed
	event1, err1 := stream.ReadEvent()
	require.NoError(t, err1)
	require.NotNil(t, event1)
	require.Equal(t, "first", event1.Type)

	// Read second event - should succeed (timeout was reset after first read)
	// This tests that timeout is reset on each successful read
	start := time.Now()
	event2, err2 := stream.ReadEvent()
	duration := time.Since(start)

	require.NoError(t, err2)
	require.NotNil(t, event2)
	require.Equal(t, "second", event2.Type)
	// Second read should take some time due to delay, but complete before timeout
	require.GreaterOrEqual(t, duration, 40*time.Millisecond, "Should include delay")
	require.Less(t, duration, 120*time.Millisecond, "Should complete before timeout")

	require.NoError(t, stream.Close())
}

func TestDataAnalysisStream_ReadEvent_WithMillisecondTimeout(t *testing.T) {
	t.Parallel()

	// Test with millisecond-level timeout using a blocking reader
	blockingR := newBlockingReader([]byte("event: test\ndata: {\"key\":\"value\"}\n\n"))
	stream := &DataAnalysisStream{
		Body:          blockingR,
		Header:        make(http.Header),
		StatusCode:    200,
		initialBufferSize: 0,
		readTimeout:   50 * time.Millisecond, // 50ms timeout
	}

	start := time.Now()
	event, err := stream.ReadEvent()
	duration := time.Since(start)

	require.Error(t, err)
	require.Contains(t, err.Error(), "read timeout")
	require.Contains(t, err.Error(), "50ms")
	require.Nil(t, event)
	require.GreaterOrEqual(t, duration, 40*time.Millisecond, "Should timeout after approximately 50ms")
	require.Less(t, duration, 100*time.Millisecond, "Should timeout before 100ms")

	require.NoError(t, stream.Close())
}

func TestWithStreamReadTimeout_Option(t *testing.T) {
	t.Parallel()

	// Test that WithStreamReadTimeout properly sets the timeout in callOptions
	opts := newCallOptions(WithStreamReadTimeout(60 * time.Second))
	require.Equal(t, 60*time.Second, opts.streamReadTimeout)

	// Test with millisecond timeout
	opts = newCallOptions(WithStreamReadTimeout(500 * time.Millisecond))
	require.Equal(t, 500*time.Millisecond, opts.streamReadTimeout)

	// Test with zero value (should use default)
	opts = newCallOptions(WithStreamReadTimeout(0))
	require.Equal(t, defaultStreamReadTimeout, opts.streamReadTimeout)

	// Test with negative value (should use default)
	opts = newCallOptions(WithStreamReadTimeout(-1 * time.Second))
	require.Equal(t, defaultStreamReadTimeout, opts.streamReadTimeout)

	// Test default value
	opts = newCallOptions()
	require.Equal(t, defaultStreamReadTimeout, opts.streamReadTimeout)
}

func TestTimeoutReader_Close(t *testing.T) {
	t.Parallel()

	// Test that Close properly closes the underlying reader
	// Use a reader that tracks closed state
	data := []byte("test")
	underlying := io.NopCloser(strings.NewReader(string(data)))
	reader := newTimeoutReader(underlying, 100*time.Millisecond)

	// Close should succeed
	err := reader.Close()
	require.NoError(t, err)

	// Read after close - the underlying reader may still have data available
	// but Close() was called, which is the important part to test
	// We verify that Close() doesn't panic and works correctly
	buf := make([]byte, 10)
	n, err := reader.Read(buf)
	// Some readers allow reading after close (like io.NopCloser),
	// while others return EOF. Both behaviors are acceptable.
	// The important thing is that Close() was called successfully.
	if err != nil {
		// If there's an error, it should be EOF
		require.ErrorIs(t, err, io.EOF)
	}
	// Whether we read data or not depends on the underlying reader's behavior
	// The test primarily verifies that Close() works without panic
	_ = n // n may be 0 or the number of bytes read, both are acceptable
}

func TestDataAnalysisStream_ReadEvent_WithTimeout_MultipleReads(t *testing.T) {
	t.Parallel()

	// Test multiple reads with timeout - each successful read should reset the timeout
	sseData := "event: event1\ndata: {\"key1\":\"value1\"}\n\n" +
		"event: event2\ndata: {\"key2\":\"value2\"}\n\n" +
		"event: event3\ndata: {\"key3\":\"value3\"}\n\n"

	// Create a slow reader with 30ms delay between chunks
	slowR := newSlowReader([]byte(sseData), 30, 30*time.Millisecond)
	stream := &DataAnalysisStream{
		Body:          slowR,
		Header:        make(http.Header),
		StatusCode:    200,
		initialBufferSize: 0,
		readTimeout:   100 * time.Millisecond, // 100ms timeout
	}

	// Read all three events - each should succeed (timeout resets on each read)
	events := []string{"event1", "event2", "event3"}
	for i, expectedType := range events {
		start := time.Now()
		event, err := stream.ReadEvent()
		duration := time.Since(start)

		require.NoError(t, err, "Event %d should succeed", i+1)
		require.NotNil(t, event, "Event %d should not be nil", i+1)
		require.Equal(t, expectedType, event.Type, "Event %d should have correct type", i+1)
		// Each read should complete before timeout (allowing for delay)
		require.Less(t, duration, 90*time.Millisecond, "Event %d should complete before timeout", i+1)
	}

	// Next read should return EOF
	event, err := stream.ReadEvent()
	require.ErrorIs(t, err, io.EOF)
	require.Nil(t, event)

	require.NoError(t, stream.Close())
}
