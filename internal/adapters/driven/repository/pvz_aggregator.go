package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shrtyk/avito-backend-spring-2025/internal/core/domain"
)

type pvzRow struct {
	PvzID        string
	PvzCity      string
	PvzCreatedAt time.Time
	RecID        sql.NullString
	RecStatus    sql.NullString
	RecDateTime  sql.NullTime
	RecPvzID     sql.NullString
	ProdID       sql.NullString
	ProdDateTime sql.NullTime
	ProdRecID    sql.NullString
	ProdType     sql.NullString
}

func scanPvzRow(rows *sql.Rows) (pvzRow, error) {
	var row pvzRow
	err := rows.Scan(
		&row.PvzID, &row.PvzCity, &row.PvzCreatedAt,
		&row.RecID, &row.RecStatus, &row.RecDateTime, &row.RecPvzID,
		&row.ProdID, &row.ProdDateTime, &row.ProdRecID, &row.ProdType,
	)
	return row, err
}

type pvzAggregator struct {
	pvzMap  map[string]*domain.PvzReceptions
	recMap  map[string]*domain.ReceptionProducts
	ordPvzs []*domain.PvzReceptions
}

func newPvzAggregator() *pvzAggregator {
	return &pvzAggregator{
		pvzMap: make(map[string]*domain.PvzReceptions),
		recMap: make(map[string]*domain.ReceptionProducts),
	}
}

func (a *pvzAggregator) processRow(row pvzRow) error {
	if err := a.processPvz(row); err != nil {
		return err
	}

	recData, err := a.processReception(row)
	if err != nil {
		return err
	}

	return a.processProduct(row, recData)
}

func (a *pvzAggregator) processPvz(row pvzRow) error {
	if _, exists := a.pvzMap[row.PvzID]; exists {
		return nil
	}

	pvzUUID, err := uuid.Parse(row.PvzID)
	if err != nil {
		return fmt.Errorf("invalid PVZ UUID: %w", err)
	}

	pvz := &domain.Pvz{
		Id:               pvzUUID,
		City:             domain.PVZCity(row.PvzCity),
		RegistrationDate: row.PvzCreatedAt,
	}

	pvzData := &domain.PvzReceptions{
		Pvz:        pvz,
		Receptions: []*domain.ReceptionProducts{},
	}

	a.pvzMap[row.PvzID] = pvzData
	a.ordPvzs = append(a.ordPvzs, pvzData)
	return nil
}

func (a *pvzAggregator) processReception(row pvzRow) (*domain.ReceptionProducts, error) {
	if !row.RecID.Valid {
		return nil, nil
	}

	if recData, exists := a.recMap[row.RecID.String]; exists {
		return recData, nil
	}

	recUUID, err := parseUUID(row.RecID.String, "reception")
	if err != nil {
		return nil, err
	}

	recPvzUUID, err := parseUUID(row.RecPvzID.String, "reception PVZ")
	if err != nil {
		return nil, err
	}

	reception := &domain.Reception{
		Id:       recUUID,
		PvzId:    recPvzUUID,
		DateTime: row.RecDateTime.Time,
		Status:   domain.ReceptionStatus(row.RecStatus.String),
	}

	recData := &domain.ReceptionProducts{
		Reception: reception,
		Products:  []*domain.Product{},
	}

	a.recMap[row.RecID.String] = recData
	a.pvzMap[row.PvzID].Receptions = append(
		a.pvzMap[row.PvzID].Receptions,
		recData,
	)

	return recData, nil
}

func (a *pvzAggregator) processProduct(row pvzRow, recData *domain.ReceptionProducts) error {
	if recData == nil || !row.ProdID.Valid {
		return nil
	}

	prodUUID, err := parseUUID(row.ProdID.String, "product")
	if err != nil {
		return err
	}

	prodRecUUID, err := parseUUID(row.ProdRecID.String, "product reception")
	if err != nil {
		return err
	}

	product := &domain.Product{
		Id:          prodUUID,
		DateTime:    row.ProdDateTime.Time,
		ReceptionId: prodRecUUID,
		Type:        domain.ProductType(row.ProdType.String),
	}

	recData.Products = append(recData.Products, product)
	return nil
}

func (a *pvzAggregator) Results() []*domain.PvzReceptions {
	return a.ordPvzs
}

func parseUUID(val, context string) (uuid.UUID, error) {
	id, err := uuid.Parse(val)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid %s UUID '%s': %w", context, val, err)
	}
	return id, nil
}
