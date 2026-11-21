package handler

import (
	"errors"
	"net/http"
	"strings"
	"unicode/utf8"

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
		return formError(err)
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

	words := normalizeWords(istrings.SplitAny(wordsStr, "\n,"))
	if len(words) == 0 {
		return ErrEmptyWordList
	}

	if err := h.validateWords(words...); err != nil {
		return err
	}

	if !h.limiter.Allow(r, len(words)) {
		return ErrRateLimited
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

// normalizeWords trims, lowercases and deduplicates words, dropping empty ones.
func normalizeWords(words []string) []string {
	return maps.Unique(trimWords(words))
}

// validateWords enforces the configured limits on the number and length of words.
func (h *handlers) validateWords(words ...string) error {
	limits := h.conf.Limits
	if limits.MaxWords > 0 && len(words) > limits.MaxWords {
		return ErrTooManyWords(limits.MaxWords)
	}

	if limits.MaxWordLength > 0 {
		for _, word := range words {
			if utf8.RuneCountInString(word) > limits.MaxWordLength {
				return ErrWordTooLong(limits.MaxWordLength)
			}
		}
	}

	return nil
}

// formError maps a ParseForm error to an HTTP error.
func formError(err error) error {
	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		return ErrBodyTooLarge
	}
	return BadRequest("Invalid request format")
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
