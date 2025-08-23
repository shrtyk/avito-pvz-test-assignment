package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/shrtyk/pvz-service/internal/config"
	"github.com/shrtyk/pvz-service/internal/core/service"
	"github.com/shrtyk/pvz-service/internal/dbs/postgres"
	"github.com/shrtyk/pvz-service/internal/infrastructure/prometheus"
	pwdservice "github.com/shrtyk/pvz-service/internal/infrastructure/pwd_service"
	"github.com/shrtyk/pvz-service/internal/infrastructure/repository"
	ts "github.com/shrtyk/pvz-service/internal/infrastructure/tservice"
	"github.com/shrtyk/pvz-service/pkg/logger"
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
