package service

import (
	"context"
	"fmt"

	"github.com/la0rg/lingo/logger"
	"github.com/la0rg/lingo/model"
)

type Card struct {
	generator CardGenerator
	repo      CardRepository
}

func NewCard(generator CardGenerator, repo CardRepository) *Card {
	return &Card{generator: generator, repo: repo}
}

func (s *Card) Get(ctx context.Context, collection *model.Collection, word string) (*model.Card, error) {
	logger := logger.WithContext(ctx).With("collection", collection.ID, "card", word)
	logger.Info("Creating a card")

	result, err := s.generator.Generate(ctx, collection, word)
	if err != nil {
		return nil, fmt.Errorf("unable to generate a card: %v", err)
	}

	if len(result.Cards) < 1 {
		return nil, fmt.Errorf("no cards generated")
	}

	if len(result.Cards) > 1 {
		logger.Warn("more than one card generated", "cards", result.Cards)
	}

	card := result.Cards[0]

	go func() {
		ctx := context.WithoutCancel(ctx)
		if err := s.repo.Put(ctx, card); err != nil {
			logger.Error("Failed to store a card in DB", "error", err)
		}
	}()

	return &card, nil
}
