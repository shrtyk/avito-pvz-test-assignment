package logger

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoggerCreation(t *testing.T) {
	prodLogger := MustCreateNewLogger(prodEnv)
	devLoggerr := MustCreateNewLogger(devEnv)

	assert.IsType(t, &slog.Logger{}, prodLogger)
	assert.IsType(t, &slog.Logger{}, devLoggerr)

	assert.Panics(t, func() {
		MustCreateNewLogger("wrong enviroment")
	})
}

func TestLoggerFromCtx(t *testing.T) {
	l := MustCreateNewLogger(devEnv)
	ctx := context.WithValue(context.Background(), logCtxKey, l)
	ctxLogger := FromCtx(ctx)
	assert.IsType(t, &slog.Logger{}, ctxLogger)

	ctxWrong := context.WithValue(context.Background(), logCtxKey, "123")
	fbLogger := FromCtx(ctxWrong)
	assert.IsType(t, &slog.Logger{}, fbLogger)
	assert.NotEqual(t, ctxLogger.Handler(), fbLogger.Handler())
}

func TestNewTestLogger(t *testing.T) {
	l, buf := NewTestLogger()
	assert.NotNil(t, l)
	assert.NotNil(t, buf)

	l.Info("test message")
	assert.Contains(t, buf.String(), "test message")
}