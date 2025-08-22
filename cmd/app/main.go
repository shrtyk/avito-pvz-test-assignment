package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/shrtyk/avito-pvz-test-assignment/internal/core/service"
	pwdservice "github.com/shrtyk/avito-pvz-test-assignment/internal/infrastructure/pwd_service"
	"github.com/shrtyk/avito-pvz-test-assignment/internal/infrastructure/repository"
	ts "github.com/shrtyk/avito-pvz-test-assignment/internal/infrastructure/tservice"
	"github.com/shrtyk/avito-pvz-test-assignment/pkg/config"
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
	appService := service.NewAppService(cfg.AppCfg.Timeout, repo, pwdService, tokenService)

	app := NewApplication()
	app.Init(
		WithConfig(cfg),
		WithLogger(log),
		WithTokenService(tokenService),
		WithRepo(repo),
		WithService(appService),
	)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	app.Serve(ctx)
}
