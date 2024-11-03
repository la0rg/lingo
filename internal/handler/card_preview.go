package handler

import (
	"errors"
	"net/http"

	"github.com/la0rg/lingo/internal/model"
	"github.com/la0rg/lingo/internal/service"
	"github.com/la0rg/lingo/web"
)

func (h *handlers) CardPreview(w http.ResponseWriter, r *http.Request) error {
	source, target := r.URL.Query().Get("source"), r.URL.Query().Get("target")
	if source == "" || target == "" {
		return ErrInvalidLanguage
	}

	collection, ok := model.CollectionsMap[model.CollectionSlug(source, target)]
	if !ok {
		return ErrCollectionNotFound
	}

	word := r.URL.Query().Get("word")
	if word == "" {
		return ErrEmptyWord
	}

	card, err := h.card.Get(r.Context(), collection, word)
	if errors.Is(err, service.ErrInvalidWord) {
		return ErrInvalidWord
	}
	if err != nil {
		return err
	}

	return h.HTML(w, r, web.CardPreview, card)
}
