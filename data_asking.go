package sdk

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// DataAnalysisStream wraps a streaming HTTP response for data analysis API.
//
// The stream returns Server-Sent Events (SSE) format. Use ReadEvent to read
// individual events from the stream.
//
// Example:
//
//	stream, err := client.AnalyzeDataStream(ctx, &sdk.DataAnalysisRequest{
//		Question: "2024年收入下降的原因是什么？",
//		SessionID: stringPtr("session_123"),
//	})
//	if err != nil {
//		return err
//	}
//	defer stream.Close()
//
//	for {
//		event, err := stream.ReadEvent()
//		if err == io.EOF {
//			break
//		}
//		if err != nil {
//			return err
//		}
//		fmt.Printf("Event type: %s\n", event.Type)
//	}
//
// timeoutReader wraps an io.ReadCloser and provides timeout control that resets on each successful read.
// The timeout is applied to the interval between reads, not the total read time.
type timeoutReader struct {
	reader    io.ReadCloser
	timeout   time.Duration
	readMutex chan struct{} // Serializes read operations
}

func newTimeoutReader(reader io.ReadCloser, timeout time.Duration) *timeoutReader {
	return &timeoutReader{
		reader:    reader,
		timeout:   timeout,
		readMutex: make(chan struct{}, 1),
	}
}

func (r *timeoutReader) Read(p []byte) (n int, err error) {
	// Serialize reads to ensure timeout is properly reset
	r.readMutex <- struct{}{}
	defer func() { <-r.readMutex }()

	if r.timeout <= 0 {
		// No timeout, read directly
		return r.reader.Read(p)
	}

	// Create a context with timeout for this read operation
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	// Use a channel to receive the read result
	type result struct {
		n   int
		err error
	}
	resultCh := make(chan result, 1)

	// Perform the read in a goroutine
	// Note: The read operation itself is thread-safe, as io.ReadCloser implementations
	// should handle concurrent reads appropriately, or we serialize them via readMutex
	go func() {
		// Create a local buffer to avoid potential race conditions
		// We'll read into a buffer and then copy to p
		buf := make([]byte, len(p))
		n, err := r.reader.Read(buf)
		if n > 0 {
			copy(p, buf[:n])
		}
		resultCh <- result{n: n, err: err}
	}()

	// Wait for either the read to complete or the timeout
	select {
	case res := <-resultCh:
		// Read completed successfully - timeout is effectively reset for the next read
		return res.n, res.err
	case <-ctx.Done():
		// Timeout - no data received within the timeout period
		return 0, fmt.Errorf("read timeout: no data received within %v", r.timeout)
	}
}

func (r *timeoutReader) Close() error {
	if r.reader != nil {
		return r.reader.Close()
	}
	return nil
}

type DataAnalysisStream struct {
	// Body is the response body that must be closed by the caller
	Body io.ReadCloser
	// Header contains the HTTP response headers
	Header http.Header
	// StatusCode is the HTTP status code
	StatusCode int
	reader     *bufio.Reader
	// initialBufferSize is the initial buffer size for the reader (0 means use default)
	// The buffer will dynamically grow as needed to handle large lines
	initialBufferSize int
	// readTimeout is the timeout between messages in streaming responses
	// This timeout is reset each time data is successfully read
	readTimeout time.Duration
}

// Close releases the underlying HTTP response body.
func (s *DataAnalysisStream) Close() error {
	if s == nil || s.Body == nil {
		return nil
	}
	return s.Body.Close()
}

// ReadEvent reads the next SSE event from the stream.
//
// Returns io.EOF when the stream is complete.
//
// Example:
//
//	for {
//		event, err := stream.ReadEvent()
//		if err == io.EOF {
//			break
//		}
//		if err != nil {
//			return err
//		}
//		// Process event
//	}
//
// readLine reads a line from the reader, dynamically growing the buffer as needed.
// This allows handling lines of arbitrary length without token size limits.
// The read timeout is reset each time data is successfully read.
func (s *DataAnalysisStream) readLine() (string, error) {
	if s.reader == nil {
		bufferSize := s.initialBufferSize
		if bufferSize == 0 {
			bufferSize = 4096 // Default: 4KB initial buffer
		}
		// Wrap the body with a timeout reader if timeout is configured
		body := s.Body
		if s.readTimeout > 0 {
			body = newTimeoutReader(s.Body, s.readTimeout)
		}
		s.reader = bufio.NewReaderSize(body, bufferSize)
	}

	var line []byte
	var isPrefix bool
	var err error

	// ReadLine may return a partial line if it's too long for the buffer.
	// We need to keep reading until we get the complete line.
	// The timeout is automatically reset on each successful read by the timeoutReader.
	for {
		var part []byte
		part, isPrefix, err = s.reader.ReadLine()
		if err != nil {
			// Check if error is due to read timeout
			if strings.Contains(err.Error(), "read timeout") {
				return "", err
			}
			if err == io.EOF && len(line) > 0 {
				// EOF but we have data, return it
				return string(line), nil
			}
			return "", err
		}

		// Data successfully read - timeout is automatically reset by timeoutReader
		line = append(line, part...)
		if !isPrefix {
			// Complete line read
			break
		}
		// Line was too long, continue reading
	}

	return string(line), nil
}

func (s *DataAnalysisStream) ReadEvent() (*DataAnalysisStreamEvent, error) {
	var event DataAnalysisStreamEvent
	var dataLines []string
	var eventType string

	for {
		line, err := s.readLine()
		if err != nil {
			if err == io.EOF {
				// Handle last event if any
				if len(dataLines) > 0 {
					dataStr := strings.Join(dataLines, "\n")
					event.RawData = []byte(dataStr)
					if err := json.Unmarshal([]byte(dataStr), &event); err != nil {
						// If JSON parsing fails, return raw data
						if eventType != "" {
							event.Type = eventType
						}
						return &event, nil
					}
					if eventType != "" {
						event.Type = eventType
					}
					return &event, nil
				}
				return nil, io.EOF
			}
			return nil, fmt.Errorf("read stream: %w", err)
		}

		if line == "" {
			// Empty line indicates end of event
			if len(dataLines) > 0 {
				// Parse the accumulated data
				dataStr := strings.Join(dataLines, "\n")
				event.RawData = []byte(dataStr)
				if err := json.Unmarshal([]byte(dataStr), &event); err != nil {
					// If JSON parsing fails, return raw data
					if eventType != "" {
						event.Type = eventType
					}
					return &event, nil
				}
				if eventType != "" {
					event.Type = eventType
				}
				return &event, nil
			}
			continue
		}

		// Parse SSE format: "field: value"
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			dataLines = append(dataLines, data)
		} else if strings.HasPrefix(line, "event: ") {
			eventType = strings.TrimPrefix(line, "event: ")
		}
		// Ignore other SSE fields (id, retry, etc.)
	}
}

// AnalyzeDataStream performs data analysis and returns a streaming response.
//
// This method sends a POST request to /byoa/api/v1/data_asking/analyze and
// returns a stream of Server-Sent Events (SSE) containing analysis results.
//
// The stream includes events such as:
//   - init: Initialization event (first event) with request_id and session_title
//     (step_type="init", data contains request_id and session_title)
//   - classification: Question classification result
//   - decomposition: Attribution question decomposition (attribution only)
//   - step_start: Step start (attribution only)
//   - step_complete: Step completion (attribution only)
//   - chunks/answer_chunk: RAG interface data (with source="rag")
//   - step_type/step_name: NL2SQL interface data (with source="nl2sql")
//   - complete: Analysis complete
//   - error: Error information
//
// Example:
//
//	stream, err := client.AnalyzeDataStream(ctx, &sdk.DataAnalysisRequest{
//		Question: "2024年收入下降的原因是什么？",
//		SessionID: stringPtr("session_123"),
//		Config: &sdk.DataAnalysisConfig{
//			DataCategory: "admin",
//			DataSource: &sdk.DataSource{
//				Type: "all",
//			},
//		},
//	}, sdk.WithStreamBufferSize(1024*1024)) // Optional: set buffer size for large data lines
//	if err != nil {
//		return err
//	}
//	defer stream.Close()
//
//	for {
//		event, err := stream.ReadEvent()
//		if err == io.EOF {
//			break
//		}
//		if err != nil {
//			return err
//		}
//		fmt.Printf("Event: %+v\n", event)
//	}
func (c *RawClient) AnalyzeDataStream(ctx context.Context, req *DataAnalysisRequest, opts ...CallOption) (*DataAnalysisStream, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	if strings.TrimSpace(req.Question) == "" {
		return nil, fmt.Errorf("question cannot be empty")
	}

	callOpts := newCallOptions(opts...)

	// Marshal request body
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request body: %w", err)
	}
	reader := bytes.NewReader(payload)

	// Build request
	path := "/byoa/api/v1/data_asking/analyze"
	fullURL := c.baseURL + ensureLeadingSlash(path)
	if len(callOpts.query) > 0 {
		delimiter := "?"
		if strings.Contains(fullURL, "?") {
			delimiter = "&"
		}
		fullURL = fullURL + delimiter + callOpts.query.Encode()
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, reader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set(headerAPIKey, c.apiKey)
	if c.userAgent != "" {
		httpReq.Header.Set(headerUserAgent, c.userAgent)
	}
	mergeHeaders(httpReq.Header, c.defaultHeaders, false)
	if callOpts.requestID != "" {
		httpReq.Header.Set(headerRequestID, callOpts.requestID)
	}
	mergeHeaders(httpReq.Header, callOpts.headers, true)
	httpReq.Header.Set(headerContentType, mimeJSON)
	httpReq.Header.Set(headerAccept, "text/event-stream")

	// Create a client with no timeout for streaming responses
	// The stream can still be cancelled via context
	// This prevents timeout errors while reading long-running streams
	streamClient := &http.Client{
		Timeout:   0,                      // No timeout - allows reading long-running streams
		Transport: c.httpClient.Transport, // Reuse the transport from the original client
	}
	if streamClient.Transport == nil {
		// If original client has no custom transport, use default
		streamClient.Transport = http.DefaultTransport
	}

	// Execute request
	resp, err := streamClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		data, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, &HTTPError{StatusCode: resp.StatusCode, Body: data}
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/event-stream") && !strings.Contains(contentType, "text/plain") {
		// Not a streaming response, try to parse as error
		data, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected content type: %s, body: %s", contentType, string(data))
	}

	return &DataAnalysisStream{
		Body:              resp.Body,
		Header:            resp.Header.Clone(),
		StatusCode:        resp.StatusCode,
		initialBufferSize: callOpts.streamBufferSize,
		readTimeout:       callOpts.streamReadTimeout,
	}, nil
}

// CancelAnalyze cancels an ongoing data analysis request.
//
// This method sends a POST request to /byoa/api/v1/data_asking/cancel to cancel
// a data analysis request that is currently in progress.
//
// The request_id parameter identifies the analysis request to cancel. Only the
// user who initiated the request can cancel it.
//
// Example:
//
//	resp, err := client.CancelAnalyze(ctx, &sdk.CancelAnalyzeRequest{
//		RequestID: "request-123",
//	})
//	if err != nil {
//		return err
//	}
//	fmt.Printf("Cancelled request: %s, Status: %s, User: %s\n", resp.RequestID, resp.Status, resp.UserName)
func (c *RawClient) CancelAnalyze(ctx context.Context, req *CancelAnalyzeRequest, opts ...CallOption) (*CancelAnalyzeResponse, error) {
	if req == nil {
		return nil, ErrNilRequest
	}
	if strings.TrimSpace(req.RequestID) == "" {
		return nil, fmt.Errorf("request_id cannot be empty")
	}

	// Add request_id as query parameter
	opts = append(opts, WithQueryParam("request_id", req.RequestID))

	var resp CancelAnalyzeResponse
	if err := c.postJSON(ctx, "/byoa/api/v1/data_asking/cancel", nil, &resp, opts...); err != nil {
		return nil, err
	}
	return &resp, nil
}
