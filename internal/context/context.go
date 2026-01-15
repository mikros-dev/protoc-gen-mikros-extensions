package context

import (
	"context"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/log"
)

type contextKey int

const (
	loggerKey contextKey = iota
)

// WithLogger returns a new context with the provided logger.
func WithLogger(ctx context.Context, logger log.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// LoggerFromContext retrieves the logger from the context. If no logger is
// found, it returns a no-op logger to prevent nil pointer panics.
func LoggerFromContext(ctx context.Context) log.Logger {
	if logger, ok := ctx.Value(loggerKey).(log.Logger); ok {
		return logger
	}

	return log.New(log.LoggerOptions{
		Verbose: false,
	})
}
