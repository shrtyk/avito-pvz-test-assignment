package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/shrtyk/avito-pvz-test-assignment/internal/adapters/driven/repository"
	tservice "github.com/shrtyk/avito-pvz-test-assignment/internal/adapters/driven/token_service"
	"github.com/shrtyk/avito-pvz-test-assignment/internal/core/service"
	"github.com/shrtyk/avito-pvz-test-assignment/pkg/config"
	"github.com/shrtyk/avito-pvz-test-assignment/pkg/dbs/postgres"
	"github.com/shrtyk/avito-pvz-test-assignment/pkg/logger"
)

func main() {
	cfg := config.MustInitConfig()
	log := logger.MustCreateNewLogger(cfg.AppCfg.Env)
	tService := tservice.MustCreateTokenService(&cfg.AuthTokenCfg)
	db := postgres.MustCreateConnectionPool(&cfg.PostgresCfg)
	repo := repository.NewRepo(db)
	appService := service.NewAppService(cfg.AppCfg.Timeout, repo)

	app := NewApplication()
	app.Init(
		WithConfig(cfg),
		WithLogger(log),
		WithTokenService(tService),
		WithRepo(repo),
		WithService(appService),
	)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	app.Serve(ctx)
}
