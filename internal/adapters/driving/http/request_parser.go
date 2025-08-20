package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shrtyk/avito-pvz-test-assignment/internal/adapters/driving/http/dto"
	"github.com/shrtyk/avito-pvz-test-assignment/internal/core/ports/auth"
	xerr "github.com/shrtyk/avito-pvz-test-assignment/pkg/xerrors"
	"github.com/tomasen/realip"
)

func ReadJson[T any](w http.ResponseWriter, r *http.Request, dst T) error {
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

func IdParam(r *http.Request) (int64, error) {
	idParam := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}
	return id, nil
}

func PvzIdParam(r *http.Request) (*uuid.UUID, error) {
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

func UserAgentAndIP(r *http.Request) (string, string) {
	return r.UserAgent(), realip.FromRequest(r)
}

func BearerToken(r *http.Request) (string, error) {
	op := "helpers.ExtractBearerToken"
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", xerr.NewErr(op, auth.NotAuthenticated)
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return "", xerr.NewErr(op, auth.InvalidJwt)
	}

	return tokenString, nil
}

func PvzParamsFromURL(r *http.Request) (*dto.GetPvzParams, error) {
	params := &dto.GetPvzParams{}
	query := r.URL.Query()

	if startDateStr := query.Get("startDate"); startDateStr != "" {
		sd, err := time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			return nil, wrapConvertionError("startDate", startDateStr, "time.Time", err)
		}
		params.StartDate = &sd
	}

	if endDateStr := query.Get("endDate"); endDateStr != "" {
		ed, err := time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			return nil, wrapConvertionError("endDate", endDateStr, "time.Time", err)
		}
		params.EndDate = &ed
	}

	if pageStr := query.Get("page"); pageStr != "" {
		pg, err := strconv.Atoi(pageStr)
		if err != nil {
			return nil, wrapConvertionError("page", pageStr, "int", err)
		}
		params.Page = &pg
	}

	if limitStr := query.Get("limit"); limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil {
			return nil, wrapConvertionError("limit", limitStr, "int", err)
		}
		params.Limit = &l
	}

	return params, nil
}

func wrapConvertionError(paramName, param, paramKind string, err error) error {
	tmp := "failed to convert '%s' query param '%s' into '%s': %w"
	return fmt.Errorf(tmp, paramName, param, paramKind, err)
}
