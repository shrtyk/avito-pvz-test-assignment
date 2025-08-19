package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func unsetEnvForTest(t *testing.T, keys ...string) {
	t.Helper()
	originalValues := make(map[string]string)
	for _, key := range keys {
		if val, ok := os.LookupEnv(key); ok {
			originalValues[key] = val
		}
		os.Unsetenv(key)
	}
	t.Cleanup(func() {
		for key, val := range originalValues {
			os.Setenv(key, val)
		}
	})
}

func createTempConfigFile(t *testing.T, content string) string {
	t.Helper()
	file, err := os.CreateTemp(t.TempDir(), "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}
	_, err = file.WriteString(content)
	if err != nil {
		t.Fatalf("Failed to write to temp config file: %v", err)
	}
	require.NoError(t, file.Close())
	return file.Name()
}

func TestMustInitConfig(t *testing.T) {
	t.Run("should load config from yaml file", func(t *testing.T) {
		unsetEnvForTest(
			t,
			"PVZ_ENV", "PVZ_TIMEOUT", "PG_HOST",
			"PG_PORT", "PG_USER", "HTTP_SERVER_PORT",
		)

		configContent := `
pvz:
  env: dev
  timeout: 10s
postgres:
  host: file_host
  port: 5432
  user: file_user
`
		filePath := createTempConfigFile(t, configContent)
		t.Setenv("CONFIG_PATH", filePath)

		cfg := MustInitConfig()

		assert.Equal(t, "dev", cfg.AppCfg.Env)
		assert.Equal(t, 10*time.Second, cfg.AppCfg.Timeout)
		assert.Equal(t, "file_host", cfg.PostgresCfg.Host)
		assert.Equal(t, "file_user", cfg.PostgresCfg.User)

		assert.Equal(t, "8080", cfg.HttpServerCfg.Port)
	})

	t.Run("should load config from environment variables", func(t *testing.T) {
		unsetEnvForTest(t, "CONFIG_PATH")

		t.Setenv("PVZ_ENV", "dev")
		t.Setenv("PG_HOST", "env_host")
		t.Setenv("PG_PORT", "5433")
		t.Setenv("HTTP_SERVER_PORT", "8080")

		cfg := MustInitConfig()

		assert.Equal(t, "dev", cfg.AppCfg.Env)
		assert.Equal(t, "env_host", cfg.PostgresCfg.Host)
		assert.Equal(t, "5433", cfg.PostgresCfg.Port)
		assert.Equal(t, "8080", cfg.HttpServerCfg.Port)
	})

	t.Run("should allow environment variables to override file config", func(t *testing.T) {
		configContent := `
pvz:
  env: file_env
postgres:
  host: file_host
`
		filePath := createTempConfigFile(t, configContent)
		t.Setenv("CONFIG_PATH", filePath)

		t.Setenv("PVZ_ENV", "env_env_override")
		t.Setenv("PG_HOST", "env_host_override")

		cfg := MustInitConfig()

		assert.Equal(t, "env_env_override", cfg.AppCfg.Env)
		assert.Equal(t, "env_host_override", cfg.PostgresCfg.Host)
	})

	t.Run("should use default values when no file or env var is provided", func(t *testing.T) {
		unsetEnvForTest(
			t,
			"CONFIG_PATH", "PVZ_ENV", "PVZ_TIMEOUT", "PG_HOST", "PG_USER",
		)

		cfg := MustInitConfig()

		assert.Equal(t, "prod", cfg.AppCfg.Env)
		assert.Equal(t, "5s", cfg.AppCfg.Timeout.String())
		assert.Equal(t, "postgres", cfg.PostgresCfg.Host)
		assert.Equal(t, "user", cfg.PostgresCfg.User)
	})

	t.Run("should panic on malformed config file", func(t *testing.T) {
		configContent := `wrong yaml`
		filePath := createTempConfigFile(t, configContent)
		t.Setenv("CONFIG_PATH", filePath)

		assert.Panics(t, func() {
			MustInitConfig()
		})
	})
}
