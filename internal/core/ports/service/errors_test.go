package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServiceErr(t *testing.T) {
	var k ServiceErrKind = "test-kind"
	assert.Equal(t, "test-kind", k.String())
}
