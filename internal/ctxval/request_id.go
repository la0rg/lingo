package ctxval

import "context"

type ctxKeyType int

const (
	requestIDKey ctxKeyType = iota
)

// RequestID returns the request ID from the context.
func RequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if reqID, ok := ctx.Value(requestIDKey).(string); ok {
		return reqID
	}

	return ""
}

// WithRequestID returns a new context with the request ID set.
func WithRequestID(ctx context.Context, reqID string) context.Context {
	return context.WithValue(ctx, requestIDKey, reqID)
}
