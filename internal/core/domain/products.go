package domain

import (
	"time"

	"github.com/google/uuid"
)

type ProductType string

const (
	ProductTypeClothing    ProductType = "одежда"
	ProductTypeElectronics ProductType = "электроника"
	ProductTypeFootwear    ProductType = "обувь"
)

type Product struct {
	Id          uuid.UUID
	PvzId       uuid.UUID
	Type        ProductType
	ReceptionId uuid.UUID
	DateTime    time.Time
}
