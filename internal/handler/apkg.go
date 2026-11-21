package handler

import (
	"encoding/json"
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
		return requestError(err)
	}

	request.Normalize()
	if err := request.Validate(h.conf.Limits.MaxWords, h.conf.Limits.MaxWordLength); err != nil {
		return requestError(err)
	}

	file, err := h.apkg.CreateOnS3(r.Context(), collection, request.Words...)
	if err != nil {
		return err
	}

	JSON(w, r, model.APKGResponse{URL: file.URL})
	return nil
}
