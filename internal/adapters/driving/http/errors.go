package http

import (
	"fmt"
	"net/http"
)

type HTTPError struct {
	Code    int
	Message string
	Err     error
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s. Error: %s", e.Code, e.Message, e.Err)
}

func BadRequestError(err error) *HTTPError {
	return &HTTPError{
		Code:    http.StatusBadRequest,
		Message: "Badly formed request body",
		Err:     err,
	}
}

func InternalError(err error) *HTTPError {
	return &HTTPError{
		Code:    http.StatusInternalServerError,
		Message: "Internal error",
		Err:     err,
	}
}
