package handler

import (
	"net/http"
	"strings"

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

	word := strings.TrimSpace(r.URL.Query().Get("word"))
	if word == "" {
		return ErrEmptyWord
	}

	if err := h.validateWords(word); err != nil {
		return err
	}

	if !h.limiter.Allow(r, 1) {
		return ErrRateLimited
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
