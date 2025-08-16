package domain

import (
	"time"

	"github.com/google/uuid"
)

type ReceptionStatus string

const (
	Close      ReceptionStatus = "close"
	InProgress ReceptionStatus = "in_progress"
)

type Reception struct {
	Id       uuid.UUID
	PvzId    uuid.UUID
	DateTime time.Time
	Status   ReceptionStatus
}
