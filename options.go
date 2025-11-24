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
	httpClient     *http.Client
	userAgent      string
	defaultHeaders http.Header
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
	headers   http.Header
	query     url.Values
	requestID string
}

func newCallOptions(opts ...CallOption) callOptions {
	co := callOptions{
		headers: make(http.Header),
		query:   make(url.Values),
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
