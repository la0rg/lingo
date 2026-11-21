package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/la0rg/lingo/internal/model"
)

func (h *handlers) APKG(w http.ResponseWriter, r *http.Request) error {
	collectionSlug := strings.ToLower(chi.URLParam(r, "collection"))

	collection := h.collection.GetBySlug(collectionSlug)
	if collection == nil {
		return ErrCollectionNotFound
	}

	var request model.APKGRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return ErrBodyTooLarge
		}
		return BadRequest("invalid request format")
	}

	words := normalizeWords(request.Words)
	if len(words) == 0 {
		return BadRequest("no words provided")
	}

	if err := h.validateWords(words...); err != nil {
		return err
	}

	file, err := h.apkg.CreateOnS3(r.Context(), collection, words...)
	if err != nil {
		return err
	}

	JSON(w, r, model.APKGResponse{URL: file.URL})
	return nil
}
