package auth

import (
	"context"
)

type ctxKey string

const claimsKey = ctxKey("JWTClaims")

func ClaimsToCtx(ctx context.Context, claims *AccessTokenClaims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

func ClaimsFromCtx(ctx context.Context) (*AccessTokenClaims, error) {
	claims, ok := ctx.Value(claimsKey).(*AccessTokenClaims)
	if !ok {
		return nil, ErrClaimsFormCtx
	}
	return claims, nil
}
