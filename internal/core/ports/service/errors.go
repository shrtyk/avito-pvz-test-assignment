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

type ErrPvzNotExists struct {
	PvzId uuid.UUID
	Err   error
}

func (e *ErrPvzNotExists) Error() string {
	return fmt.Sprintf("pvz %s doesn't not exists", e.PvzId)
}

func (e *ErrPvzNotExists) Unwrap() error {
	return e.Err
}

type ErrNoOpenedReception struct {
	PvzId uuid.UUID
	Err   error
}

func (e *ErrNoOpenedReception) Error() string {
	return fmt.Sprintf("no opened reception in pvz %s", e.PvzId)
}

func (e *ErrNoOpenedReception) Unwrap() error {
	return e.Err
}
