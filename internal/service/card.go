package service

import (
	"context"
	"fmt"

	"github.com/sourcegraph/conc/pool"

	"github.com/la0rg/lingo/apkg"
	"github.com/la0rg/lingo/internal/logger"
	"github.com/la0rg/lingo/internal/model"
)

const maxWorkers = 10

type Card struct {
	generator CardGenerator
	repo      CardRepository
}

func NewCard(generator CardGenerator, repo CardRepository) *Card {
	return &Card{generator: generator, repo: repo}
}

func (s *Card) Get(ctx context.Context, collection *model.Collection, word string) (*model.Card, error) {
	logger := logger.WithContext(ctx).With("collection", collection.ID, "word", word)
	logger.Info("Getting a card")

	card, err := s.repo.Get(ctx, collection.ID, word)
	if err != nil {
		return nil, err
	}

	if card != nil {
		logger.Info("Card is found in DB.")
		return card, nil
	}

	logger.Info("Card is not found. Generating a new one.")

	card, err = s.generator.Generate(ctx, collection, word)
	if err != nil {
		return nil, fmt.Errorf("unable to generate a card: %v", err)
	}

	if card == nil {
		logger.Info("card is not created")
		return nil, nil
	}

	// fill card's metadata
	card.ID = apkg.NewID()
	card.CollectionID = collection.ID

	go func() {
		ctx := context.WithoutCancel(ctx)
		if err := s.repo.Put(ctx, *card); err != nil {
			logger.Error("Failed to store a card in DB", "error", err)
		}
	}()

	return card, nil
}

type getResult struct {
	word string
	card *model.Card
}

func (s *Card) GetAll(
	ctx context.Context, collection *model.Collection, words ...string,
) (*model.GetAllResult, error) {
	pool := pool.NewWithResults[getResult]().
		WithContext(ctx).
		WithCancelOnError().
		WithFirstError().
		WithMaxGoroutines(maxWorkers)

	for _, word := range words {
		pool.Go(func(ctx context.Context) (getResult, error) {
			card, err := s.Get(ctx, collection, word)
			return getResult{word: word, card: card}, err
		})
	}

	results, err := pool.Wait()
	if err != nil {
		return nil, err
	}

	var result model.GetAllResult
	for _, res := range results {
		if res.card != nil {
			result.Cards = append(result.Cards, *res.card)
			continue
		}

		result.MissingWords = append(result.MissingWords, res.word)
	}

	return &result, nil
}
