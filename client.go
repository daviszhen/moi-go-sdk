// Package sdk provides a Go client library for interacting with the MOI Catalog Service.
//
// The package provides two types of clients:
//   - RawClient: Low-level client that provides direct access to API endpoints
//   - SDKClient: High-level client that provides convenient business-oriented APIs
//
// Example usage:
//
//	// Create a raw client
//	client, err := sdk.NewRawClient("https://api.example.com", "your-api-key")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Create a catalog
//	resp, err := client.CreateCatalog(ctx, &sdk.CatalogCreateRequest{
//		CatalogName: "my-catalog",
//		Comment:     "My catalog",
//	})
//
// For more information, see the documentation at https://github.com/matrixorigin/moi-go-sdk/docs
package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	headerAPIKey      = "moi-key"
	headerRequestID   = "X-Request-ID"
	headerUserAgent   = "User-Agent"
	headerContentType = "Content-Type"
	headerAccept      = "Accept"

	mimeJSON = "application/json"
)

// RawClient provides typed access to the catalog service HTTP APIs.
type RawClient struct {
	baseURL        string
	apiKey         string
	httpClient     *http.Client
	userAgent      string
	defaultHeaders http.Header
}

// NewRawClient creates a new client using the provided baseURL and apiKey.
// opts can be used to customize the underlying HTTP client behaviour.
func NewRawClient(baseURL, apiKey string, opts ...ClientOption) (*RawClient, error) {
	trimmedBase := strings.TrimSpace(baseURL)
	if trimmedBase == "" {
		return nil, ErrBaseURLRequired
	}
	trimmedKey := strings.TrimSpace(apiKey)
	if trimmedKey == "" {
		return nil, ErrAPIKeyRequired
	}

	parsed, err := url.Parse(trimmedBase)
	if err != nil {
		return nil, fmt.Errorf("invalid baseURL: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("baseURL must include scheme and host")
	}
	parsed.RawQuery = ""
	parsed.Fragment = ""
	normalized := strings.TrimRight(parsed.String(), "/")

	cfg := clientOptions{
		userAgent:      defaultUserAgent,
		defaultHeaders: make(http.Header),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	httpClient := cfg.httpClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultHTTPTimeout}
	}
	if cfg.defaultHeaders == nil {
		cfg.defaultHeaders = make(http.Header)
	}

	return &RawClient{
		baseURL:        normalized,
		apiKey:         trimmedKey,
		httpClient:     httpClient,
		userAgent:      cfg.userAgent,
		defaultHeaders: cloneHeader(cfg.defaultHeaders),
	}, nil
}

// postJSON issues a JSON request and decodes the enveloped response payload.
func (c *RawClient) postJSON(ctx context.Context, path string, reqBody interface{}, respBody interface{}, opts ...CallOption) error {
	return c.doJSON(ctx, http.MethodPost, path, reqBody, respBody, opts...)
}

// getJSON issues a JSON GET request and decodes the enveloped response payload.
func (c *RawClient) getJSON(ctx context.Context, path string, respBody interface{}, opts ...CallOption) error {
	return c.doJSON(ctx, http.MethodGet, path, nil, respBody, opts...)
}

func (c *RawClient) doJSON(ctx context.Context, method, path string, body interface{}, respBody interface{}, opts ...CallOption) error {
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

	resp, err := c.doRaw(ctx, method, path, reader, callOpts, func(req *http.Request) {
		req.Header.Set(headerAccept, mimeJSON)
		if body != nil {
			req.Header.Set(headerContentType, mimeJSON)
		}
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var envelope apiEnvelope
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&envelope); err != nil {
		// Check if response body is empty
		if err == io.EOF {
			return fmt.Errorf("empty response body")
		}
		return fmt.Errorf("decode response: %w", err)
	}

	if envelope.Code != "" && envelope.Code != "OK" {
		return &APIError{
			Code:       envelope.Code,
			Message:    envelope.Msg,
			RequestID:  envelope.RequestID,
			HTTPStatus: resp.StatusCode,
		}
	}

	if respBody != nil && len(envelope.Data) > 0 && string(envelope.Data) != "null" {
		if err := json.Unmarshal(envelope.Data, respBody); err != nil {
			return fmt.Errorf("decode data field: %w", err)
		}
	}
	return nil
}

func (c *RawClient) doRaw(ctx context.Context, method, path string, body io.Reader, opts callOptions, prepare func(*http.Request)) (*http.Response, error) {
	req, err := c.buildRequest(ctx, method, path, body, opts)
	if err != nil {
		return nil, err
	}
	if prepare != nil {
		prepare(req)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		data, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, &HTTPError{StatusCode: resp.StatusCode, Body: data}
	}
	return resp, nil
}

func (c *RawClient) buildRequest(ctx context.Context, method, path string, body io.Reader, opts callOptions) (*http.Request, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if path == "" {
		return nil, fmt.Errorf("request path cannot be empty")
	}
	fullURL := c.baseURL + ensureLeadingSlash(path)
	if len(opts.query) > 0 {
		delimiter := "?"
		if strings.Contains(fullURL, "?") {
			delimiter = "&"
		}
		fullURL = fullURL + delimiter + opts.query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set(headerAPIKey, c.apiKey)
	if c.userAgent != "" {
		req.Header.Set(headerUserAgent, c.userAgent)
	}
	mergeHeaders(req.Header, c.defaultHeaders, false)
	if opts.requestID != "" {
		req.Header.Set(headerRequestID, opts.requestID)
	}
	mergeHeaders(req.Header, opts.headers, true)
	return req, nil
}

func ensureLeadingSlash(p string) string {
	if strings.HasPrefix(p, "/") {
		return p
	}
	return "/" + p
}
