package http

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/shrtyk/avito-pvz-test-assignment/internal/core/domain/auth"
	pAuth "github.com/shrtyk/avito-pvz-test-assignment/internal/core/ports/auth"
	aService "github.com/shrtyk/avito-pvz-test-assignment/internal/core/ports/service"
)

type Router struct {
	chi.Router
	aService aService.Service
	tService pAuth.TokenService
	logger   *slog.Logger
}

func NewRouter(
	aService aService.Service,
	tService pAuth.TokenService,
	logger *slog.Logger,
) *Router {
	r := &Router{
		Router:   chi.NewRouter(),
		aService: aService,
		tService: tService,
		logger:   logger,
	}

	r.initRoutes()

	return r
}

func (r *Router) initRoutes() {
	mws := NewMiddlewares(r.tService, r.logger)
	h := NewHandlers(r.aService, r.tService)

	r.Use(mws.PanicRecoveryMW, mws.LoggingMW)
	r.Post("/dummyLogin", Handle(h.DummyLoginHandler))
	r.Get("/healthz", Handle(h.HealthZ))

	r.Post("/register", Handle(h.RegisterUserHandler))
	r.Post("/login", Handle(h.LoginUserHandler))
	r.Post("/tokens/refresh", Handle(h.RefreshTokensHandler))

	// Authenticated only:
	r.Group(func(r chi.Router) {
		r.Use(mws.AuthenticationMW)

		// Moderators only:
		r.Group(func(r chi.Router) {
			r.Use(mws.AuthorizeRoles(auth.UserRoleModerator))

			r.Post("/pvz", Handle(h.NewPVZHandler))
		})

		// Employees only:
		r.Group(func(r chi.Router) {
			r.Use(mws.AuthorizeRoles(auth.UserRoleEmployee))

			r.Post("/receptions", Handle(h.NewReceptionHandler))
			r.Post("/products", Handle(h.AddProductHandler))
			r.Post("/{pvzId}/delete_last_product", Handle(h.DeleteLastProductHandler))
			r.Post("/{pvzId}/close_last_reception", Handle(h.CloseReceptionHandler))
		})

		// Moderators and employees:
		r.Group(func(r chi.Router) {
			r.Use(mws.AuthorizeRoles(auth.UserRoleEmployee, auth.UserRoleModerator))

			r.Get("/pvz", Handle(h.GetPvzHandler))
		})
	})
}
