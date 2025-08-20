//go:generate mockery
package auth

import "github.com/shrtyk/avito-pvz-test-assignment/internal/core/domain/auth"

//go:generate mockery
type TokensService interface {
	GenerateAccessToken(tokenData auth.AccessTokenData) (string, error)
	GetTokenClaims(token string) (*auth.AccessTokenClaims, error)
}
