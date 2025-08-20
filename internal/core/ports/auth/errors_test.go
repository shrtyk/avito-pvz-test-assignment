package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthErr(t *testing.T) {
	var k AuthErrKind = "test-kind"
	assert.Equal(t, "test-kind", k.String())
}
