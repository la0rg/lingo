package handler

import (
	"net/http"

	"github.com/la0rg/lingo/internal/model"
	"github.com/la0rg/lingo/web"
)

const defaultWord = "Bad"

var (
	defaultLanguageSelector = model.NewLanguageSelector("de", "en")
	defaultCollectionSlug   = model.CollectionSlug(
		defaultLanguageSelector.SourceCountryCode,
		defaultLanguageSelector.TargetCountryCode,
	)
)

func (h *handlers) Index(w http.ResponseWriter, r *http.Request) error {
	collection, ok := model.CollectionsMap[defaultCollectionSlug]
	if !ok {
		return ErrCollectionNotFound
	}

	card, err := h.card.Get(r.Context(), collection, defaultWord)
	if err != nil {
		return err
	}

	data := map[string]any{
		"LanguageSelector": defaultLanguageSelector,
		"Card":             card,
	}
	return h.HTML(w, r, web.Index, data)
}
