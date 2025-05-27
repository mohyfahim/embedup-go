package apiclient

import (
	// Make sure this import path is correct for your project structure.
	// If cstmerr is in 'your_module_path/internal/cstmerr', it would be:
	// "your_module_path/internal/cstmerr"
	// For now, using the path from your original code.
	"embedup-go/internal/cstmerr"
	"fmt"
	"strconv"
	"time"

	"resty.dev/v3"
)

// RestyAdapter implements the HTTPClient interface using the resty library.
type RestyAdapter struct {
	client *resty.Client
}

// NewRestyAdapter creates a new RestyAdapter with default transport settings.
// These settings mirror the ones from your original code.
func NewRestyAdapter() *RestyAdapter {
	transportSettings := &resty.TransportSettings{
		IdleConnTimeout:     30 * time.Second,
		TLSHandshakeTimeout: 60 * time.Second,
	}
	client := resty.NewWithTransportSettings(transportSettings)
	// You can enable Resty debugging if needed:
	// client.SetDebug(true)
	return &RestyAdapter{
		client: client,
	}
}

// NewRestyAdapterWithClient creates a new RestyAdapter using a pre-configured *resty.Client.
// This is useful if you need more customized Resty client settings.
func NewRestyAdapterWithClient(client *resty.Client) *RestyAdapter {
	if client == nil {
		// Fallback to default if nil client is passed, or panic, or return error
		return NewRestyAdapter()
	}
	return &RestyAdapter{client: client}
}

// buildRequest is a helper to configure a resty request from RequestOptions.
func (ra *RestyAdapter) buildRequest(baseRequest *resty.Request, opts *RequestOptions) *resty.Request {
	req := baseRequest
	if opts != nil {
		if opts.Headers != nil {
			req.SetHeaders(opts.Headers)
		}
		if opts.QueryParams != nil {
			req.SetQueryParams(opts.QueryParams)
		}
		if opts.Body != nil {
			req.SetBody(opts.Body) // Resty handles JSON marshaling for struct bodies by default
		}
		if opts.SuccessResult != nil {
			req.SetResult(opts.SuccessResult)
		}
		if opts.ErrorResult != nil {
			// Resty's SetError unmarshals the response body into ErrorResult if the HTTP status indicates an error.
			req.SetError(opts.ErrorResult)
		}
		// Note on opts.Timeout: Resty's client-level SetTimeout applies to the whole request-response cycle.
		// Per-request timeouts in Resty are typically handled via context.Context.
		// The transport settings provide robust connection/handshake timeouts.
		// If granular per-request timeout is needed, the interface and adapter would need enhancement (e.g., pass context).
		// For now, we rely on transport timeouts and any client-wide timeout set on the resty.Client.
	}
	return req
}

// Get implements the HTTPClient interface Get method.
func (ra *RestyAdapter) Get(url string, opts *RequestOptions) (*Response, error) {
	restyReq := ra.buildRequest(ra.client.R(), opts)
	restyResp, err := restyReq.Get(url)

	if err != nil { // Network errors, client-side timeouts before response, etc.
		return nil, cstmerr.NewAPIClientError(fmt.Errorf("HTTP GET request to %s failed: %w", url, err))
	}

	return &Response{
		StatusCode: restyResp.StatusCode(),
		Body:       restyResp.Bytes(), // Contains raw body bytes
		Headers:    restyResp.Header(),
		RequestURL: restyResp.Request.URL,
	}, nil
}

// Post implements the HTTPClient interface Post method.
func (ra *RestyAdapter) Post(url string, opts *RequestOptions) (*Response, error) {
	restyReq := ra.buildRequest(ra.client.R(), opts)
	restyResp, err := restyReq.Post(url)

	if err != nil {
		return nil, cstmerr.NewAPIClientError(fmt.Errorf("HTTP POST request to %s failed: %w", url, err))
	}

	return &Response{
		StatusCode: restyResp.StatusCode(),
		Body:       restyResp.Bytes(),
		Headers:    restyResp.Header(),
		RequestURL: restyResp.Request.URL,
	}, nil
}

// Put implements the HTTPClient interface Put method.
func (ra *RestyAdapter) Put(url string, opts *RequestOptions) (*Response, error) {
	restyReq := ra.buildRequest(ra.client.R(), opts)
	restyResp, err := restyReq.Put(url)

	if err != nil {
		return nil, cstmerr.NewAPIClientError(fmt.Errorf("HTTP PUT request to %s failed: %w", url, err))
	}

	return &Response{
		StatusCode: restyResp.StatusCode(),
		Body:       restyResp.Bytes(),
		Headers:    restyResp.Header(),
		RequestURL: restyResp.Request.URL,
	}, nil
}

// Head implements the HTTPClient interface Head method.
func (ra *RestyAdapter) Head(url string, opts *RequestOptions) (*Response, error) {
	// For HEAD, Body, SuccessResult, ErrorResult in opts are usually not applicable.
	// We only care about headers and status code.
	restyReq := ra.client.R() // Start with a fresh request
	if opts != nil {
		if opts.Headers != nil {
			restyReq.SetHeaders(opts.Headers)
		}
		if opts.QueryParams != nil {
			restyReq.SetQueryParams(opts.QueryParams)
		}
	}

	restyResp, err := restyReq.Head(url)
	if err != nil {
		return nil, cstmerr.NewHeadError(fmt.Sprintf("HTTP HEAD request to %s failed: %v", url, err))
	}

	return &Response{
		StatusCode: restyResp.StatusCode(),
		Body:       nil, // HEAD responses should not have a body processed by Body()
		Headers:    restyResp.Header(),
		RequestURL: restyResp.Request.URL,
	}, nil
}

// GetStream implements the HTTPClient interface GetStream method.
func (ra *RestyAdapter) GetStream(url string, opts *RequestOptions) (*StreamResponse, error) {
	restyReq := ra.client.R()
	if opts != nil {
		if opts.Headers != nil {
			restyReq.SetHeaders(opts.Headers)
		}
		if opts.QueryParams != nil {
			restyReq.SetQueryParams(opts.QueryParams)
		}
	}
	// Crucial for streaming: tell Resty not to parse or automatically close the response body.
	restyReq.SetDoNotParseResponse(true)

	restyResp, err := restyReq.Get(url)
	if err != nil {
		return nil, cstmerr.NewDownloadError(fmt.Sprintf("HTTP GET (stream) request to %s failed: %v", url, err))
	}

	// The caller is responsible for closing restyResp.RawResponse.Body
	// This body is an io.ReadCloser.
	contentLengthStr := restyResp.Header().Get("Content-Length")
	contentLength, _ := strconv.ParseInt(contentLengthStr, 10, 64) // Defaults to 0 if error or not present

	return &StreamResponse{
		StatusCode:    restyResp.StatusCode(),
		Body:          restyResp.RawResponse.Body,
		Headers:       restyResp.Header(),
		ContentLength: contentLength,
		RequestURL:    restyResp.Request.URL,
	}, nil
}
