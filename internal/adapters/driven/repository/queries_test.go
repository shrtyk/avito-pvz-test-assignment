package repository

import (
	"strings"
	"testing"
	"time"

	"github.com/shrtyk/avito-pvz-test-assignment/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

func Test_buildGetPvzDataQuery(t *testing.T) {
	t.Parallel()

	const baseQuery = `
		WITH pvzs_ids AS (
			SELECT DISTINCT pvz.id
			FROM pvzs AS pvz
	`
	const finalQuery = `
		SELECT
			pvz.id, pvz.city, pvz.created_at,
			r.id, r.status, r.created_at, r.pvz_id,
			p.id, p.added_at, p.reception_id, p.type
		FROM pvzs AS pvz
		LEFT JOIN receptions AS r
			ON pvz.id = r.pvz_id
		LEFT JOIN products AS p
			ON p.reception_id = r.id
		WHERE
			pvz.id IN (SELECT id FROM pvzs_ids)
		ORDER BY
			pvz.id, r.id, p.id
	`

	type args struct {
		params *domain.PvzsReadParams
	}
	tests := []struct {
		name             string
		args             args
		conditionalQuery string
		wantArgs         []any
	}{
		{
			name: "no params",
			args: args{
				params: &domain.PvzsReadParams{
					Limit: 10,
					Page:  1,
				},
			},
			conditionalQuery: `
			ORDER BY pvz.id
			LIMIT $1
			OFFSET $2
		)
	`,
			wantArgs: []any{10, 0},
		},
		{
			name: "with start date",
			args: args{
				params: &domain.PvzsReadParams{
					Limit:     10,
					Page:      1,
					StartDate: &time.Time{},
				},
			},
			conditionalQuery: ` INNER JOIN receptions AS r ON pvz.id = r.pvz_id WHERE r.created_at >= $1
			ORDER BY pvz.id
			LIMIT $2
			OFFSET $3
		)
	`,
			wantArgs: []any{&time.Time{}, 10, 0},
		},
		{
			name: "with end date",
			args: args{
				params: &domain.PvzsReadParams{
					Limit:   10,
					Page:    1,
					EndDate: &time.Time{},
				},
			},
			conditionalQuery: ` INNER JOIN receptions AS r ON pvz.id = r.pvz_id WHERE r.created_at <= $1
			ORDER BY pvz.id
			LIMIT $2
			OFFSET $3
		)
	`,
			wantArgs: []any{&time.Time{}, 10, 0},
		},
		{
			name: "with both dates",
			args: args{
				params: &domain.PvzsReadParams{
					Limit:     10,
					Page:      1,
					StartDate: &time.Time{},
					EndDate:   &time.Time{},
				},
			},
			conditionalQuery: ` INNER JOIN receptions AS r ON pvz.id = r.pvz_id WHERE r.created_at >= $1 AND r.created_at <= $2
			ORDER BY pvz.id
			LIMIT $3
			OFFSET $4
		)
	`,
			wantArgs: []any{&time.Time{}, &time.Time{}, 10, 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var wantQueryBuilder strings.Builder
			wantQueryBuilder.WriteString(baseQuery)
			wantQueryBuilder.WriteString(tt.conditionalQuery)
			wantQueryBuilder.WriteString(finalQuery)

			gotQuery, gotArgs := buildGetPvzDataQuery(tt.args.params)
			assert.Equal(t, wantQueryBuilder.String(), string(gotQuery))
			assert.Equal(t, tt.wantArgs, gotArgs)
		})
	}
}
