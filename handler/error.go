package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/la0rg/lingo/logger"
)

var errorMapping = map[error]int{}

type HandlerFunc = func(http.ResponseWriter, *http.Request) error

func ErrorHandler(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h(w, r)
		if err == nil {
			return
		}

		code := mapError(r.Context(), err)
		http.Error(w, http.StatusText(code), code)
	}
}

func mapError(ctx context.Context, err error) int {
	for e, code := range errorMapping {
		if errors.Is(err, e) {
			return code
		}
	}

	logger.WithContext(ctx).Error("unknown error", "error", err)
	return http.StatusInternalServerError
}
