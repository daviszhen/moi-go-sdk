package sdk

import (
	"io"
	"net/http"
)

// FileStream wraps a streaming HTTP response body that callers must close.
//
// FileStream is returned by methods that download files or stream content.
// The caller is responsible for closing the Body to release resources.
//
// Example:
//
//	stream, err := client.DownloadGenAIResult(ctx, "file-id-123")
//	if err != nil {
//		return err
//	}
//	defer stream.Close()
//
//	data, err := io.ReadAll(stream.Body)
type FileStream struct {
	// Body is the response body that must be closed by the caller
	Body io.ReadCloser
	// Header contains the HTTP response headers
	Header http.Header
	// StatusCode is the HTTP status code
	StatusCode int
}

// Close releases the underlying HTTP response body.
//
// This should always be called when done with the FileStream to prevent
// resource leaks. It's safe to call Close multiple times.
func (s *FileStream) Close() error {
	if s == nil || s.Body == nil {
		return nil
	}
	return s.Body.Close()
}
