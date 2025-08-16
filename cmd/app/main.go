package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/shrtyk/avito-backend-spring-2025/internal/adapters/driven/repository"
	"github.com/shrtyk/avito-backend-spring-2025/internal/adapters/driven/tokens"
	"github.com/shrtyk/avito-backend-spring-2025/internal/core/service"
	"github.com/shrtyk/avito-backend-spring-2025/pkg/config"
	"github.com/shrtyk/avito-backend-spring-2025/pkg/dbs/postgres"
	"github.com/shrtyk/avito-backend-spring-2025/pkg/logger"
)

func main() {
	cfg := config.MustInitConfig()
	log := logger.MustCreateNewLogger(cfg.AppCfg.Env)
	tService := tokens.MustCreateTokenService(&cfg.AuthTokenCfg)
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
