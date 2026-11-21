package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateLanguageSelector(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                   string
		source, target         string
		sourceNew, targetNew   string
		wantSource, wantTarget string
	}{
		{"no change", "de", "en", "", "", "de", "en"},
		{"new source", "de", "en", "ru", "", "ru", "en"},
		{"new target", "de", "en", "", "ru", "de", "ru"},
		{"source equals target swaps", "de", "en", "en", "", "en", "de"},
		{"target equals source swaps", "de", "en", "", "de", "en", "de"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			ls := UpdateLanguageSelector(c.source, c.target, c.sourceNew, c.targetNew)

			assert.Equal(t, c.wantSource, ls.SourceCountryCode)
			assert.Equal(t, c.wantTarget, ls.TargetCountryCode)
			assert.Equal(t, languages, ls.Languages)
		})
	}
}

func TestLanguageSelector_CollectionSlug(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "de-en", NewLanguageSelector("de", "en").CollectionSlug())
	assert.Equal(t, "ru-de", CollectionSlug("ru", "de"))
}
