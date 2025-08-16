package service

import (
	"fmt"

	"github.com/google/uuid"
)

type ErrReceptionInProgress struct {
	PvzId uuid.UUID
	Err   error
}

func (e *ErrReceptionInProgress) Error() string {
	return fmt.Sprintf("reception for pvz %s is already in progress", e.PvzId)
}

func (e *ErrReceptionInProgress) Unwrap() error {
	return e.Err
}
