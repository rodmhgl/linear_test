// Package api provides an HTTP client for the LinkDing REST API.
// All requests are authenticated with a Bearer token and all non-2xx
// responses are converted into typed *ldcerr.Error values so that
// callers never have to inspect raw HTTP status codes.
package api

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	ldcerr "github.com/rodmhgl/ldctl/internal/errors"
)

// defaultTimeout is applied to every request when no custom client is provided.
const defaultTimeout = 30 * time.Second

// Client is an authenticated HTTP client for the LinkDing REST API.
// Create one with [New]; do not construct the zero value directly.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// Option is a functional option for [New].
type Option func(*Client)

// WithHTTPClient replaces the default HTTP client. Useful for testing with
// httptest servers or for injecting a transport with custom TLS settings.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		c.httpClient = hc
	}
}

// New creates a Client that talks to baseURL using the provided API token.
// baseURL must not have a trailing slash (e.g. "https://links.example.com").
func New(baseURL, token string, opts ...Option) *Client {
	c := &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// Do executes an HTTP request, sets the Authorization header, and maps any
// non-2xx response or transport error into a *ldcerr.Error. The caller is
// responsible for reading and closing resp.Body on success (non-nil return).
//
// On error the response body has already been drained and closed.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Token "+c.token)
	if req.Header.Get("Content-Type") == "" && req.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Distinguish network-level errors (DNS, connection refused, timeout)
		// from other transport errors.
		return nil, c.wrapTransportError(req.URL.String(), err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		return resp, nil
	}

	// Drain and close the body so the connection can be reused, then return a
	// typed error.
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	return nil, c.mapHTTPError(resp)
}

// Get is a convenience wrapper around [Do] for GET requests.
func (c *Client) Get(ctx context.Context, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(path), nil)
	if err != nil {
		return nil, ldcerr.NewIOError(path, err)
	}
	return c.Do(req)
}

// url resolves a relative API path against the client's base URL.
func (c *Client) url(path string) string {
	return fmt.Sprintf("%s%s", c.baseURL, path)
}

// wrapTransportError converts a transport-level error into a *ldcerr.Error.
// Network unreachable / connection refused errors become NetworkError; all
// others become NetworkError too since they all represent connectivity issues
// from the user's perspective.
func (c *Client) wrapTransportError(url string, err error) *ldcerr.Error {
	// Unwrap url.Error to check for net.Error (timeout / connection refused).
	if netErr, ok := err.(net.Error); ok {
		if netErr.Timeout() {
			return &ldcerr.Error{
				Type:    ldcerr.NetworkError,
				Message: fmt.Sprintf("request timed out connecting to %s", url),
				Details: map[string]interface{}{
					"url": url,
				},
				Cause: err,
			}
		}
	}
	return ldcerr.NewNetworkError(url, err)
}

// mapHTTPError converts a non-2xx HTTP response into a typed *ldcerr.Error.
// The URL is included in Details in addition to the HTTP status so that error
// messages can guide the user to the exact endpoint that failed.
func (c *Client) mapHTTPError(resp *http.Response) *ldcerr.Error {
	status := resp.StatusCode
	url := resp.Request.URL.String()

	switch {
	case status == http.StatusUnauthorized || status == http.StatusForbidden:
		return &ldcerr.Error{
			Type:    ldcerr.AuthError,
			Message: fmt.Sprintf("authentication failed (%d)", status),
			Details: map[string]interface{}{
				"http_status":  status,
				"url":          url,
				"instance_url": c.baseURL,
			},
		}
	case status == http.StatusNotFound:
		return &ldcerr.Error{
			Type:    ldcerr.NotFound,
			Message: "resource not found",
			Details: map[string]interface{}{
				"http_status": status,
				"url":         url,
			},
		}
	case status == http.StatusBadRequest:
		return &ldcerr.Error{
			Type:    ldcerr.ValidationError,
			Message: "validation failed: bad request",
			Details: map[string]interface{}{
				"http_status": status,
				"url":         url,
			},
		}
	case status == http.StatusTooManyRequests:
		return &ldcerr.Error{
			Type:    ldcerr.APIError,
			Message: "rate limit exceeded",
			Details: map[string]interface{}{
				"http_status": status,
				"url":         url,
				"retry_after": resp.Header.Get("Retry-After"),
			},
		}
	case status >= 500 && status <= 599:
		return ldcerr.NewAPIError(status, url)
	default:
		return &ldcerr.Error{
			Type:    ldcerr.APIError,
			Message: fmt.Sprintf("unexpected response (%d)", status),
			Details: map[string]interface{}{
				"http_status": status,
				"url":         url,
			},
		}
	}
}
