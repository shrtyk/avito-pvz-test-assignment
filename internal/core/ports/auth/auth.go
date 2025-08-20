//go:generate mockery
package auth

import "github.com/shrtyk/avito-backend-spring-2025/internal/core/domain/auth"

//go:generate mockery
type TokensService interface {
	GenerateAccessToken(tokenData auth.AccessTokenData) (string, error)
	GetTokenClaims(token string) (*auth.AccessTokenClaims, error)
}
