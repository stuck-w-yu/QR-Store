package logger

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLog_DefaultNotNull(t *testing.T) {
	assert.NotNil(t, Log, "Log should have a default logger and not be nil")
	assert.NotPanics(t, func() {
		Log.Info("testing default logger info message", "key", "value")
	})
}

func TestInit(t *testing.T) {
	t.Run("production", func(t *testing.T) {
		Init("production")
		assert.NotNil(t, Log)
		assert.NotPanics(t, func() {
			Log.Info("prod log test")
		})
	})

	t.Run("development", func(t *testing.T) {
		Init("development")
		assert.NotNil(t, Log)
		assert.NotPanics(t, func() {
			Log.Info("dev log test")
		})
	})
}

func TestPrettyHandler(t *testing.T) {
	buf := &bytes.Buffer{}
	handler := NewPrettyHandler(buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	l := slog.New(handler)
	l.Info("hello world", "component", "test")
	assert.Contains(t, buf.String(), "hello world")
	assert.Contains(t, buf.String(), "component=")
	assert.Contains(t, buf.String(), "test")
}
