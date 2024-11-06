package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/la0rg/lingo/internal/model"
	"github.com/la0rg/lingo/internal/service"
	pkgstrings "github.com/la0rg/lingo/pkg/strings"
	"github.com/la0rg/lingo/web"
)

func (h *handlers) Collection(w http.ResponseWriter, r *http.Request) error {
	data := map[string]any{"LanguageSelector": defaultLanguageSelector}
	return h.HTML(w, r, web.Collection, data)
}

func (h *handlers) CreateCollection(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return BadRequest("Invalid request format")
	}

	source, target := r.Form.Get("source"), r.Form.Get("target")
	if source == "" || target == "" {
		return ErrInvalidLanguage
	}

	collection, ok := model.CollectionsMap[model.CollectionSlug(source, target)]
	if !ok {
		return ErrCollectionNotFound
	}

	wordsStr := r.Form.Get("words")
	if wordsStr == "" {
		return ErrEmptyWordList
	}

	words := trimWords(pkgstrings.SplitAny(wordsStr, "\n,"))
	if len(words) == 0 {
		return ErrEmptyWordList
	}

	file, err := h.apkg.CreateOnS3(r.Context(), collection, words...)
	if err != nil {
		if errors.Is(err, service.ErrEmptyCards) {
			return ErrNoValidWords
		}

		return err
	}

	data := map[string]any{
		"URL":            file.URL,
		"CardsCount":     file.CardsCount,
		"RequestedCount": len(words),
	}

	return h.HTML(w, r, web.CollectionDownload, data)
}

// trimWords trims spaces and removes empty words from the slice.
// The slice is modified in place.
func trimWords(words []string) []string {
	total := 0
	for _, word := range words {
		word := strings.TrimSpace(word)
		if word != "" {
			words[total] = word
			total++
		}
	}

	return words[:total]
}
