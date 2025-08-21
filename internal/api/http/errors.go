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
	return fmt.Sprintf("HTTP Code: %d. Message: %s. Error: %s", e.Code, e.Message, e.Err)
}

func (e *HTTPError) Unwrap() error {
	return e.Err
}

func BadRequestBodyError(err error) *HTTPError {
	return &HTTPError{
		Code:    http.StatusBadRequest,
		Message: "Badly formed request body",
		Err:     err,
	}
}

func BadRequestQueryParamsError(err error) *HTTPError {
	return &HTTPError{
		Code:    http.StatusBadRequest,
		Message: "Wrong formated url query params",
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

func ValidationError(err error) *HTTPError {
	return &HTTPError{
		// Would be better to use StatusUnprocessableEntity but i followed swagger.yaml
		Code:    http.StatusBadRequest,
		Message: "Invalid data in request body",
		Err:     err,
	}
}
