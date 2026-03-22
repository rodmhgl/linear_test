package api_test

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rodmhgl/ldctl/internal/api"
	ldcerr "github.com/rodmhgl/ldctl/internal/errors"
)

// newTestServer creates an httptest server that always responds with the given
// HTTP status code. The server is closed automatically when the test ends.
func newTestServer(t *testing.T, status int) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// newTestServerWithHandler creates an httptest server with a custom handler.
func newTestServerWithHandler(t *testing.T, h http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return srv
}

// ldcErr extracts the *ldcerr.Error from err and fails the test if not found.
func ldcErr(t *testing.T, err error) *ldcerr.Error {
	t.Helper()
	var e *ldcerr.Error
	require.True(t, errors.As(err, &e), "expected *ldcerr.Error, got: %T %v", err, err)
	return e
}

// ── Constructor helpers ──────────────────────────────────────────────────────

func TestNew_SetsDefaults(t *testing.T) {
	c := api.New("https://links.example.com", "mytoken")
	require.NotNil(t, c)
}

// ── Authorization header ─────────────────────────────────────────────────────

func TestClient_Do_SetsAuthorizationHeader(t *testing.T) {
	var gotHeader string
	srv := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	})

	c := api.New(srv.URL, "secret-token")
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/bookmarks/", nil)
	require.NoError(t, err)

	resp, err := c.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, "Token secret-token", gotHeader)
}

// ── 2xx success ──────────────────────────────────────────────────────────────

func TestClient_Do_200_ReturnsResponse(t *testing.T) {
	srv := newTestServer(t, http.StatusOK)
	c := api.New(srv.URL, "tok")
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/", nil)
	require.NoError(t, err)

	resp, err := c.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestClient_Do_201_ReturnsResponse(t *testing.T) {
	srv := newTestServer(t, http.StatusCreated)
	c := api.New(srv.URL, "tok")
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL+"/", nil)
	require.NoError(t, err)

	resp, err := c.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

// ── 401 → AuthError ──────────────────────────────────────────────────────────

func TestClient_Do_401_IsAuthError(t *testing.T) {
	srv := newTestServer(t, http.StatusUnauthorized)
	c := api.New(srv.URL, "bad-token")
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/bookmarks/", nil)
	require.NoError(t, err)

	_, err = c.Do(req)
	require.Error(t, err)
	e := ldcErr(t, err)
	assert.Equal(t, ldcerr.AuthError, e.Type)
	assert.Equal(t, ldcerr.ExitConfigError, e.ExitCode())
	assert.Equal(t, 401, e.Details["http_status"])
}

func TestClient_Do_403_IsAuthError(t *testing.T) {
	srv := newTestServer(t, http.StatusForbidden)
	c := api.New(srv.URL, "bad-token")
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/bookmarks/", nil)
	require.NoError(t, err)

	_, err = c.Do(req)
	require.Error(t, err)
	e := ldcErr(t, err)
	assert.Equal(t, ldcerr.AuthError, e.Type)
	assert.Equal(t, ldcerr.ExitConfigError, e.ExitCode())
}

// ── 404 → NotFound ───────────────────────────────────────────────────────────

func TestClient_Do_404_IsNotFound(t *testing.T) {
	srv := newTestServer(t, http.StatusNotFound)
	c := api.New(srv.URL, "tok")
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/bookmarks/9999/", nil)
	require.NoError(t, err)

	_, err = c.Do(req)
	require.Error(t, err)
	e := ldcErr(t, err)
	assert.Equal(t, ldcerr.NotFound, e.Type)
	assert.Equal(t, ldcerr.ExitError, e.ExitCode())
	assert.Equal(t, 404, e.Details["http_status"])
}

// ── 400 → ValidationError ────────────────────────────────────────────────────

func TestClient_Do_400_IsValidationError(t *testing.T) {
	srv := newTestServer(t, http.StatusBadRequest)
	c := api.New(srv.URL, "tok")
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL+"/api/bookmarks/", nil)
	require.NoError(t, err)

	_, err = c.Do(req)
	require.Error(t, err)
	e := ldcErr(t, err)
	assert.Equal(t, ldcerr.ValidationError, e.Type)
	assert.Equal(t, ldcerr.ExitError, e.ExitCode())
	assert.Equal(t, 400, e.Details["http_status"])
}

// ── 429 → APIError ───────────────────────────────────────────────────────────

func TestClient_Do_429_IsAPIError(t *testing.T) {
	srv := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
	})
	c := api.New(srv.URL, "tok")
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/bookmarks/", nil)
	require.NoError(t, err)

	_, err = c.Do(req)
	require.Error(t, err)
	e := ldcErr(t, err)
	assert.Equal(t, ldcerr.APIError, e.Type)
	assert.Equal(t, ldcerr.ExitError, e.ExitCode())
	assert.Equal(t, "60", e.Details["retry_after"])
}

// ── 5xx → APIError ───────────────────────────────────────────────────────────

func TestClient_Do_500_IsAPIError(t *testing.T) {
	srv := newTestServer(t, http.StatusInternalServerError)
	c := api.New(srv.URL, "tok")
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/bookmarks/", nil)
	require.NoError(t, err)

	_, err = c.Do(req)
	require.Error(t, err)
	e := ldcErr(t, err)
	assert.Equal(t, ldcerr.APIError, e.Type)
	assert.Equal(t, ldcerr.ExitError, e.ExitCode())
	assert.Equal(t, 500, e.Details["http_status"])
}

func TestClient_Do_503_IsAPIError(t *testing.T) {
	srv := newTestServer(t, http.StatusServiceUnavailable)
	c := api.New(srv.URL, "tok")
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/", nil)
	require.NoError(t, err)

	_, err = c.Do(req)
	require.Error(t, err)
	e := ldcErr(t, err)
	assert.Equal(t, ldcerr.APIError, e.Type)
	assert.Equal(t, ldcerr.ExitError, e.ExitCode())
}

// ── Unknown non-2xx → APIError ───────────────────────────────────────────────

func TestClient_Do_418_IsAPIError(t *testing.T) {
	srv := newTestServer(t, http.StatusTeapot)
	c := api.New(srv.URL, "tok")
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/", nil)
	require.NoError(t, err)

	_, err = c.Do(req)
	require.Error(t, err)
	e := ldcErr(t, err)
	assert.Equal(t, ldcerr.APIError, e.Type)
}

// ── Network failure → NetworkError ───────────────────────────────────────────

func TestClient_Do_NetworkFailure_IsNetworkError(t *testing.T) {
	// Use a port on localhost that refuses connections. By binding and
	// immediately closing a listener we get a free port guaranteed to be
	// unused (and therefore refusing connections).
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := ln.Addr().String()
	ln.Close() // immediately close — port now refuses connections

	c := api.New("http://"+addr, "tok")
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://"+addr+"/", nil)
	require.NoError(t, err)

	_, err = c.Do(req)
	require.Error(t, err)
	e := ldcErr(t, err)
	assert.Equal(t, ldcerr.NetworkError, e.Type)
	assert.Equal(t, ldcerr.ExitConfigError, e.ExitCode())
}

func TestClient_Do_Timeout_IsNetworkError(t *testing.T) {
	// Server that never responds within the client timeout.
	srv := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		// Block long enough for the client to time out.
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	})

	hc := &http.Client{Timeout: 50 * time.Millisecond}
	c := api.New(srv.URL, "tok", api.WithHTTPClient(hc))
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/", nil)
	require.NoError(t, err)

	_, err = c.Do(req)
	require.Error(t, err)
	e := ldcErr(t, err)
	assert.Equal(t, ldcerr.NetworkError, e.Type)
	assert.Equal(t, ldcerr.ExitConfigError, e.ExitCode())
}

// ── Details contain URL ───────────────────────────────────────────────────────

func TestClient_Do_ErrorDetails_ContainURL(t *testing.T) {
	srv := newTestServer(t, http.StatusNotFound)
	c := api.New(srv.URL, "tok")
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/bookmarks/42/", nil)
	require.NoError(t, err)

	_, err = c.Do(req)
	require.Error(t, err)
	e := ldcErr(t, err)
	assert.NotEmpty(t, e.Details["url"], "error details should contain the request URL")
}

// ── AuthError details contain instance_url ───────────────────────────────────

func TestClient_Do_AuthError_ContainsInstanceURL(t *testing.T) {
	srv := newTestServer(t, http.StatusUnauthorized)
	c := api.New(srv.URL, "bad")
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/bookmarks/", nil)
	require.NoError(t, err)

	_, err = c.Do(req)
	require.Error(t, err)
	e := ldcErr(t, err)
	assert.Equal(t, srv.URL, e.Details["instance_url"])
}

// ── Get convenience method ───────────────────────────────────────────────────

func TestClient_Get_Success(t *testing.T) {
	srv := newTestServer(t, http.StatusOK)
	c := api.New(srv.URL, "tok")

	resp, err := c.Get(context.Background(), "/api/bookmarks/")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestClient_Get_MapsErrors(t *testing.T) {
	srv := newTestServer(t, http.StatusUnauthorized)
	c := api.New(srv.URL, "bad")

	_, err := c.Get(context.Background(), "/api/bookmarks/")
	require.Error(t, err)
	e := ldcErr(t, err)
	assert.Equal(t, ldcerr.AuthError, e.Type)
}
