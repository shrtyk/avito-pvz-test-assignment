package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	appHttp "github.com/shrtyk/avito-backend-spring-2025/internal/adapters/driving/http"
	"github.com/shrtyk/avito-backend-spring-2025/internal/core/domain/auth"
	"github.com/shrtyk/avito-backend-spring-2025/pkg/logger"
)

func (app Application) Serve(ctx context.Context) {
	s := http.Server{
		Addr:         ":" + app.Cfg.HttpServerCfg.Port,
		Handler:      app.router(),
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
		"Aplication successfully started",
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

func (app *Application) router() *chi.Mux {
	mws := appHttp.NewMiddlewares(app.TokenService, app.Logger)
	h := appHttp.NewHandlers(app.AppService, app.TokenService)

	r := chi.NewRouter()

	r.Use(mws.PanicRecoveryMW, mws.LoggingMW)
	r.Post("/dummyLogin", appHttp.Handle(h.DummyLoginHandler))

	// Authenticated only:
	r.Group(func(r chi.Router) {
		r.Use(mws.AuthenticationMW)

		// Moderators only:
		r.Group(func(r chi.Router) {
			r.Use(mws.AuthorizeRoles(auth.UserRoleModerator))

			r.Post("/pvz", appHttp.Handle(h.NewPVZHandler))
		})

		// Employees only:
		r.Group(func(r chi.Router) {
			r.Use(mws.AuthorizeRoles(auth.UserRoleEmployee))

			r.Post("/receptions", appHttp.Handle(h.NewReceptionHandler))
			r.Post("/products", appHttp.Handle(h.AddProductHandler))
			r.Post("/{pvzId}/delete_last_product", appHttp.Handle(h.DeleteLastProductHandler))
			r.Post("/{pvzId}/close_last_reception", appHttp.Handle(h.CloseReceptionHandler))
		})

		// Moderators and employees:
		r.Group(func(r chi.Router) {
			r.Use(mws.AuthorizeRoles(auth.UserRoleEmployee, auth.UserRoleModerator))

			r.Get("/pvz", appHttp.Handle(h.GetPvzHandler))
		})

	})

	return r
}
