package errors_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	ldcerr "github.com/rodmhgl/ldctl/internal/errors"
)

func TestExitCodeConstants(t *testing.T) {
	assert.Equal(t, 0, ldcerr.ExitSuccess)
	assert.Equal(t, 1, ldcerr.ExitError)
	assert.Equal(t, 2, ldcerr.ExitConfigError)
}
