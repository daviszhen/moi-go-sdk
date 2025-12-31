package sdk

import (
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultUserAgent   = "matrixflow-sdk-go/0.1.0"
	defaultHTTPTimeout = 30 * time.Second
)

type clientOptions struct {
	httpClient      *http.Client
	userAgent       string
	defaultHeaders  http.Header
	llmProxyBaseURL string // Optional: direct LLM Proxy base URL for direct connection
}

// ClientOption customizes the SDK client during construction.
//
// ClientOption functions are used with NewRawClient to configure the client
// behavior, such as HTTP timeout, custom HTTP client, or default headers.
type ClientOption func(*clientOptions)

// WithHTTPClient overrides the default http.Client used by the SDK.
//
// This allows you to provide a custom HTTP client with specific configuration,
// such as custom transport, connection pooling, or proxy settings.
//
// Example:
//
//	customClient := &http.Client{
//		Timeout: 60 * time.Second,
//	}
//	client, err := sdk.NewRawClient(baseURL, apiKey, sdk.WithHTTPClient(customClient))
func WithHTTPClient(client *http.Client) ClientOption {
	return func(o *clientOptions) {
		if client != nil {
			o.httpClient = client
		}
	}
}

// WithHTTPTimeout configures the timeout on the underlying http.Client.
//
// The timeout applies to the entire request, including connection establishment,
// request sending, and response reading.
//
// Example:
//
//	client, err := sdk.NewRawClient(baseURL, apiKey,
//		sdk.WithHTTPTimeout(60 * time.Second))
func WithHTTPTimeout(timeout time.Duration) ClientOption {
	return func(o *clientOptions) {
		if timeout <= 0 {
			return
		}
		if o.httpClient == nil {
			o.httpClient = &http.Client{}
		}
		o.httpClient.Timeout = timeout
	}
}

// WithUserAgent overrides the default User-Agent header that is sent with every request.
//
// The default User-Agent is "matrixflow-sdk-go/0.1.0".
//
// Example:
//
//	client, err := sdk.NewRawClient(baseURL, apiKey,
//		sdk.WithUserAgent("my-app/1.0.0"))
func WithUserAgent(userAgent string) ClientOption {
	return func(o *clientOptions) {
		ua := strings.TrimSpace(userAgent)
		if ua != "" {
			o.userAgent = ua
		}
	}
}

// WithDefaultHeader adds a header that will be included on every request.
//
// Headers added via WithDefaultHeader are sent with all API calls made by the client.
//
// Example:
//
//	client, err := sdk.NewRawClient(baseURL, apiKey,
//		sdk.WithDefaultHeader("X-Custom-Header", "value"))
func WithDefaultHeader(key, value string) ClientOption {
	return func(o *clientOptions) {
		if key == "" {
			return
		}
		if o.defaultHeaders == nil {
			o.defaultHeaders = make(http.Header)
		}
		o.defaultHeaders.Add(key, value)
	}
}

// WithDefaultHeaders merges a set of headers that will be included on every request.
//
// This is useful when you need to add multiple default headers at once.
//
// Example:
//
//	headers := http.Header{}
//	headers.Set("X-Custom-1", "value1")
//	headers.Set("X-Custom-2", "value2")
//	client, err := sdk.NewRawClient(baseURL, apiKey,
//		sdk.WithDefaultHeaders(headers))
func WithDefaultHeaders(headers http.Header) ClientOption {
	return func(o *clientOptions) {
		if len(headers) == 0 {
			return
		}
		if o.defaultHeaders == nil {
			o.defaultHeaders = make(http.Header)
		}
		mergeHeaders(o.defaultHeaders, headers, false)
	}
}

// WithLLMProxyBaseURL sets the base URL for direct connection to LLM Proxy.
//
// When set, you can use WithDirectLLMProxy in CallOption to directly connect
// to LLM Proxy instead of going through the MOI SDK gateway (which adds /llm-proxy prefix).
//
// Example:
//
//	client, err := sdk.NewRawClient(baseURL, apiKey,
//		sdk.WithLLMProxyBaseURL("https://llm-proxy.example.com"))
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Use direct connection
//	session, err := client.CreateLLMSession(ctx, req,
//		sdk.WithDirectLLMProxy())
func WithLLMProxyBaseURL(baseURL string) ClientOption {
	return func(o *clientOptions) {
		trimmed := strings.TrimSpace(baseURL)
		if trimmed != "" {
			parsed, err := url.Parse(trimmed)
			if err == nil && parsed.Scheme != "" && parsed.Host != "" {
				parsed.RawQuery = ""
				parsed.Fragment = ""
				o.llmProxyBaseURL = strings.TrimRight(parsed.String(), "/")
			}
		}
	}
}

// CallOption customizes individual SDK operations.
//
// CallOption functions are used with individual API method calls to customize
// request behavior, such as adding headers, query parameters, or request IDs.
//
// Example:
//
//	resp, err := client.CreateCatalog(ctx, req,
//		sdk.WithRequestID("my-request-id"),
//		sdk.WithHeader("X-Custom", "value"))
type CallOption func(*callOptions)

type callOptions struct {
	headers           http.Header
	query             url.Values
	requestID         string
	useDirectLLMProxy bool // Whether to use direct LLM Proxy connection
	streamBufferSize  int  // Buffer size for stream scanner (in bytes)
}

func newCallOptions(opts ...CallOption) callOptions {
	co := callOptions{
		headers:          make(http.Header),
		query:            make(url.Values),
		streamBufferSize: 0, // 0 means use default
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&co)
		}
	}
	return co
}

// WithRequestID sets the X-Request-ID header on the outgoing request.
//
// The request ID is useful for tracking and debugging requests on the server side.
//
// Example:
//
//	resp, err := client.CreateCatalog(ctx, req,
//		sdk.WithRequestID("catalog-create-001"))
func WithRequestID(id string) CallOption {
	return func(co *callOptions) {
		co.requestID = strings.TrimSpace(id)
	}
}

// WithHeader sets or overrides a header on the outgoing request.
//
// Headers set via WithHeader will override default headers and any headers
// set via WithDefaultHeader for this specific request.
//
// Example:
//
//	resp, err := client.CreateCatalog(ctx, req,
//		sdk.WithHeader("X-Custom-Header", "value"))
func WithHeader(key, value string) CallOption {
	return func(co *callOptions) {
		if key == "" {
			return
		}
		if co.headers == nil {
			co.headers = make(http.Header)
		}
		co.headers.Set(key, value)
	}
}

// WithHeaders merges headers into the outgoing request.
//
// This is useful when you need to add multiple headers to a single request.
//
// Example:
//
//	headers := http.Header{}
//	headers.Set("X-Custom-1", "value1")
//	headers.Set("X-Custom-2", "value2")
//	resp, err := client.CreateCatalog(ctx, req, sdk.WithHeaders(headers))
func WithHeaders(headers http.Header) CallOption {
	return func(co *callOptions) {
		if len(headers) == 0 {
			return
		}
		if co.headers == nil {
			co.headers = make(http.Header)
		}
		mergeHeaders(co.headers, headers, false)
	}
}

// WithQueryParam appends a single query parameter to the request URL.
//
// Multiple calls to WithQueryParam will append multiple parameters.
//
// Example:
//
//	resp, err := client.ListCatalogs(ctx,
//		sdk.WithQueryParam("page", "1"),
//		sdk.WithQueryParam("size", "10"))
func WithQueryParam(key, value string) CallOption {
	return func(co *callOptions) {
		if key == "" {
			return
		}
		if co.query == nil {
			co.query = make(url.Values)
		}
		co.query.Add(key, value)
	}
}

// WithQuery merges an entire query parameter map into the request URL.
func WithQuery(values url.Values) CallOption {
	return func(co *callOptions) {
		if len(values) == 0 {
			return
		}
		if co.query == nil {
			co.query = make(url.Values)
		}
		for key, vv := range values {
			for _, v := range vv {
				co.query.Add(key, v)
			}
		}
	}
}

// WithDirectLLMProxy enables direct connection to LLM Proxy instead of going through the MOI SDK gateway.
//
// When this option is used, the request will be sent directly to the LLM Proxy base URL
// (configured via WithLLMProxyBaseURL) without the /llm-proxy prefix.
//
// This option only works if WithLLMProxyBaseURL was set during client initialization.
// If llmProxyBaseURL is not configured, this option will be ignored and the request
// will fall back to the default behavior (through MOI SDK gateway with /llm-proxy prefix).
//
// Example:
//
//	// Initialize client with LLM Proxy base URL
//	client, err := sdk.NewRawClient(baseURL, apiKey,
//		sdk.WithLLMProxyBaseURL("https://llm-proxy.example.com"))
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Use direct connection
//	session, err := client.CreateLLMSession(ctx, req,
//		sdk.WithDirectLLMProxy())
func WithDirectLLMProxy() CallOption {
	return func(co *callOptions) {
		co.useDirectLLMProxy = true
	}
}

// WithStreamBufferSize sets the initial buffer size for stream reader.
//
// The buffer will dynamically grow as needed to handle lines of arbitrary length,
// so this option only sets the initial buffer size for better performance.
// If not set, the default initial buffer size is 4KB.
//
// The size is specified in bytes. A larger initial buffer can improve performance
// for streams with consistently large lines, but is not required for correctness.
//
// Example:
//
//	stream, err := client.AnalyzeDataStream(ctx, req,
//		sdk.WithStreamBufferSize(64*1024)) // 64KB initial buffer
func WithStreamBufferSize(size int) CallOption {
	return func(co *callOptions) {
		if size > 0 {
			co.streamBufferSize = size
		}
	}
}

func cloneHeader(src http.Header) http.Header {
	if len(src) == 0 {
		return make(http.Header)
	}
	dst := make(http.Header, len(src))
	for k, vv := range src {
		copied := make([]string, len(vv))
		copy(copied, vv)
		dst[k] = copied
	}
	return dst
}

func mergeHeaders(dst, src http.Header, override bool) {
	if len(src) == 0 {
		return
	}
	for k, vv := range src {
		if override {
			dst.Del(k)
		}
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
