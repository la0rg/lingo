package service

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/la0rg/lingo/internal/model"
	"github.com/la0rg/lingo/internal/service/mocks"
)

var testCollection = &model.Collection{ID: 42, SourceLanguage: "de", TargetLanguage: "en"}

func TestCard_Get_FromRepository(t *testing.T) {
	t.Parallel()

	repo := mocks.NewCardRepository(t)
	generator := mocks.NewCardGenerator(t) // no expectations: must not be called
	stored := &model.Card{ID: 1, Word: "Apfel", CollectionID: 42}
	repo.EXPECT().Get(mock.Anything, int64(42), "apfel").Return(stored, nil)

	card, err := NewCard(generator, repo).Get(context.Background(), testCollection, "apfel")

	require.NoError(t, err)
	assert.Same(t, stored, card)
}

func TestCard_Get_GeneratesAndStores(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		repo := mocks.NewCardRepository(t)
		generator := mocks.NewCardGenerator(t)

		repo.EXPECT().Get(mock.Anything, int64(42), "apfel").Return(nil, nil)
		generator.EXPECT().Generate(mock.Anything, testCollection, "apfel").
			Return(&model.Card{Word: "Apfel"}, nil)

		var put model.Card
		repo.EXPECT().Put(mock.Anything, mock.Anything).
			Run(func(_ context.Context, card model.Card) { put = card }).
			Return(nil)

		ctx, cancel := context.WithCancel(context.Background())
		card, err := NewCard(generator, repo).Get(ctx, testCollection, "apfel")
		cancel() // the card must be stored even after the request is done

		synctest.Wait()

		require.NoError(t, err)
		assert.Equal(t, "Apfel", card.Word)
		assert.NotZero(t, card.ID)
		assert.Equal(t, int64(42), card.CollectionID)
		assert.Equal(t, *card, put)
	})
}

func TestCard_Get_StoreFailureIsNotAnError(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		repo := mocks.NewCardRepository(t)
		generator := mocks.NewCardGenerator(t)

		repo.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
		generator.EXPECT().Generate(mock.Anything, mock.Anything, mock.Anything).Return(&model.Card{Word: "x"}, nil)
		repo.EXPECT().Put(mock.Anything, mock.Anything).Return(errors.New("db down"))

		card, err := NewCard(generator, repo).Get(context.Background(), testCollection, "x")
		synctest.Wait()

		require.NoError(t, err)
		assert.NotNil(t, card)
	})
}

func TestCard_Get_InvalidWord(t *testing.T) {
	t.Parallel()

	repo := mocks.NewCardRepository(t)
	generator := mocks.NewCardGenerator(t)
	repo.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
	generator.EXPECT().Generate(mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

	card, err := NewCard(generator, repo).Get(context.Background(), testCollection, "qwrt")

	require.NoError(t, err)
	assert.Nil(t, card, "nothing is stored for invalid words")
}

func TestCard_Get_Errors(t *testing.T) {
	t.Parallel()

	t.Run("repository", func(t *testing.T) {
		t.Parallel()
		repo := mocks.NewCardRepository(t)
		repoErr := errors.New("db down")
		repo.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(nil, repoErr)

		_, err := NewCard(mocks.NewCardGenerator(t), repo).Get(context.Background(), testCollection, "x")

		assert.ErrorIs(t, err, repoErr)
	})

	t.Run("generator", func(t *testing.T) {
		t.Parallel()
		repo := mocks.NewCardRepository(t)
		generator := mocks.NewCardGenerator(t)
		repo.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
		generator.EXPECT().Generate(mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("openai down"))

		_, err := NewCard(generator, repo).Get(context.Background(), testCollection, "x")

		assert.ErrorContains(t, err, "openai down")
	})
}

func TestCard_GetAll(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		repo := mocks.NewCardRepository(t)
		generator := mocks.NewCardGenerator(t)

		repo.EXPECT().Get(mock.Anything, int64(42), "apfel").Return(&model.Card{Word: "Apfel"}, nil)
		repo.EXPECT().Get(mock.Anything, int64(42), "haus").Return(nil, nil)
		repo.EXPECT().Get(mock.Anything, int64(42), "qwrt").Return(nil, nil)
		generator.EXPECT().Generate(mock.Anything, mock.Anything, "haus").Return(&model.Card{Word: "Haus"}, nil)
		generator.EXPECT().Generate(mock.Anything, mock.Anything, "qwrt").Return(nil, nil)
		repo.EXPECT().Put(mock.Anything, mock.Anything).Return(nil).Once()

		result, err := NewCard(generator, repo).GetAll(context.Background(), testCollection, "apfel", "haus", "qwrt")
		synctest.Wait()

		require.NoError(t, err)
		words := make([]string, 0, len(result.Cards))
		for _, card := range result.Cards {
			words = append(words, card.Word)
		}
		assert.ElementsMatch(t, []string{"Apfel", "Haus"}, words)
		assert.Equal(t, []string{"qwrt"}, result.MissingWords)
	})
}

func TestCard_GetAll_Error(t *testing.T) {
	t.Parallel()

	repo := mocks.NewCardRepository(t)
	repoErr := errors.New("db down")
	repo.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(nil, repoErr)

	_, err := NewCard(mocks.NewCardGenerator(t), repo).GetAll(context.Background(), testCollection, "a", "b")

	assert.ErrorIs(t, err, repoErr)
}
