//go:build integration
// +build integration

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/shrtyk/pvz-service/internal/api/http/dto"
	"github.com/shrtyk/pvz-service/internal/config"
	"github.com/shrtyk/pvz-service/internal/core/service"
	"github.com/shrtyk/pvz-service/internal/infrastructure/prometheus"
	pwdservice "github.com/shrtyk/pvz-service/internal/infrastructure/pwd_service"
	"github.com/shrtyk/pvz-service/internal/infrastructure/repository"
	ts "github.com/shrtyk/pvz-service/internal/infrastructure/tservice"
	pkgpg "github.com/shrtyk/pvz-service/pkg/dbs/postgres"
	"github.com/shrtyk/pvz-service/pkg/logger"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	roleModerator = "moderator"
	roleEmployee  = "employee"

	contentTypeJSON = "application/json"
)

var testHTTPClient = &http.Client{Timeout: 10 * time.Second}

type envVar struct {
	key, value string
}

type testAppConfig struct {
	dbHost         string
	dbPort         string
	dbUser         string
	dbPassword     string
	dbName         string
	dbSSLMode      string
	publicRsaPath  string
	privateRsaPath string
}

func TestIntegration(t *testing.T) {
	pgContainer, cancel := setupDB(t.Context(), t)
	defer cancel()

	host, err := pgContainer.Host(t.Context())
	require.NoError(t, err)
	port, err := pgContainer.MappedPort(t.Context(), "5432")
	require.NoError(t, err)

	dsn := fmt.Sprintf("user=user password=password host=%s port=%s dbname=pvz-db sslmode=disable", host, port.Port())

	migrationsPath, err := filepath.Abs("../../migrations")
	require.NoError(t, err)

	err = migrationsUp(t.Context(), migrationsPath, dsn)
	require.NoError(t, err)

	baseURL := startTestApp(t, &testAppConfig{
		dbHost:         host,
		dbPort:         port.Port(),
		dbUser:         "user",
		dbPassword:     "password",
		dbName:         "pvz-db",
		dbSSLMode:      "disable",
		publicRsaPath:  "../../keys/rsa/public_key.pem",
		privateRsaPath: "../../keys/rsa/private_key.pem",
	})

	moderatorToken := getDummyToken(t, baseURL, roleModerator)
	employeeToken := getDummyToken(t, baseURL, roleEmployee)

	var pvzID uuid.UUID
	t.Run("Create PVZ", func(t *testing.T) {
		pvz := createPVZ(t, baseURL, moderatorToken)
		require.NotNil(t, pvz.Id)
		pvzID = *pvz.Id
	})

	require.NotNil(t, pvzID, "PVZ ID should not be nil after creation")

	t.Run("Create Reception", func(t *testing.T) {
		reception := createReception(t, baseURL, employeeToken, pvzID)
		require.NotNil(t, reception.Id)
	})

	t.Run("Add Products", func(t *testing.T) {
		for range 50 {
			addProduct(t, baseURL, employeeToken, pvzID)
		}
	})

	t.Run("Close Reception", func(t *testing.T) {
		closeReception(t, baseURL, employeeToken, pvzID)
	})

	t.Run("Verify Reception Closed", func(t *testing.T) {
		reqBody, err := json.Marshal(dto.PostProductsJSONBody{PvzId: pvzID, Type: "одежда"})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", fmt.Sprintf("%s/products", baseURL), bytes.NewBuffer(reqBody))
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+employeeToken)
		req.Header.Set("Content-Type", contentTypeJSON)

		resp, err := testHTTPClient.Do(req)
		require.NoError(t, err)
		defer require.NoError(t, resp.Body.Close())

		require.Equal(t, http.StatusBadRequest, resp.StatusCode, "should not be able to add products to a closed reception")
	})
}

func setEnvForTest(t *testing.T, vars ...envVar) {
	t.Helper()
	originalValues := make(map[string]string)
	for _, v := range vars {
		if val, ok := os.LookupEnv(v.key); ok {
			originalValues[v.key] = val
		}
		require.NoError(t, os.Unsetenv(v.key))
		require.NoError(t, os.Setenv(v.key, v.value))
	}

	t.Cleanup(func() {
		for key, val := range originalValues {
			require.NoError(t, os.Setenv(key, val))
		}
	})
}

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = l.Close()
	}()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func setupDB(ctx context.Context, t testing.TB) (ctr *postgres.PostgresContainer, cancel func()) {
	t.Helper()

	ctr, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("pvz-db"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
		postgres.BasicWaitStrategies(),
	)
	require.NoError(t, err)

	cancel = func() { _ = ctr.Terminate(ctx) }
	return
}

func migrationsUp(ctx context.Context, paths, dsn string) error {
	db, err := goose.OpenDBWithDriver("pgx", dsn)
	if err != nil {
		return err
	}
	defer func() {
		_ = db.Close()
	}()

	if err := goose.UpContext(ctx, db, paths); err != nil {
		return err
	}

	return nil
}

func startTestApp(t *testing.T, appCfg *testAppConfig) string {
	t.Helper()

	httpPort, err := getFreePort()
	require.NoError(t, err)
	httpPortStr := strconv.Itoa(httpPort)

	setEnvForTest(
		t,
		envVar{key: "HTTP_SERVER_PORT", value: httpPortStr},
		envVar{key: "PG_HOST", value: appCfg.dbHost},
		envVar{key: "PG_USER", value: appCfg.dbUser},
		envVar{key: "PG_PASSWORD", value: appCfg.dbPassword},
		envVar{key: "PG_DBNAME", value: appCfg.dbName},
		envVar{key: "PG_PORT", value: appCfg.dbPort},
		envVar{key: "PG_SSL_MODE", value: appCfg.dbSSLMode},
		envVar{key: "PUBLIC_RSA_PATH", value: appCfg.publicRsaPath},
		envVar{key: "PRIVATE_RSA_PATH", value: appCfg.privateRsaPath},
	)

	cfg := config.MustInitConfig()
	log, _ := logger.NewTestLogger()
	tService := ts.MustCreateTokenService(&cfg.AuthTokenCfg)
	db := pkgpg.MustCreateConnectionPool(&cfg.PostgresCfg)
	repo := repository.NewRepo(db)
	metrics := prometheus.NewPrometheusCollector()
	pwdService := pwdservice.NewPasswordService()
	appService := service.NewAppService(cfg.AppCfg.Timeout, repo, pwdService, tService, metrics)

	app := NewApplication()
	app.Init(
		WithConfig(cfg),
		WithLogger(log),
		WithTokenService(tService),
		WithRepo(repo),
		WithService(appService),
		WithMetrics(metrics),
	)

	go func() {
		app.Serve(t.Context())
	}()

	baseURL := fmt.Sprintf("http://localhost:%s", httpPortStr)

	require.Eventually(t, func() bool {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, fmt.Sprintf("%s/healthz", baseURL), nil)
		if err != nil {
			return false
		}
		resp, err := testHTTPClient.Do(req)
		if err != nil {
			t.Logf("healthz check failed: %v", err)
			return false
		}
		defer require.NoError(t, resp.Body.Close())
		if resp.StatusCode != http.StatusOK {
			t.Logf("healthz check returned status %d", resp.StatusCode)
			return false
		}
		return true
	}, 15*time.Second, 200*time.Millisecond)

	return baseURL
}

func getDummyToken(t *testing.T, baseURL, role string) string {
	t.Helper()
	reqBody, err := json.Marshal(map[string]string{"role": role})
	require.NoError(t, err)

	resp, err := testHTTPClient.Post(fmt.Sprintf("%s/dummyLogin", baseURL), contentTypeJSON, bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	defer func() {
		_ = resp.Body.Close()
	}()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var token dto.Token
	err = json.NewDecoder(resp.Body).Decode(&token)
	require.NoError(t, err)

	return token.Jwt
}

func createPVZ(t *testing.T, baseURL, token string) *dto.PVZ {
	t.Helper()
	reqBody, err := json.Marshal(dto.PVZ{City: "Москва"})
	require.NoError(t, err)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/pvz", baseURL), bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", contentTypeJSON)

	resp, err := testHTTPClient.Do(req)
	require.NoError(t, err)
	defer func() {
		_ = resp.Body.Close()
	}()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var pvz dto.PVZ
	err = json.NewDecoder(resp.Body).Decode(&pvz)
	require.NoError(t, err)

	return &pvz
}

func createReception(t *testing.T, baseURL, token string, pvzID uuid.UUID) *dto.Reception {
	t.Helper()
	reqBody, err := json.Marshal(dto.PostReceptionsJSONBody{PvzId: pvzID})
	require.NoError(t, err)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/receptions", baseURL), bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", contentTypeJSON)

	resp, err := testHTTPClient.Do(req)
	require.NoError(t, err)
	defer func() {
		_ = resp.Body.Close()
	}()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var reception dto.Reception
	err = json.NewDecoder(resp.Body).Decode(&reception)
	require.NoError(t, err)

	return &reception
}

func addProduct(t *testing.T, baseURL, token string, pvzID uuid.UUID) {
	t.Helper()
	reqBody, err := json.Marshal(dto.PostProductsJSONBody{PvzId: pvzID, Type: "одежда"})
	require.NoError(t, err)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/products", baseURL), bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", contentTypeJSON)

	resp, err := testHTTPClient.Do(req)
	require.NoError(t, err)
	defer func() {
		_ = resp.Body.Close()
	}()

	require.Equal(t, http.StatusCreated, resp.StatusCode)
}

func closeReception(t *testing.T, baseURL, token string, pvzID uuid.UUID) {
	t.Helper()
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s/close_last_reception", baseURL, pvzID), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := testHTTPClient.Do(req)
	require.NoError(t, err)
	defer func() {
		_ = resp.Body.Close()
	}()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}
