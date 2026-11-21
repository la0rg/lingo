package handler

import (
	"net/http"
	"strings"

	"github.com/la0rg/lingo/internal/model"
)

func (h *handlers) Card(w http.ResponseWriter, r *http.Request) error {
	collectionSlug := strings.ToLower(r.PathValue("collection"))

	collection := h.collection.GetBySlug(collectionSlug)
	if collection == nil {
		return ErrCollectionNotFound
	}

	word := strings.TrimSpace(r.PathValue("word"))
	if len(word) == 0 {
		return ErrEmptyWord
	}

	if err := model.ValidateWord(word, h.conf.Limits.MaxWordLength); err != nil {
		return requestError(err)
	}

	card, err := h.card.Get(r.Context(), collection, word)
	if err != nil {
		return err
	}

	JSON(w, r, card)
	return nil
}
