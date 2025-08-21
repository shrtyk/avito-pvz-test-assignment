package repository

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func Test_pvzAggregator(t *testing.T) {
	t.Parallel()
	pvzID1 := uuid.New()
	recID1 := uuid.New()
	prodID1 := uuid.New()

	pvzID2 := uuid.New()
	recID2 := uuid.New()
	prodID2 := uuid.New()

	rows := []pvzRow{
		{
			PvzID:        pvzID1.String(),
			PvzCity:      "Moscow",
			PvzCreatedAt: time.Now(),
			RecID:        sql.NullString{String: recID1.String(), Valid: true},
			RecStatus:    sql.NullString{String: "opened", Valid: true},
			RecDateTime:  sql.NullTime{Time: time.Now(), Valid: true},
			RecPvzID:     sql.NullString{String: pvzID1.String(), Valid: true},
			ProdID:       sql.NullString{String: prodID1.String(), Valid: true},
			ProdDateTime: sql.NullTime{Time: time.Now(), Valid: true},
			ProdRecID:    sql.NullString{String: recID1.String(), Valid: true},
			ProdType:     sql.NullString{String: "clothes", Valid: true},
		},
		{
			PvzID:        pvzID2.String(),
			PvzCity:      "Tula",
			PvzCreatedAt: time.Now(),
			RecID:        sql.NullString{String: recID2.String(), Valid: true},
			RecStatus:    sql.NullString{String: "opened", Valid: true},
			RecDateTime:  sql.NullTime{Time: time.Now(), Valid: true},
			RecPvzID:     sql.NullString{String: pvzID2.String(), Valid: true},
			ProdID:       sql.NullString{String: prodID2.String(), Valid: true},
			ProdDateTime: sql.NullTime{Time: time.Now(), Valid: true},
			ProdRecID:    sql.NullString{String: recID2.String(), Valid: true},
			ProdType:     sql.NullString{String: "electronics", Valid: true},
		},
	}

	aggregator := newPvzAggregator()
	for _, row := range rows {
		err := aggregator.processRow(row)
		assert.NoError(t, err)
	}

	results := aggregator.Results()
	assert.Len(t, results, 2)

	assert.Equal(t, pvzID1, results[0].Pvz.Id)
	assert.Len(t, results[0].Receptions, 1)
	assert.Equal(t, recID1, results[0].Receptions[0].Reception.Id)
	assert.Len(t, results[0].Receptions[0].Products, 1)
	assert.Equal(t, prodID1, results[0].Receptions[0].Products[0].Id)

	assert.Equal(t, pvzID2, results[1].Pvz.Id)
	assert.Len(t, results[1].Receptions, 1)
	assert.Equal(t, recID2, results[1].Receptions[0].Reception.Id)
	assert.Len(t, results[1].Receptions[0].Products, 1)
	assert.Equal(t, prodID2, results[1].Receptions[0].Products[0].Id)
}

func Test_pvzAggregator_error_cases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		row    pvzRow
		expErr string
	}{
		{
			name: "invalid pvz uuid",
			row: pvzRow{
				PvzID: "invalid-uuid",
			},
			expErr: "invalid PVZ UUID",
		},
		{
			name: "invalid reception uuid",
			row: pvzRow{
				PvzID: uuid.New().String(),
				RecID: sql.NullString{String: "invalid-uuid", Valid: true},
			},
			expErr: "invalid reception UUID",
		},
		{
			name: "invalid product uuid",
			row: pvzRow{
				PvzID:    uuid.New().String(),
				RecID:    sql.NullString{String: uuid.New().String(), Valid: true},
				RecPvzID: sql.NullString{String: uuid.New().String(), Valid: true},
				ProdID:   sql.NullString{String: "invalid-uuid", Valid: true},
			},
			expErr: "invalid product UUID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			aggregator := newPvzAggregator()
			err := aggregator.processRow(tt.row)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expErr)
		})
	}
}

func Test_pvzAggregator_untested_parts(t *testing.T) {
	t.Parallel()

	t.Run("no reception id", func(t *testing.T) {
		t.Parallel()
		aggregator := newPvzAggregator()
		row := pvzRow{
			PvzID: uuid.New().String(),
			RecID: sql.NullString{Valid: false},
		}
		err := aggregator.processRow(row)
		assert.NoError(t, err)
	})

	t.Run("no product id", func(t *testing.T) {
		t.Parallel()
		aggregator := newPvzAggregator()
		row := pvzRow{
			PvzID:    uuid.New().String(),
			RecID:    sql.NullString{String: uuid.New().String(), Valid: true},
			RecPvzID: sql.NullString{String: uuid.New().String(), Valid: true},
			ProdID:   sql.NullString{Valid: false},
		}
		err := aggregator.processRow(row)
		assert.NoError(t, err)
	})

	t.Run("invalid reception pvz uuid", func(t *testing.T) {
		t.Parallel()
		aggregator := newPvzAggregator()
		row := pvzRow{
			PvzID:    uuid.New().String(),
			RecID:    sql.NullString{String: uuid.New().String(), Valid: true},
			RecPvzID: sql.NullString{String: "invalid-uuid", Valid: true},
		}
		err := aggregator.processRow(row)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid reception PVZ UUID")
	})

	t.Run("invalid product reception uuid", func(t *testing.T) {
		t.Parallel()
		aggregator := newPvzAggregator()
		row := pvzRow{
			PvzID:     uuid.New().String(),
			RecID:     sql.NullString{String: uuid.New().String(), Valid: true},
			RecPvzID:  sql.NullString{String: uuid.New().String(), Valid: true},
			ProdID:    sql.NullString{String: uuid.New().String(), Valid: true},
			ProdRecID: sql.NullString{String: "invalid-uuid", Valid: true},
		}
		err := aggregator.processRow(row)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid product reception UUID")
	})

	t.Run("existing pvz", func(t *testing.T) {
		t.Parallel()
		aggregator := newPvzAggregator()
		pvzID := uuid.New().String()
		row1 := pvzRow{PvzID: pvzID}
		row2 := pvzRow{PvzID: pvzID}
		err := aggregator.processRow(row1)
		assert.NoError(t, err)
		err = aggregator.processRow(row2)
		assert.NoError(t, err)
		assert.Len(t, aggregator.Results(), 1)
	})

	t.Run("existing reception", func(t *testing.T) {
		t.Parallel()
		aggregator := newPvzAggregator()
		pvzID := uuid.New().String()
		recID := uuid.New().String()
		row1 := pvzRow{
			PvzID:    pvzID,
			RecID:    sql.NullString{String: recID, Valid: true},
			RecPvzID: sql.NullString{String: pvzID, Valid: true},
		}
		row2 := pvzRow{
			PvzID:    pvzID,
			RecID:    sql.NullString{String: recID, Valid: true},
			RecPvzID: sql.NullString{String: pvzID, Valid: true},
		}
		err := aggregator.processRow(row1)
		assert.NoError(t, err)
		err = aggregator.processRow(row2)
		assert.NoError(t, err)
		assert.Len(t, aggregator.Results()[0].Receptions, 1)
	})
}
