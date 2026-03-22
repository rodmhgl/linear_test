package errors_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ldcerr "github.com/rodmhgl/ldctl/internal/errors"
)

func TestExitCodeConstants(t *testing.T) {
	assert.Equal(t, 0, ldcerr.ExitSuccess)
	assert.Equal(t, 1, ldcerr.ExitError)
	assert.Equal(t, 2, ldcerr.ExitConfigError)
}

func TestErrorTypes_StringValues(t *testing.T) {
	assert.Equal(t, ldcerr.Type("config_error"), ldcerr.ConfigError)
	assert.Equal(t, ldcerr.Type("auth_error"), ldcerr.AuthError)
	assert.Equal(t, ldcerr.Type("network_error"), ldcerr.NetworkError)
	assert.Equal(t, ldcerr.Type("api_error"), ldcerr.APIError)
	assert.Equal(t, ldcerr.Type("validation_error"), ldcerr.ValidationError)
	assert.Equal(t, ldcerr.Type("io_error"), ldcerr.IOError)
	assert.Equal(t, ldcerr.Type("not_found"), ldcerr.NotFound)
	assert.Equal(t, ldcerr.Type("user_cancelled"), ldcerr.UserCancelled)
}

func TestError_ImplementsErrorInterface(t *testing.T) {
	var err error = &ldcerr.Error{Type: ldcerr.APIError, Message: "boom"}
	assert.Equal(t, "boom", err.Error())
}

func TestError_ExitCode_ConfigTypes_Return2(t *testing.T) {
	for _, typ := range []ldcerr.Type{ldcerr.ConfigError, ldcerr.AuthError, ldcerr.NetworkError} {
		e := &ldcerr.Error{Type: typ}
		assert.Equal(t, 2, e.ExitCode(), "expected exit 2 for type %s", typ)
	}
}

func TestError_ExitCode_OperationalTypes_Return1(t *testing.T) {
	for _, typ := range []ldcerr.Type{
		ldcerr.APIError, ldcerr.ValidationError,
		ldcerr.IOError, ldcerr.NotFound, ldcerr.UserCancelled,
	} {
		e := &ldcerr.Error{Type: typ}
		assert.Equal(t, 1, e.ExitCode(), "expected exit 1 for type %s", typ)
	}
}

func TestError_Unwrap_ReturnsCause(t *testing.T) {
	cause := fmt.Errorf("root cause")
	e := &ldcerr.Error{Type: ldcerr.IOError, Message: "wrapper", Cause: cause}
	assert.Equal(t, cause, e.Unwrap())
}

func TestNewConfigNotFound(t *testing.T) {
	e := ldcerr.NewConfigNotFound("/home/user/.config/ldctl/config.toml")
	assert.Equal(t, ldcerr.ConfigError, e.Type)
	assert.Equal(t, "no configuration found", e.Message)
	assert.Equal(t, "/home/user/.config/ldctl/config.toml", e.Details["config_path"])
	assert.NotEmpty(t, e.Details["suggestion"])
	assert.Equal(t, 2, e.ExitCode())
}

func TestNewAuthFailed(t *testing.T) {
	e := ldcerr.NewAuthFailed("https://links.example.com", 401)
	assert.Equal(t, ldcerr.AuthError, e.Type)
	assert.Equal(t, "authentication failed", e.Message)
	assert.Equal(t, 401, e.Details["http_status"])
	assert.Equal(t, "https://links.example.com", e.Details["instance_url"])
	assert.Equal(t, 2, e.ExitCode())
}

func TestNewNotFound(t *testing.T) {
	e := ldcerr.NewNotFound("bookmark", 42)
	assert.Equal(t, ldcerr.NotFound, e.Type)
	assert.Equal(t, "bookmark not found", e.Message)
	assert.Equal(t, "bookmark", e.Details["resource"])
	assert.Equal(t, 42, e.Details["id"])
	assert.Equal(t, 1, e.ExitCode())
}

func TestPrintHuman_BasicFormat(t *testing.T) {
	var buf bytes.Buffer
	e := &ldcerr.Error{Type: ldcerr.APIError, Message: "bookmark not found"}
	ldcerr.PrintHuman(&buf, e)
	assert.Equal(t, "Error: bookmark not found\n", buf.String())
}

func TestPrintHuman_WithHint(t *testing.T) {
	var buf bytes.Buffer
	e := ldcerr.NewConfigNotFound("/cfg/config.toml")
	ldcerr.PrintHuman(&buf, e)
	out := buf.String()
	assert.Contains(t, out, "Error: no configuration found")
	assert.Contains(t, out, "ldctl config init")
}

func TestPrintJSON_ValidEnvelope(t *testing.T) {
	var buf bytes.Buffer
	e := &ldcerr.Error{
		Type:    ldcerr.APIError,
		Message: "bookmark not found",
		Details: map[string]interface{}{"http_status": 404},
	}
	err := ldcerr.PrintJSON(&buf, e)
	require.NoError(t, err)

	var envelope struct {
		Error struct {
			Type    string                 `json:"type"`
			Message string                 `json:"message"`
			Details map[string]interface{} `json:"details"`
		} `json:"error"`
	}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &envelope))
	assert.Equal(t, "api_error", envelope.Error.Type)
	assert.Equal(t, "bookmark not found", envelope.Error.Message)
	assert.InDelta(t, 404, envelope.Error.Details["http_status"], 0)
}

func TestPrintJSON_NilDetails_OmitsField(t *testing.T) {
	var buf bytes.Buffer
	e := &ldcerr.Error{Type: ldcerr.IOError, Message: "write failed"}
	require.NoError(t, ldcerr.PrintJSON(&buf, e))

	var raw map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &raw))
	errObj := raw["error"].(map[string]interface{})
	_, hasDetails := errObj["details"]
	assert.False(t, hasDetails, "details key should be absent when nil")
}

func TestMapHTTPError_401_IsAuthError(t *testing.T) {
	resp := &http.Response{StatusCode: 401, Header: make(http.Header)}
	e := ldcerr.MapHTTPError(resp)
	assert.Equal(t, ldcerr.AuthError, e.Type)
	assert.Equal(t, 2, e.ExitCode())
}

func TestMapHTTPError_403_IsAuthError(t *testing.T) {
	resp := &http.Response{StatusCode: 403, Header: make(http.Header)}
	e := ldcerr.MapHTTPError(resp)
	assert.Equal(t, ldcerr.AuthError, e.Type)
}

func TestMapHTTPError_404_IsNotFound(t *testing.T) {
	resp := &http.Response{StatusCode: 404, Header: make(http.Header)}
	e := ldcerr.MapHTTPError(resp)
	assert.Equal(t, ldcerr.NotFound, e.Type)
	assert.Equal(t, 1, e.ExitCode())
}

func TestMapHTTPError_400_IsValidationError(t *testing.T) {
	resp := &http.Response{StatusCode: 400, Header: make(http.Header)}
	e := ldcerr.MapHTTPError(resp)
	assert.Equal(t, ldcerr.ValidationError, e.Type)
}

func TestMapHTTPError_500_IsAPIError(t *testing.T) {
	resp := &http.Response{StatusCode: 500, Header: make(http.Header)}
	e := ldcerr.MapHTTPError(resp)
	assert.Equal(t, ldcerr.APIError, e.Type)
	assert.Equal(t, 1, e.ExitCode())
}

func TestMapHTTPError_UnknownCode_IsAPIError(t *testing.T) {
	resp := &http.Response{StatusCode: 418, Header: make(http.Header)}
	e := ldcerr.MapHTTPError(resp)
	assert.Equal(t, ldcerr.APIError, e.Type)
}
