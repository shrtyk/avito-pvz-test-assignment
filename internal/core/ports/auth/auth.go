//go:generate mockery
package auth

import "github.com/shrtyk/pvz-service/internal/core/domain/auth"

//go:generate mockery
type TokenService interface {
	GenerateAccessToken(tokenData auth.AccessTokenData) (string, error)
	GetTokenClaims(token string) (*auth.AccessTokenClaims, error)
	GenerateRefreshToken(userID, ua, ip string) *auth.RefreshToken
	Fingerprint(rToken *auth.RefreshToken) string
	Hash(token string) []byte
}
