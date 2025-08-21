package http

import (
	"bytes"
	"encoding/json"
	"maps"
	"net/http"

	"github.com/shrtyk/avito-pvz-test-assignment/pkg/logger"
)

func WriteHTTPError(w http.ResponseWriter, r *http.Request, e *HTTPError) {
	l := logger.FromCtx(r.Context())
	if e.Code >= 500 {
		l.Error("Server error", logger.WithErr(e))
	} else {
		l.Info("Client error", logger.WithErr(e))
	}

	err := WriteJSON(
		w,
		map[string]string{"message": e.Message},
		e.Code,
		nil,
	)
	if err != nil {
		l.Error("Failed to response with error", logger.WithErr(err))
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func WriteJSON[T any](w http.ResponseWriter, data T, status int, headers http.Header) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Uncomment to add extra readability during manual testing:
	b = append(b, '\n')
	buf := bytes.Buffer{}
	if err := json.Indent(&buf, b, "", "\t"); err != nil {
		return err
	}
	b = buf.Bytes()

	maps.Copy(w.Header(), headers)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if _, err = w.Write(b); err != nil {
		return err
	}

	return nil
}
