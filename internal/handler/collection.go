package handler

import (
	"errors"
	"net/http"

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
		return requestError(err)
	}

	source, target := r.Form.Get("source"), r.Form.Get("target")
	if source == "" || target == "" {
		return ErrInvalidLanguage
	}

	collection := h.collection.GetBySlug(model.CollectionSlug(source, target))
	if collection == nil {
		return ErrCollectionNotFound
	}

	request := model.APKGRequest{Words: istrings.SplitAny(r.Form.Get("words"), "\n,")}
	request.Normalize()
	if err := request.Validate(h.conf.Limits.MaxWords, h.conf.Limits.MaxWordLength); err != nil {
		return requestError(err)
	}

	if !h.limiter.Allow(w, r, len(request.Words)) {
		return ErrRateLimited
	}

	file, err := h.apkg.CreateOnS3(r.Context(), collection, request.Words...)
	if err != nil {
		if errors.Is(err, service.ErrEmptyCards) {
			return ErrNoValidWords
		}

		return err
	}

	return h.HTML(w, r, web.CollectionDownload, map[string]any{
		"URL":            file.URL,
		"CardsCount":     file.CardsCount,
		"RequestedCount": len(request.Words),
		"MissingWords":   file.MissingWords,
	})
}
