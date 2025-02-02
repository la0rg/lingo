package handler

import (
	"net/http"

	"github.com/la0rg/lingo/internal/model"
)

func (h *handlers) CardPreview(w http.ResponseWriter, r *http.Request) error {
	source, target := r.URL.Query().Get("source"), r.URL.Query().Get("target")
	if source == "" || target == "" {
		return ErrInvalidLanguage
	}

	collection := h.collection.GetBySlug(model.CollectionSlug(source, target))
	if collection == nil {
		return ErrCollectionNotFound
	}

	word := r.URL.Query().Get("word")
	if word == "" {
		return ErrEmptyWord
	}

	card, err := h.card.Get(r.Context(), collection, word)
	if err != nil {
		return err
	}
	if card == nil {
		return ErrInvalidWord
	}

	return h.HTML(w, r, collection.HTMLTemplateName, card)
}
