package logger

import (
	"context"
	"log/slog"

	"github.com/go-chi/chi/middleware"
)

const RequestIDKey = "request_id"

// type ctxKey struct{}

// type ctxValues map[string]any

// func WithValue(ctx context.Context, key string, value any) context.Context {
// 	values, ok := ctx.Value(ctxKey{}).(ctxValues)
// 	if !ok {
// 		values = make(ctxValues)
// 	}

// 	values[key] = value
// 	return context.WithValue(ctx, ctxKey{}, values)
// }

func WithContext(ctx context.Context) *slog.Logger {
	return slog.With(RequestIDKey, middleware.GetReqID(ctx))
}
