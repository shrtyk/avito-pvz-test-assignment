package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/shrtyk/avito-pvz-test-assignment/internal/config"
	"github.com/shrtyk/avito-pvz-test-assignment/internal/core/service"
	"github.com/shrtyk/avito-pvz-test-assignment/internal/infrastructure/prometheus"
	pwdservice "github.com/shrtyk/avito-pvz-test-assignment/internal/infrastructure/pwd_service"
	"github.com/shrtyk/avito-pvz-test-assignment/internal/infrastructure/repository"
	ts "github.com/shrtyk/avito-pvz-test-assignment/internal/infrastructure/tservice"
	"github.com/shrtyk/avito-pvz-test-assignment/pkg/dbs/postgres"
	"github.com/shrtyk/avito-pvz-test-assignment/pkg/logger"
)

func main() {
	cfg := config.MustInitConfig()
	log := logger.MustCreateNewLogger(cfg.AppCfg.Env)
	db := postgres.MustCreateConnectionPool(&cfg.PostgresCfg)
	repo := repository.NewRepo(db)
	tokenService := ts.MustCreateTokenService(&cfg.AuthTokenCfg)
	pwdService := pwdservice.NewPasswordService()
	metrics := prometheus.NewPrometheusCollector()
	appService := service.NewAppService(
		cfg.AppCfg.Timeout,
		repo,
		pwdService,
		tokenService,
		metrics,
	)

	app := NewApplication()
	app.Init(
		WithConfig(cfg),
		WithLogger(log),
		WithTokenService(tokenService),
		WithRepo(repo),
		WithService(appService),
		WithMetrics(metrics),
	)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	app.Serve(ctx)
}
