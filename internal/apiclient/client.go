package apiclient

import (
	"io"
	"net/http"
	"time"
)

// RequestOptions holds options for an HTTP request.
type RequestOptions struct {
	Headers       map[string]string
	QueryParams   map[string]string
	Body          any           // For POST, PUT, PATCH - will be JSON marshaled by adapter
	SuccessResult any           // Pointer to struct to unmarshal success JSON response
	ErrorResult   any           // Pointer to struct to unmarshal error JSON response
	Timeout       time.Duration // Optional per-request timeout (behavior depends on adapter)
}

// Response represents a general HTTP response.
type Response struct {
	StatusCode int
	Body       []byte      // Raw response body
	Headers    http.Header // Standard http.Header
	RequestURL string      // The URL that was requested
}

// IsSuccess checks if the status code is in the 2xx range.
func (r *Response) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// IsError checks if the status code indicates an error (>= 400).
func (r *Response) IsError() bool {
	return r.StatusCode >= 400
}

// StreamResponse is for responses where the body is streamed, e.g., file downloads.
type StreamResponse struct {
	StatusCode    int
	Body          io.ReadCloser // The response body stream; caller must close it.
	Headers       http.Header
	ContentLength int64  // Content-Length from header, or -1 if not available/applicable
	RequestURL    string // The URL that was requested
}

// IsSuccess checks if the status code is in the 2xx range.
func (sr *StreamResponse) IsSuccess() bool {
	return sr.StatusCode >= 200 && sr.StatusCode < 300
}

// HTTPClient defines the interface for a generic HTTP client.
// Implementations of this interface will handle the actual HTTP communication.
type HTTPClient interface {
	// Get performs an HTTP GET request.
	// If opts.SuccessResult is provided, the adapter will attempt to unmarshal a successful response into it.
	// If opts.ErrorResult is provided, the adapter will attempt to unmarshal an error response into it.
	Get(url string, opts *RequestOptions) (*Response, error)

	// Post performs an HTTP POST request.
	// opts.Body will typically be marshaled to JSON by the adapter.
	Post(url string, opts *RequestOptions) (*Response, error)

	// Put performs an HTTP PUT request.
	// opts.Body will typically be marshaled to JSON by the adapter.
	Put(url string, opts *RequestOptions) (*Response, error)

	// Head performs an HTTP HEAD request.
	// Typically used to get headers without fetching the body.
	Head(url string, opts *RequestOptions) (*Response, error)

	// GetStream performs an HTTP GET request and returns a response with a body stream.
	// This is suitable for downloading large files. The caller is responsible for closing StreamResponse.Body.
	GetStream(url string, opts *RequestOptions) (*StreamResponse, error)
}
