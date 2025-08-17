package repository

import (
	"fmt"
)

type ErrConstraintViolation struct {
	Constraint string
	Err        error
}

func (e *ErrConstraintViolation) Error() string {
	return fmt.Sprintf("repository: constraint %s violated", e.Constraint)
}

func (e *ErrConstraintViolation) Unwrap() error {
	return e.Err
}

type ErrNoRowsInserted struct {
	Err error
}

func (e *ErrNoRowsInserted) Error() string {
	return e.Err.Error()
}

func (e *ErrNoRowsInserted) Unwrap() error {
	return e.Err
}

type ErrNullConstraint struct {
	Err error
}

func (e *ErrNullConstraint) Error() string {
	return e.Err.Error()
}

func (e *ErrNullConstraint) Unwrap() error {
	return e.Err
}
