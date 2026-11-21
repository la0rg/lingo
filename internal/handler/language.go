package handler

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/la0rg/lingo/internal/model"
	"github.com/la0rg/lingo/web"
)

const languageCookie = "language"

func (h *handlers) Language(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return formError(err)
	}

	source, target := r.Form.Get("source"), r.Form.Get("target")
	if source == "" || target == "" {
		return ErrInvalidLanguage
	}

	sourceNew, targetNew := r.Form.Get("sourceNew"), r.Form.Get("targetNew")
	languageSelector := model.UpdateLanguageSelector(
		source, target, sourceNew, targetNew,
	)

	http.SetCookie(w, &http.Cookie{
		Name:     languageCookie,
		Value:    fmt.Sprintf("%s,%s", languageSelector.SourceCountryCode, languageSelector.TargetCountryCode),
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
		HttpOnly: true,
	})

	if refererURL, err := url.Parse(r.Referer()); err == nil && refererURL.Path == "/" {
		// reload index page to get a new default word preview for the selected lang pair
		w.Header().Set("HX-Refresh", "true")
	}

	return h.HTML(w, r, web.LanguageSelector, languageSelector)
}
