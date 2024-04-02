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

func (s *Card) Get(ctx context.Context, collection *model.Collection, word model.Word) (*model.Card, error) {
	logger := logger.WithContext(ctx).With("collection", collection.ID, "card", word)
	logger.Info("Getting a card")

	card, err := s.repo.Get(ctx, collection.ID, word)
	if err != nil {
		return nil, err
	}

	if card != nil {
		return card, nil
	}

	logger.Info("Card is not found. Generatinng a new one.")

	cards, err := s.generator.Generate(ctx, collection, word)
	if err != nil {
		return nil, fmt.Errorf("unable to generate a card: %v", err)
	}

	if len(cards) < 1 {
		return nil, fmt.Errorf("no cards generated")
	}

	if len(cards) > 1 {
		logger.Warn("more than one card generated", "cards", cards)
	}

	go func() {
		ctx := context.WithoutCancel(ctx)
		if err := s.repo.Put(ctx, cards[0]); err != nil {
			logger.Error("Failed to store a card in DB", "error", err)
		}
	}()

	return &cards[0], nil
}
