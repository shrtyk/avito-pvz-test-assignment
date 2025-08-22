package auth

import (
	"time"

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

type RefreshTokenData struct {
	Token     string
	UserID    string
	IP        string
	UserAgent string
	CreatedAt time.Time
	ExpireAt  time.Time
}
