package middleware

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/shrtyk/avito-backend-spring-2025/internal/core/ports"
)

type Middlewares struct {
	tService ports.TokensService
	log      *slog.Logger
}

func (m Middlewares) PanicRecoveryMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				m.log.Error("error occured", "error", fmt.Sprintf("%s", err))
				w.Header().Set("Connection", "close")
				http.Error(
					w,
					"The server encountered a problem and could not process your request",
					http.StatusInternalServerError,
				)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
