package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/la0rg/lingo/internal/logger"
	"github.com/la0rg/lingo/internal/model"
	"github.com/la0rg/lingo/web"
)

var (
	ErrCollectionNotFound = NotFound("The language selection is not supported")
	ErrEmptyWord          = BadRequest("Word is empty")
	ErrEmptyWordList      = BadRequest("Word list is empty")
	ErrNoValidWords       = BadRequest("The collection does not contain any valid words")
	ErrInvalidWord        = BadRequest("Unable to create a card for this word")
	ErrInvalidLanguage    = BadRequest("Invalid language selection")
	ErrRateLimited        = &HTTPError{
		Code:    http.StatusTooManyRequests,
		Status:  http.StatusText(http.StatusTooManyRequests),
		Message: "Too many words requested, please try again in a minute",
	}
	ErrBodyTooLarge = &HTTPError{
		Code:    http.StatusRequestEntityTooLarge,
		Status:  http.StatusText(http.StatusRequestEntityTooLarge),
		Message: "Request is too large",
	}
)

func ErrTooManyWords(limit int) *HTTPError {
	return BadRequest(fmt.Sprintf("Too many words, the maximum is %d", limit))
}

func ErrWordTooLong(limit int) *HTTPError {
	return BadRequest(fmt.Sprintf("Word is too long, the maximum is %d characters", limit))
}

type HTTPError struct {
	Code    int    `json:"code"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("%s: %s", e.Status, e.Message)
}

func (e *HTTPError) Is4xx() bool {
	return e.Code > 399 && e.Code < 500
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
	Message: "Unexpected error happened, please try again later",
}

type HandlerFunc = func(http.ResponseWriter, *http.Request) error

func ErrorHandler(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h(w, r)
		if err == nil {
			return
		}

		httpErr := mapError(r.Context(), err)
		logger.WithContext(r.Context()).Info("request failed", "error", err, "code", httpErr.Code)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(httpErr.Code)
		JSON(w, r, httpErr)
	}
}

func (h *handlers) ErrorHandlerHTML(handler HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := handler(w, r)
		if err == nil {
			return
		}

		httpErr := mapError(r.Context(), err)
		logger.WithContext(r.Context()).Info("request failed", "error", err, "code", httpErr.Code)

		w.WriteHeader(httpErr.Code)
		if err := h.HTML(w, r, web.Error, httpErr); err != nil {
			logger.WithContext(r.Context()).Error("unable to render error page", "error", err)
		}
	}
}

// requestError maps a request validation error to an HTTP error.
func requestError(err error) error {
	var tooMany *model.TooManyWordsError
	var tooLong *model.WordTooLongError
	var maxBytes *http.MaxBytesError

	switch {
	case errors.Is(err, model.ErrEmptyWordList):
		return ErrEmptyWordList
	case errors.As(err, &tooMany):
		return ErrTooManyWords(tooMany.Max)
	case errors.As(err, &tooLong):
		return ErrWordTooLong(tooLong.Max)
	case errors.As(err, &maxBytes):
		return ErrBodyTooLarge
	default:
		return BadRequest("Invalid request format")
	}
}

func mapError(ctx context.Context, err error) *HTTPError {
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return httpErr
	}

	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		return ErrBodyTooLarge
	}

	logger.WithContext(ctx).Error("unexpected error", "error", err)
	return ErrInternal
}
