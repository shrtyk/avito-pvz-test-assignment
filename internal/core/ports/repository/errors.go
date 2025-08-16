package repository

import (
	"fmt"
)

type ErrConstraintViolation struct {
	Constraint string
	Err        error
}

func (e *ErrConstraintViolation) Error() string {
	return fmt.Sprintf("constraint %s violated", e.Constraint)
}

func (e *ErrConstraintViolation) Unwrap() error {
	return e.Err
}
