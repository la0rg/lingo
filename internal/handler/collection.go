package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/la0rg/lingo/internal/maps"
	"github.com/la0rg/lingo/internal/model"
	"github.com/la0rg/lingo/internal/service"
	istrings "github.com/la0rg/lingo/internal/strings"
	"github.com/la0rg/lingo/web"
)

func (h *handlers) Collection(w http.ResponseWriter, r *http.Request) error {
	data := map[string]any{"LanguageSelector": h.languageSelectorFromCookie(r)}
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

	collection := h.collection.GetBySlug(model.CollectionSlug(source, target))
	if collection == nil {
		return ErrCollectionNotFound
	}

	wordsStr := r.Form.Get("words")
	if wordsStr == "" {
		return ErrEmptyWordList
	}

	words := maps.Unique(trimWords(istrings.SplitAny(wordsStr, "\n,")))
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

	return h.HTML(w, r, web.CollectionDownload, map[string]any{
		"URL":            file.URL,
		"CardsCount":     file.CardsCount,
		"RequestedCount": len(words),
		"MissingWords":   file.MissingWords,
	})
}

// trimWords trims and removes empty words from the slice. Words are converted to lowercase.
// The slice is modified in place.
func trimWords(words []string) []string {
	total := 0
	for _, word := range words {
		word := strings.ToLower(strings.TrimSpace(word))
		if word != "" {
			words[total] = word
			total++
		}
	}

	return words[:total]
}
