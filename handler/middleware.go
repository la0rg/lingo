package handler

import (
	"net/http"
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
