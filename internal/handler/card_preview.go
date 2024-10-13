package handler

import (
	"net/http"

	"github.com/la0rg/lingo/internal/model"
	"github.com/la0rg/lingo/web"
)

func (h *handlers) CardPreview(w http.ResponseWriter, r *http.Request) error {
	// TODO: get collection from request
	collection, ok := model.Collections["de-en"]
	if !ok {
		return ErrCollectionNotFound
	}

	word := r.URL.Query().Get("word")
	if word == "" {
		return BadRequest("empty word")
	}

	card, err := h.card.Get(r.Context(), collection, word)
	if err != nil {
		return err
	}

	return h.HTML(w, r, web.CardPreview, card)
}
