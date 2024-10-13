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

	data := struct {
		Languages model.LanguageSelector
		Card      *model.Card
	}{
		Languages: model.Languages,
		Card:      card,
	}
	return h.HTML(w, r, web.Index, data)
}
