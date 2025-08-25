package repository

import (
	"fmt"
	"testing"
	"time"

	"github.com/shrtyk/pvz-service/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

func Test_buildGetPvzDataQuery(t *testing.T) {
	t.Parallel()

	const mainQueryTpl = `
WITH pvzs_ids AS (%s)
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
		name      string
		args      args
		wantQuery string
		wantArgs  []any
	}{
		{
			name: "no params",
			args: args{
				params: &domain.PvzsReadParams{
					Limit: 10,
					Page:  1,
				},
			},
			wantQuery: fmt.Sprintf(mainQueryTpl, "SELECT DISTINCT pvz.id FROM pvzs AS pvz ORDER BY pvz.id LIMIT 10 OFFSET 0"),
			wantArgs:  nil,
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
			wantQuery: fmt.Sprintf(mainQueryTpl,
				"SELECT DISTINCT pvz.id FROM pvzs AS pvz "+
					"JOIN receptions AS r ON pvz.id = r.pvz_id "+
					"WHERE r.created_at >= $1 ORDER BY pvz.id LIMIT 10 OFFSET 0",
			),
			wantArgs: []any{&time.Time{}},
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
			wantQuery: fmt.Sprintf(mainQueryTpl,
				"SELECT DISTINCT pvz.id FROM pvzs AS pvz "+
					"JOIN receptions AS r ON pvz.id = r.pvz_id "+
					"WHERE r.created_at <= $1 ORDER BY pvz.id LIMIT 10 OFFSET 0",
			),
			wantArgs: []any{&time.Time{}},
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
			wantQuery: fmt.Sprintf(mainQueryTpl,
				"SELECT DISTINCT pvz.id FROM pvzs AS pvz "+
					"JOIN receptions AS r ON pvz.id = r.pvz_id "+
					"WHERE r.created_at >= $1 AND r.created_at <= $2 ORDER BY pvz.id LIMIT 10 OFFSET 0",
			),
			wantArgs: []any{&time.Time{}, &time.Time{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotQuery, gotArgs, err := buildGetPvzDataQuery(tt.args.params)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantQuery, gotQuery)
			assert.Equal(t, tt.wantArgs, gotArgs)
		})
	}
}
