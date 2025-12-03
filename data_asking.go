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
type DataAnalysisStream struct {
	// Body is the response body that must be closed by the caller
	Body io.ReadCloser
	// Header contains the HTTP response headers
	Header http.Header
	// StatusCode is the HTTP status code
	StatusCode int
	scanner    *bufio.Scanner
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
func (s *DataAnalysisStream) ReadEvent() (*DataAnalysisStreamEvent, error) {
	if s.scanner == nil {
		s.scanner = bufio.NewScanner(s.Body)
	}

	var event DataAnalysisStreamEvent
	var dataLines []string
	var eventType string

	for s.scanner.Scan() {
		line := s.scanner.Text()
		if line == "" {
			// Empty line indicates end of event
			if len(dataLines) > 0 {
				// Parse the accumulated data
				dataStr := strings.Join(dataLines, "\n")
				event.RawData = []byte(dataStr)
				if err := json.Unmarshal([]byte(dataStr), &event); err != nil {
					// If JSON parsing fails, return raw data
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

	if err := s.scanner.Err(); err != nil {
		return nil, fmt.Errorf("read stream: %w", err)
	}

	// Handle last event if any
	if len(dataLines) > 0 {
		dataStr := strings.Join(dataLines, "\n")
		event.RawData = []byte(dataStr)
		if err := json.Unmarshal([]byte(dataStr), &event); err != nil {
			return &event, nil
		}
		if eventType != "" {
			event.Type = eventType
		}
		return &event, nil
	}

	return nil, io.EOF
}

// AnalyzeDataStream performs data analysis and returns a streaming response.
//
// This method sends a POST request to /byoa/api/v1/data_asking/analyze and
// returns a stream of Server-Sent Events (SSE) containing analysis results.
//
// The stream includes events such as:
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

	// Execute request
	resp, err := c.httpClient.Do(httpReq)
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
		Body:       resp.Body,
		Header:     resp.Header.Clone(),
		StatusCode: resp.StatusCode,
	}, nil
}
