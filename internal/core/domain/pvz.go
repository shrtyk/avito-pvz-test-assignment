package domain

import (
	"time"

	"github.com/google/uuid"
)

type PVZCity string

const (
	Moscow          PVZCity = "Москва"
	SaintPetersburg PVZCity = "Санкт-Петербург"
	Kazan           PVZCity = "Казань"
)

type PVZ struct {
	Id               uuid.UUID
	City             PVZCity
	RegistrationDate time.Time
}
