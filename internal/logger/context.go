package logger

import (
	"context"
	"log/slog"

	"github.com/la0rg/lingo/internal/ctxval"
)

func WithContext(ctx context.Context) *slog.Logger {
	return slog.With("request_id", ctxval.RequestID(ctx))
}
