package auth

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestUserID(t *testing.T) {
	at := AccessTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "1",
		},
	}
	assert.Equal(t, "1", at.UserID())
}
