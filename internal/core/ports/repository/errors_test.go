package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRepoErr(t *testing.T) {
	var k RepoErrKind = "test-kind"
	assert.Equal(t, "test-kind", k.String())
}
