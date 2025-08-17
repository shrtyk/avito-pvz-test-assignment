package http

import (
	"errors"
	"net/http"
)

type AppHandler func(http.ResponseWriter, *http.Request) error

func (h AppHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h(w, r); err != nil {
		var httpErr *HTTPError
		switch {
		case errors.As(err, &httpErr):
			WriteHTTPError(w, r, httpErr)
		default:
			WriteHTTPError(w, r, InternalError(err))
		}
	}
}
