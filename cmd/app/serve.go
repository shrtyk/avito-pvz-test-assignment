package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	appHttp "github.com/shrtyk/avito-pvz-test-assignment/internal/adapters/driving/http"
	"github.com/shrtyk/avito-pvz-test-assignment/pkg/logger"
)

func (app Application) Serve(ctx context.Context) {
	s := http.Server{
		Addr:         ":" + app.Cfg.HttpServerCfg.Port,
		Handler:      appHttp.NewRouter(app.AppService, app.TokenService, app.Logger),
		IdleTimeout:  app.Cfg.HttpServerCfg.IdleTimeout,
		WriteTimeout: app.Cfg.HttpServerCfg.WriteTimeout,
		ReadTimeout:  app.Cfg.HttpServerCfg.ReadTimeout,
		ErrorLog:     slog.NewLogLogger(app.Logger.Handler(), slog.LevelError),
	}

	eChan := make(chan error, 1)
	go func() {
		<-ctx.Done()

		tCtx, tCancel := context.WithTimeout(context.Background(), app.Cfg.AppCfg.ShutdownTimeout)
		defer tCancel()

		eChan <- s.Shutdown(tCtx)
	}()

	app.Logger.Info(
		"Application successfully started",
		slog.String("address", ":"+app.Cfg.HttpServerCfg.Port),
	)
	if err := s.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			app.Logger.Error("Server failed", logger.WithErr(err))
			return
		}
	}

	if cerr := <-eChan; cerr != nil {
		app.Logger.Error("Failed graceful shutdown", logger.WithErr(cerr))
		return
	}

	app.Logger.Info("Graceful shutdown completed successfully")
}
