package web

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type card struct {
	Word, Article, Gender, Transcription, Plural, Context, ContextTranslation, Definition string
	Translations                                                                          string
}

type httpError struct{ Message string }

func (httpError) Is4xx() bool { return true }

type language struct{ CountryCode, Flag, Name string }

type languageSelector struct {
	SourceCountryCode, TargetCountryCode string
	Languages                            []language
}

func TestTemplates_Render(t *testing.T) {
	t.Parallel()

	templates, err := NewTemplates()
	require.NoError(t, err)

	c := card{Word: "Apfel", Translations: "apple", Context: "Ein Apfel"}
	selector := languageSelector{"de", "en", []language{{"de", "de", "German"}, {"en", "us", "English"}}}

	cases := []struct {
		name TemplateName
		data any
		want string
	}{
		{Index, map[string]any{"LanguageSelector": selector, "Card": c, "HTMLTemplateName": CardPreviewDE}, "Apfel"},
		{Collection, map[string]any{"LanguageSelector": selector}, "Create a collection"},
		{CollectionDownload, map[string]any{"URL": "https://x/f.apkg", "CardsCount": 1, "RequestedCount": 2,
			"MissingWords": []string{"qwrt"}}, "https://x/f.apkg"},
		{CardPreview, c, "Apfel"},
		{CardPreviewDE, c, "Apfel"},
		{CardPreviewRU, c, "Apfel"},
		{LanguageSelector, selector, "German"},
		{Error, httpError{Message: "Oops"}, "Oops"},
	}

	for _, c := range cases {
		t.Run(string(c.name), func(t *testing.T) {
			t.Parallel()

			html, err := templates.Render(c.name, c.data)

			require.NoError(t, err)
			assert.Contains(t, string(html), c.want)
		})
	}
}

func TestTemplates_RenderUnknown(t *testing.T) {
	t.Parallel()

	templates, err := NewTemplates()
	require.NoError(t, err)

	_, err = templates.Render("missing.html", nil)
	assert.ErrorContains(t, err, "not found")
}

func TestEmbeddedAssets(t *testing.T) {
	t.Parallel()

	assert.NotEmpty(t, Styles)
	_, err := PublicFS.ReadFile("public/favicon.svg")
	assert.NoError(t, err)
}
