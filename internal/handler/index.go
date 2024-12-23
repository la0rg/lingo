package handler

import (
	"net/http"
	"strings"

	"github.com/la0rg/lingo/internal/model"
	"github.com/la0rg/lingo/web"
)

const defaultWord = "Bad"

var defaultLanguageSelector = model.NewLanguageSelector("de", "en")

func (h *handlers) Index(w http.ResponseWriter, r *http.Request) error {
	langSelector := h.languageSelectorFromCookie(r)
	collection, ok := model.CollectionsMap[langSelector.CollectionSlug()]
	if !ok {
		http.SetCookie(w, &http.Cookie{Name: languageCookie, MaxAge: -1}) // reset the language pair
		return ErrCollectionNotFound
	}

	card, err := h.card.Get(r.Context(), collection, defaultWord)
	if err != nil {
		return err
	}

	data := map[string]any{
		"LanguageSelector": langSelector,
		"Card":             card,
	}
	return h.HTML(w, r, web.Index, data)
}

func (h *handlers) languageSelectorFromCookie(r *http.Request) model.LanguageSelector {
	langCookie, err := r.Cookie(languageCookie)
	if err != nil || langCookie == nil {
		return defaultLanguageSelector
	}

	parts := strings.Split(langCookie.Value, ",")
	if len(parts) != 2 || len(parts[0]) == 0 || len(parts[1]) == 0 {
		return defaultLanguageSelector
	}

	return model.NewLanguageSelector(parts[0], parts[1])
}
