package xerr_test

import (
	"errors"
	"fmt"
	"testing"

	xerr "github.com/shrtyk/pvz-service/pkg/xerrors"
	"github.com/stretchr/testify/assert"
)

type testKind string

const (
	testKind1 testKind = "test1"
	testKind2 testKind = "test2"
)

func (t testKind) String() string {
	return string(t)
}

func TestErrors(t *testing.T) {
	op := "errors.Test"

	nerr := errors.New("new error")

	e1 := xerr.NewErr(op, testKind1)
	e2 := xerr.WrapErr(op, testKind2, nerr)

	exp1 := fmt.Sprintf("Op: %s, Kind: %s, Error: %s", op, testKind1, string(testKind1))
	assert.Equal(t, exp1, e1.Error())

	exp2 := fmt.Sprintf("Op: %s, Kind: %s, Error: %s", op, testKind2, nerr)
	assert.Equal(t, exp2, e2.Error())

	assert.True(t, errors.Is(e2, nerr))
}
