package repository

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/la0rg/lingo/internal/model"
)

// Collection and model IDs are baked into the Anki decks users have already imported.
// Changing them breaks updating existing decks, so they are pinned here.
var pinnedIDs = map[string]struct{ collection, model int64 }{
	"de-en": {1715808757, 1715808522},
	"de-ru": {1715808758, 1715808523},
	"ru-en": {1715808759, 1715808524},
	"ru-de": {1715808760, 1715808525},
	"en-de": {1715808761, 1715808526},
	"en-ru": {1715808762, 1715808527},
}

func TestCollection_PinnedIDs(t *testing.T) {
	t.Parallel()

	c, err := NewCollection()
	require.NoError(t, err)

	assert.Len(t, c.m, len(pinnedIDs), "update pinnedIDs when adding a collection")
	for slug, ids := range pinnedIDs {
		collection := c.GetBySlug(slug)
		require.NotNil(t, collection, slug)
		assert.Equal(t, ids.collection, collection.ID, slug)
		assert.Equal(t, ids.model, collection.Template.ID, slug)
	}
}

func TestCollection_LoadsData(t *testing.T) {
	t.Parallel()

	c, err := NewCollection()
	require.NoError(t, err)

	for slug, collection := range c.m {
		assert.Equal(t, slug, model.CollectionSlug(collection.SourceLanguage, collection.TargetLanguage))
		assert.NotEmpty(t, collection.DefaultWord, slug)
		assert.NotEmpty(t, collection.Template.Fields, slug)
		assert.Contains(t, collection.Template.Css, "{", "%s: css is loaded", slug)

		require.Len(t, collection.Template.Templates, 2, slug)
		for _, tmpl := range collection.Template.Templates {
			assert.False(t, strings.HasPrefix(tmpl.Qfmt, "data/"), "%s: qfmt is loaded", slug)
			assert.False(t, strings.HasPrefix(tmpl.Afmt, "data/"), "%s: afmt is loaded", slug)
			assert.Contains(t, tmpl.Qfmt, "{{", slug)
			assert.Contains(t, tmpl.Afmt, "{{", slug)
		}
	}

	assert.Nil(t, c.GetBySlug("xx-yy"))
}

func TestCollection_CanBeLoadedTwice(t *testing.T) {
	t.Parallel()

	first, err := NewCollection()
	require.NoError(t, err)
	second, err := NewCollection()
	require.NoError(t, err)

	assert.Equal(t, first.GetBySlug("de-en").Template.Css, second.GetBySlug("de-en").Template.Css)
	assert.NotSame(t, first.GetBySlug("de-en"), second.GetBySlug("de-en"))
}
