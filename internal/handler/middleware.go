package handler

import (
	"crypto/subtle"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/la0rg/lingo/internal/config"
	"github.com/la0rg/lingo/internal/ctxval"
	"github.com/la0rg/lingo/internal/logger"
)

const (
	requestIDHeader       = "X-Request-ID"
	cloudflareRayIDHeader = "Cf-Ray"

	// maxBodyBytes is enough for MaxWords words of MaxWordLength runes with the default limits.
	maxBodyBytes = 64 << 10
)

// MaxBodySize caps the size of request bodies.
func MaxBodySize(limit int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, limit)
			next.ServeHTTP(w, r)
		})
	}
}

func APIKey(c *config.Config) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
			if authHeader == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			if !validAPIKey(c.APIKeys, parts[1]) {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}

// validAPIKey compares the key against every configured key in constant time.
func validAPIKey(keys []string, key string) bool {
	valid := 0
	for _, k := range keys {
		valid |= subtle.ConstantTimeCompare([]byte(k), []byte(key))
	}
	return valid == 1
}

var sensitiveHeaders = []string{"Authorization", "Cookie", "Set-Cookie", "Proxy-Authorization", "X-Api-Key"}

// redactHeaders returns a copy of h with sensitive header values replaced.
func redactHeaders(h http.Header) http.Header {
	c := h.Clone()
	for _, name := range sensitiveHeaders {
		if _, ok := c[http.CanonicalHeaderKey(name)]; ok {
			c.Set(name, "[REDACTED]")
		}
	}
	return c
}

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

func RequestLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		h.ServeHTTP(ww, r)

		logger.WithContext(r.Context()).Info(
			"request handled",
			"status", ww.Status(),
			"bytes", ww.BytesWritten(),
			"request_headers", redactHeaders(r.Header),
			"response_headers", redactHeaders(ww.Header()),
			"latency", time.Since(t1).Milliseconds(),
			"url", r.URL.String(),
		)
	})
}

func Recoverer(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil && rvr != http.ErrAbortHandler {
				logger.WithContext(r.Context()).Error("panic recovered", "stack", debug.Stack())
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func RequestID(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		requestID := r.Header.Get(requestIDHeader)
		if requestID == "" {
			requestID = r.Header.Get(cloudflareRayIDHeader)
		}
		if requestID == "" {
			requestID = uuid.New().String()
		}

		ctx = ctxval.WithRequestID(ctx, requestID)

		w.Header().Set(requestIDHeader, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
