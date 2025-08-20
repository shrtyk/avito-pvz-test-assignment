package auth

import (
	"github.com/golang-jwt/jwt/v5"
)

type AccessTokenData struct {
	UserID int64
	Role   UserRole
}

type AccessTokenClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func (atc AccessTokenClaims) UserID() string {
	return atc.Subject
}
