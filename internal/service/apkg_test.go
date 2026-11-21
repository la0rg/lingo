package service

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/la0rg/lingo/internal/model"
	"github.com/la0rg/lingo/internal/repository"
	"github.com/la0rg/lingo/internal/service/mocks"
)

func deEnCollection(t *testing.T) *model.Collection {
	t.Helper()
	collections, err := repository.NewCollection()
	require.NoError(t, err)
	c := collections.GetBySlug("de-en")
	require.NotNil(t, c)
	return c
}

func TestAPKG_CreateOnS3(t *testing.T) {
	t.Parallel()

	collection := deEnCollection(t)
	cards := mocks.NewCardGetter(t)
	repo := mocks.NewAPKGRepository(t)

	cards.EXPECT().GetAll(mock.Anything, collection, []string{"apfel", "qwrt"}).Return(&model.GetAllResult{
		Cards:        []model.Card{{ID: 1, Word: "Apfel", Article: "der"}},
		MissingWords: []string{"qwrt"},
	}, nil)

	var key string
	var archive []byte
	repo.EXPECT().StoreAndGetURL(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, k string, a []byte) (string, error) {
			key, archive = k, a
			return "https://s3.example/file.apkg", nil
		})

	file, err := NewAPKG(cards, repo).CreateOnS3(context.Background(), collection, "apfel", "qwrt")

	require.NoError(t, err)
	assert.Equal(t, &model.CollectionFile{
		URL:            "https://s3.example/file.apkg",
		CardsCount:     1,
		CollectionName: "Lingo DE-EN",
		MissingWords:   []string{"qwrt"},
	}, file)

	assert.Regexp(t, regexp.MustCompile(`^apkg/1715808757/lingo_\d+_[0-9a-f-]{36}\.apkg$`), key)

	zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	require.NoError(t, err)
	var names []string
	for _, f := range zr.File {
		names = append(names, f.Name)
	}
	assert.ElementsMatch(t, []string{"collection.anki21", "media"}, names)
}

func TestAPKG_CreateOnS3_Errors(t *testing.T) {
	t.Parallel()

	collection := deEnCollection(t)

	t.Run("no cards", func(t *testing.T) {
		t.Parallel()
		cards := mocks.NewCardGetter(t)
		cards.EXPECT().GetAll(mock.Anything, mock.Anything, mock.Anything).
			Return(&model.GetAllResult{MissingWords: []string{"qwrt"}}, nil)

		_, err := NewAPKG(cards, mocks.NewAPKGRepository(t)).CreateOnS3(context.Background(), collection, "qwrt")

		assert.ErrorIs(t, err, ErrEmptyCards)
	})

	t.Run("cards", func(t *testing.T) {
		t.Parallel()
		cards := mocks.NewCardGetter(t)
		cardsErr := errors.New("db down")
		cards.EXPECT().GetAll(mock.Anything, mock.Anything, mock.Anything).Return(nil, cardsErr)

		_, err := NewAPKG(cards, mocks.NewAPKGRepository(t)).CreateOnS3(context.Background(), collection, "a")

		assert.ErrorIs(t, err, cardsErr)
	})

	t.Run("storage", func(t *testing.T) {
		t.Parallel()
		cards := mocks.NewCardGetter(t)
		repo := mocks.NewAPKGRepository(t)
		storageErr := errors.New("s3 down")
		cards.EXPECT().GetAll(mock.Anything, mock.Anything, mock.Anything).
			Return(&model.GetAllResult{Cards: []model.Card{{ID: 1, Word: "a"}}}, nil)
		repo.EXPECT().StoreAndGetURL(mock.Anything, mock.Anything, mock.Anything).Return("", storageErr)

		_, err := NewAPKG(cards, repo).CreateOnS3(context.Background(), collection, "a")

		assert.ErrorIs(t, err, storageErr)
	})
}

func TestAPKG_BuildEscapesFields(t *testing.T) {
	t.Parallel()

	collection := deEnCollection(t)
	card := model.Card{
		ID:           7,
		Word:         `<img src=x onerror="alert(1)">`,
		Article:      `der" onmouseover="x`,
		Translations: model.Translations{{Translations: []string{"<b>apple</b>", "Tom & Jerry"}}},
	}

	apkgCollection, err := NewAPKG(nil, nil).buildAPKG(collection, []model.Card{card})
	require.NoError(t, err)

	require.Len(t, apkgCollection.Decks, 1)
	deck := apkgCollection.Decks[0]
	assert.Equal(t, collection.ID, deck.ID)
	assert.Equal(t, collection.Name, deck.Name)
	require.Len(t, deck.Notes, 1)

	note := deck.Notes[0]
	assert.Equal(t, int64(7), note.ID)
	require.Len(t, note.Fields, len(collection.Template.Fields))
	assert.Equal(t, "&lt;img src=x onerror=&#34;alert(1)&#34;&gt;", note.Fields[0])
	assert.Equal(t, "der&#34; onmouseover=&#34;x", note.Fields[1])
	assert.Equal(t, "&lt;b&gt;apple&lt;/b&gt;, Tom &amp; Jerry", note.Fields[4])
}
