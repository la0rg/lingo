package handler

import (
	"net/http"
	"runtime/debug"
	"slices"
	"strings"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/la0rg/lingo/internal/config"
	"github.com/la0rg/lingo/internal/logger"
)

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

			apiKey := parts[1]
			if !slices.Contains(c.APIKeys, apiKey) {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}

func RequestLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		defer func() {
			logger.WithContext(r.Context()).Info(
				"request handled",
				"status", ww.Status(),
				"bytes", ww.BytesWritten(),
				"headers", ww.Header(),
				"latency", time.Since(t1).Milliseconds(),
				"url", r.URL.String(),
			)
		}()

		h.ServeHTTP(ww, r)
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
