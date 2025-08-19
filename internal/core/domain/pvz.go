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

type Pvz struct {
	Id               uuid.UUID
	City             PVZCity
	RegistrationDate time.Time
}

type PvzsReadParams struct {
	StartDate *time.Time
	EndDate   *time.Time
	Page      int
	Limit     int
}

type PvzReceptions struct {
	Pvz        *Pvz
	Receptions []*ReceptionProducts
}

type ReceptionProducts struct {
	Reception *Reception
	Products  []*Product
}
