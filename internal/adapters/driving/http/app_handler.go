package http

import (
	"errors"
	"net/http"

	"github.com/shrtyk/avito-backend-spring-2025/pkg/logger"
)

type appHandler func(http.ResponseWriter, *http.Request) error

func (h appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l := logger.FromCtx(r.Context())
	if err := h(w, r); err != nil {
		l.Error("Error during processing request", logger.WithErr(err))
		var httpErr *HTTPError
		switch {
		case errors.As(err, &httpErr):
			WriteError(w, r, httpErr.Code, httpErr.Message)
		default:
			WriteError(w, r, http.StatusInternalServerError, "Internal error")
		}
	}
}
