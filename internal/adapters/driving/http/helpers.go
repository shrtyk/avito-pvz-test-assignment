package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shrtyk/avito-backend-spring-2025/internal/core/domain/auth"
	"github.com/shrtyk/avito-backend-spring-2025/pkg/logger"
	"github.com/tomasen/realip"
)

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

func ReadJSON[T any](w http.ResponseWriter, r *http.Request, dst T) error {
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)
		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		case errors.As(err, &invalidUnmarshalError):
			// Shouldn't occur at all. Only possible if wrong value passed as dst
			panic(err)
		default:
			return err
		}
	}

	if err := dec.Decode(&struct{}{}); err != nil {
		if !errors.Is(err, io.EOF) {
			return errors.New("body must contain a single JSON value")
		}
	}

	return nil
}

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

func ReadIDParam(r *http.Request) (int64, error) {
	idParam := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}
	return id, nil
}

func ReadPvzIDParam(r *http.Request) (*uuid.UUID, error) {
	strId := chi.URLParam(r, "pvzId")
	if err := uuid.Validate(strId); err != nil {
		return nil, err
	}

	pvzId, err := uuid.Parse(strId)
	if err != nil {
		return nil, err
	}

	return &pvzId, nil
}

func GetUserAgentAndIP(r *http.Request) (string, string) {
	return r.UserAgent(), realip.FromRequest(r)
}

func ExtractBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", auth.ErrNotAuthenticated
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return "", auth.ErrInvalidJWT
	}

	return tokenString, nil
}
