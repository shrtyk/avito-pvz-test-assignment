package repository

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/shrtyk/pvz-service/internal/core/domain"
)

type query string

const (
	createPvzQuery query = `
		INSERT INTO pvzs
			(city)
		VALUES
			($1)
		RETURNING
			id, created_at
	`

	createReceptionQuery query = `
		INSERT INTO receptions
			(pvz_id)
		VALUES
			($1)
		RETURNING
			id, created_at, status
	`

	createProductQuery query = `
		INSERT INTO products
	 		(reception_id, type)
		VALUES(
			(SELECT id FROM receptions WHERE pvz_id = $1 AND status = $2), $3
		)
		RETURNING
			id, added_at, reception_id, type
	`

	deleteLastProductQuery query = `
		DELETE FROM
			products
		WHERE
			id = (
			SELECT id FROM products
			WHERE reception_id = (
				SELECT id FROM receptions WHERE pvz_id = $1 AND status = $2
			)
			ORDER BY added_at DESC
			LIMIT 1
		)
	`

	closeReceptionPvzQuery query = `
		UPDATE
			receptions
		SET
			status = $1
		WHERE
			id = (SELECT id FROM receptions WHERE pvz_id = $2 AND status = $3)
	`

	getAllPvzsQuery query = `
		SELECT
			id, created_at, city
		FROM
			pvzs
	`

	insertUserQuery query = `
		INSERT INTO users
	 		(email, role, password_hash)
		VALUES
			($1, $2, $3)
		RETURNING
			id, email, role
	`

	getUserByEmailQuery query = `
		SELECT
			id, password_hash, role, created_at
		FROM
			users
		WHERE
			email = $1
	`

	insertRefreshTokenQuery query = `
		INSERT INTO refresh_tokens
			(token_hash, fingerprint, user_id, user_agent, ip_address, created_at, expires_at)
		VALUES
			($1, $2, $3, $4, $5, $6, $7)
	`

	getRefreshTokenByHashQuery query = `
		SELECT
			u.role, rt.fingerprint, rt.user_id, rt.user_agent,
			rt.ip_address, rt.created_at, rt.expires_at, rt.revoked
		FROM
			refresh_tokens AS rt
		JOIN users AS u
			ON u.id = rt.user_id
		WHERE
			token_hash = $1
	`

	revokeOldRefreshTokenQuery query = `
		UPDATE
			refresh_tokens
		SET
			revoked = true
		WHERE
			token_hash = $1
	`
)

func buildGetPvzDataQuery(params *domain.PvzsReadParams) (string, []any, error) {
	sb := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	sub := sb.
		Select("DISTINCT pvz.id").
		From("pvzs AS pvz")

	if params.StartDate != nil || params.EndDate != nil {
		sub = sub.Join("receptions AS r ON pvz.id = r.pvz_id")
		if params.StartDate != nil {
			sub = sub.Where("r.created_at >= ?", params.StartDate)
		}
		if params.EndDate != nil {
			sub = sub.Where("r.created_at <= ?", params.EndDate)
		}
	}

	offset := (params.Page - 1) * params.Limit
	sub = sub.
		OrderBy("pvz.id").
		Limit(uint64(params.Limit)).
		Offset(uint64(offset))

	subSQL, subArgs, err := sub.ToSql()
	if err != nil {
		return "", nil, err
	}

	mainSQL := fmt.Sprintf(`
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
`, subSQL)

	return mainSQL, subArgs, nil
}
