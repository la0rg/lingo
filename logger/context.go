package logger

import (
	"context"
	"log/slog"

	"github.com/go-chi/chi/middleware"
)

func WithContext(ctx context.Context) *slog.Logger {
	return slog.With(
		"request_id", middleware.GetReqID(ctx),
	)
}
