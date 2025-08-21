package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sync"

	"github.com/shrtyk/avito-pvz-test-assignment/internal/adapters/driving/grpc"
	appHttp "github.com/shrtyk/avito-pvz-test-assignment/internal/adapters/driving/http"
	"github.com/shrtyk/avito-pvz-test-assignment/pkg/logger"
)

func (app Application) Serve(ctx context.Context) {
	var wg sync.WaitGroup

	httpServ := http.Server{
		Addr:         ":" + app.Cfg.HttpServerCfg.Port,
		Handler:      appHttp.NewRouter(app.AppService, app.TokenService, app.Logger),
		IdleTimeout:  app.Cfg.HttpServerCfg.IdleTimeout,
		WriteTimeout: app.Cfg.HttpServerCfg.WriteTimeout,
		ReadTimeout:  app.Cfg.HttpServerCfg.ReadTimeout,
		ErrorLog:     slog.NewLogLogger(app.Logger.Handler(), slog.LevelError),
	}
	grpcServ := grpc.NewGRPCServer(&wg, app.AppService, app.Logger, app.Cfg.GrpcServerCfg.Port)

	eChan := make(chan error, 1)
	go func() {
		<-ctx.Done()

		tCtx, tCancel := context.WithTimeout(context.Background(), app.Cfg.AppCfg.ShutdownTimeout)
		defer tCancel()

		eChan <- grpcServ.Shutdown(tCtx)
		eChan <- httpServ.Shutdown(tCtx)
	}()

	app.Logger.Info(
		"GRPC server successfully started",
		slog.String("address", ":"+app.Cfg.GrpcServerCfg.Port),
	)
	grpcServ.MustStart()

	app.Logger.Info(
		"HTTP server successfully started",
		slog.String("address", ":"+app.Cfg.HttpServerCfg.Port),
	)
	if err := httpServ.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			app.Logger.Error("Server failed", logger.WithErr(err))
			return
		}
	}

	wg.Wait()

	if cerr := <-eChan; cerr != nil {
		app.Logger.Error("Failed graceful shutdown", logger.WithErr(cerr))
		return
	}

	app.Logger.Info("Graceful shutdown completed successfully")
}
