package tservice

import (
	"context"

	dAuth "github.com/shrtyk/avito-backend-spring-2025/internal/core/domain/auth"
	pAuth "github.com/shrtyk/avito-backend-spring-2025/internal/core/ports/auth"
	xerr "github.com/shrtyk/avito-backend-spring-2025/pkg/xerrors"
)

type ctxKey string

const claimsKey = ctxKey("JWTClaims")

func ClaimsToCtx(ctx context.Context, claims *dAuth.AccessTokenClaims) context.Context {

	return context.WithValue(ctx, claimsKey, claims)
}

func ClaimsFromCtx(ctx context.Context) (*dAuth.AccessTokenClaims, error) {
	op := "tokens.ClaimsFromCtx"

	claims, ok := ctx.Value(claimsKey).(*dAuth.AccessTokenClaims)
	if !ok {
		return nil, xerr.NewErr(op, pAuth.JwtClaimsFromCtx)
	}
	return claims, nil
}
