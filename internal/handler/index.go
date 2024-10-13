package handler

import (
	"net/http"

	"github.com/la0rg/lingo/internal/model"
	"github.com/la0rg/lingo/web"
)

func (h *handlers) Index(w http.ResponseWriter, r *http.Request) error {
	collection, ok := model.Collections["de-en"]
	if !ok {
		return ErrCollectionNotFound
	}

	card, err := h.card.Get(r.Context(), collection, "Bad")
	if err != nil {
		return err
	}

	return h.HTML(w, r, web.Index, card)
}
