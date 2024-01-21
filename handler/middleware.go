package handler

import (
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/la0rg/lingo/logger"
)

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
