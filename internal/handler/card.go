package handler

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/la0rg/lingo/internal/model"
)

func (h *handlers) Card(w http.ResponseWriter, r *http.Request) error {
	collectionName := strings.ToLower(chi.URLParam(r, "collection"))

	collection, ok := model.CollectionsMap[collectionName]
	if !ok {
		return ErrCollectionNotFound
	}

	word := strings.TrimSpace(chi.URLParam(r, "word"))
	if len(word) == 0 {
		return ErrEmptyWord
	}

	card, err := h.card.Get(r.Context(), collection, word)
	if err != nil {
		return err
	}

	JSON(w, r, card)
	return nil
}
