package main

import (
	"log/slog"

	"github.com/shrtyk/avito-pvz-test-assignment/internal/config"
	pAuth "github.com/shrtyk/avito-pvz-test-assignment/internal/core/ports/auth"
	pRepo "github.com/shrtyk/avito-pvz-test-assignment/internal/core/ports/repository"
	pService "github.com/shrtyk/avito-pvz-test-assignment/internal/core/ports/service"
)

type Application struct {
	Cfg          *config.Config
	Logger       *slog.Logger
	Repo         pRepo.Repository
	TokenService pAuth.TokenService
	AppService   pService.Service
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

func WithRepo(repo pRepo.Repository) option {
	return func(app *Application) {
		app.Repo = repo
	}
}

func WithTokenService(tokenService pAuth.TokenService) option {
	return func(app *Application) {
		app.TokenService = tokenService
	}
}

func WithService(s pService.Service) option {
	return func(app *Application) {
		app.AppService = s
	}
}
