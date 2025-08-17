package xerr

import (
	"fmt"
)

type Stringer interface {
	String() string
}

type BaseErr[T Stringer] struct {
	Op   string
	Kind T
	Err  error
}

func NewErr[T Stringer](op string, kind T, err error) error {
	return &BaseErr[T]{Op: op, Kind: kind, Err: err}
}

func (e *BaseErr[T]) Error() string {
	return fmt.Sprintf("Op: %s, Kind: %s, Error: %s", e.Op, e.Kind, e.Err)
}

func (e *BaseErr[T]) Unwrap() error {
	return e.Err
}
