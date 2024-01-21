package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/la0rg/lingo/logger"
)

type HTTPError struct {
	Code    int    `json:"code"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("%s: %s", e.Status, e.Message)
}

func BadRequest(message string) *HTTPError {
	return &HTTPError{
		Code:    http.StatusBadRequest,
		Status:  http.StatusText(http.StatusBadRequest),
		Message: message,
	}
}

func NotFound(message string) *HTTPError {
	return &HTTPError{
		Code:    http.StatusNotFound,
		Status:  http.StatusText(http.StatusNotFound),
		Message: message,
	}
}

var ErrInternal = &HTTPError{
	Code:    http.StatusInternalServerError,
	Status:  http.StatusText(http.StatusInternalServerError),
	Message: "unexpected error happened, please try again later",
}

var errorMapping = map[error]*HTTPError{
	ErrCollectionNotFound: NotFound("collection not found"),
}

type HandlerFunc = func(http.ResponseWriter, *http.Request) error

func ErrorHandler(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h(w, r)
		if err == nil {
			return
		}

		httpErr := mapError(r.Context(), err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(httpErr.Code)
		JSON(w, r, httpErr)
	}
}

func mapError(ctx context.Context, err error) *HTTPError {
	for e, httpErr := range errorMapping {
		if errors.Is(err, e) {
			logger.WithContext(ctx).Info("request failed", "error", err)
			return httpErr
		}
	}

	logger.WithContext(ctx).Error("unexpected error", "error", err)
	return ErrInternal
}
