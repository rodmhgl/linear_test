package errors_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

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
