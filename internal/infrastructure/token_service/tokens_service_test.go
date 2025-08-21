package tservice

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/shrtyk/avito-pvz-test-assignment/internal/core/domain/auth"
	"github.com/shrtyk/avito-pvz-test-assignment/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestTokensService(t *testing.T, accessTokenLifetime time.Duration) *tokensService {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	return &tokensService{
		publicKey:           &privateKey.PublicKey,
		privateKey:          privateKey,
		accessTokenLifetime: accessTokenLifetime,
	}
}

func TestTokenService_Context(t *testing.T) {
	expected := &auth.AccessTokenClaims{
		Role: string(auth.UserRoleModerator),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "0",
			ExpiresAt: &jwt.NumericDate{Time: time.Now()},
			IssuedAt:  &jwt.NumericDate{Time: time.Now()},
			ID:        "0",
		},
	}

	ctx := ClaimsToCtx(context.Background(), expected)
	got, err := ClaimsFromCtx(ctx)

	require.NoError(t, err)
	assert.Equal(t, expected, got)

	wrongCtx := context.WithValue(context.Background(), claimsKey, "wrong key")
	_, err = ClaimsFromCtx(wrongCtx)
	require.Error(t, err)
}

func TestTokensService(t *testing.T) {
	t.Parallel()

	t.Run("success case", func(t *testing.T) {
		t.Parallel()

		tokensService := newTestTokensService(t, time.Hour)

		tokenData := auth.AccessTokenData{
			UserID: 1,
			Role:   "admin",
		}

		accessToken, err := tokensService.GenerateAccessToken(tokenData)
		require.NoError(t, err)

		claims, err := tokensService.GetTokenClaims(accessToken)
		require.NoError(t, err)

		assert.Equal(t, "1", claims.Subject)
		assert.Equal(t, "admin", claims.Role)
	})

	t.Run("expired token", func(t *testing.T) {
		t.Parallel()

		tokensService := newTestTokensService(t, -time.Hour)

		tokenData := auth.AccessTokenData{
			UserID: 1,
			Role:   "admin",
		}

		accessToken, err := tokensService.GenerateAccessToken(tokenData)
		require.NoError(t, err)

		_, err = tokensService.GetTokenClaims(accessToken)
		assert.Error(t, err)
	})

	t.Run("invalid token", func(t *testing.T) {
		t.Parallel()

		tokensService := newTestTokensService(t, time.Hour)

		_, err := tokensService.GetTokenClaims("invalid-token")
		assert.Error(t, err)
	})
}

func TestMustCreateTokenService(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
		privateKeyPEM := pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privateKeyBytes,
		})

		publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
		require.NoError(t, err)
		publicKeyPEM := pem.EncodeToMemory(&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: publicKeyBytes,
		})

		privateKeyPath := filepath.Join(tempDir, "private.pem")
		publicKeyPath := filepath.Join(tempDir, "public.pem")

		require.NoError(t, os.WriteFile(privateKeyPath, privateKeyPEM, 0600))
		require.NoError(t, os.WriteFile(publicKeyPath, publicKeyPEM, 0600))

		cfg := &config.AuthTokensCfg{
			PublicRSAPath:  publicKeyPath,
			PrivateRSAPath: privateKeyPath,
		}

		assert.NotPanics(t, func() {
			MustCreateTokenService(cfg)
		})
	})

	t.Run("no public key", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		privateKeyPath := filepath.Join(tempDir, "private.pem")

		cfg := &config.AuthTokensCfg{
			PublicRSAPath:  "fake-file",
			PrivateRSAPath: privateKeyPath,
		}
		assert.Panics(t, func() {
			MustCreateTokenService(cfg)
		})
	})

	t.Run("no private key", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		publicKeyPath := filepath.Join(tempDir, "public.pem")

		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
		require.NoError(t, err)
		publicKeyPEM := pem.EncodeToMemory(&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: publicKeyBytes,
		})

		require.NoError(t, os.WriteFile(publicKeyPath, publicKeyPEM, 0600))

		cfg := &config.AuthTokensCfg{
			PublicRSAPath:  publicKeyPath,
			PrivateRSAPath: "fake-file",
		}
		assert.Panics(t, func() {
			MustCreateTokenService(cfg)
		})
	})

	t.Run("invalid public key", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		privateKeyPath := filepath.Join(tempDir, "private.pem")
		publicKeyPath := filepath.Join(tempDir, "public.pem")

		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
		privateKeyPEM := pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privateKeyBytes,
		})

		require.NoError(t, os.WriteFile(privateKeyPath, privateKeyPEM, 0600))
		require.NoError(t, os.WriteFile(publicKeyPath, []byte("invalid key"), 0600))

		cfg := &config.AuthTokensCfg{
			PublicRSAPath:  publicKeyPath,
			PrivateRSAPath: privateKeyPath,
		}

		assert.Panics(t, func() {
			MustCreateTokenService(cfg)
		})
	})

	t.Run("invalid private key", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		privateKeyPath := filepath.Join(tempDir, "private.pem")
		publicKeyPath := filepath.Join(tempDir, "public.pem")

		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
		require.NoError(t, err)
		publicKeyPEM := pem.EncodeToMemory(&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: publicKeyBytes,
		})

		require.NoError(t, os.WriteFile(publicKeyPath, publicKeyPEM, 0600))
		require.NoError(t, os.WriteFile(privateKeyPath, []byte("invalid key"), 0600))

		cfg := &config.AuthTokensCfg{
			PublicRSAPath:  publicKeyPath,
			PrivateRSAPath: privateKeyPath,
		}

		assert.Panics(t, func() {
			MustCreateTokenService(cfg)
		})
	})
}
