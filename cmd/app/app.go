package main

import (
	"log/slog"

	"github.com/shrtyk/avito-backend-spring-2025/internal/core/ports"
	"github.com/shrtyk/avito-backend-spring-2025/pkg/config"
)

type Application struct {
	Cfg          *config.Config
	Logger       *slog.Logger
	Repo         ports.Repository
	TokenService ports.TokensService
}

type option func(*Application)

func NewApplication() *Application {
	return &Application{}
}

func (app *Application) Init(opts ...option) {
	for _, opt := range opts {
		opt(app)
	}
}

func WithConfig(cfg *config.Config) option {
	return func(app *Application) {
		app.Cfg = cfg
	}
}

func WithLogger(log *slog.Logger) option {
	return func(app *Application) {
		app.Logger = log
	}
}

func WithRepo(repo ports.Repository) option {
	return func(app *Application) {
		app.Repo = repo
	}
}

func WithTokenService(tService ports.TokensService) option {
	return func(app *Application) {
		app.TokenService = tService
	}
}
