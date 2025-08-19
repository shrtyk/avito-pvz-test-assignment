package http

import (
	"errors"
	"net/http"
)

type AppHandler func(http.ResponseWriter, *http.Request) error

func Handle(ch AppHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := ch(w, r); err != nil {
			var httpErr *HTTPError
			switch {
			case errors.As(err, &httpErr):
				WriteHTTPError(w, r, httpErr)
			default:
				WriteHTTPError(w, r, InternalError(err))
			}
		}
	}
}
