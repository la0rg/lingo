package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/la0rg/lingo/internal/logger"
	"github.com/la0rg/lingo/internal/model"
	"github.com/la0rg/lingo/pkg/apkg"
)

var apkgConfig = apkg.DefaultConfig(6493485093845, "lingo_config")

var ErrEmptyCards = fmt.Errorf("empty cards")

type APKG struct {
	generator CardGenerator
	cardRepo  CardRepository
	apkgRepo  APKGRepository
}

func NewAPKG(generator CardGenerator, apkgRepo APKGRepository, cardRepo CardRepository) *APKG {
	return &APKG{generator: generator, apkgRepo: apkgRepo, cardRepo: cardRepo}
}

func (s *APKG) CreateOnS3(ctx context.Context, collection *model.Collection, words ...string) (string, error) {
	logger := logger.WithContext(ctx).With("collection", collection.ID, "words", words)
	logger.Info("Creating an APKG file")

	var cards []model.Card
	var wordsToGenerate []string
	for _, word := range words {
		card, err := s.cardRepo.Get(ctx, collection.ID, word)
		if err != nil {
			return "", err
		}

		if card == nil {
			wordsToGenerate = append(wordsToGenerate, word)
			continue
		}

		cards = append(cards, *card)
	}

	if len(wordsToGenerate) > 0 {
		logger.Info("Generating new cards", "count", len(wordsToGenerate))

		newCards, err := s.generator.Generate(ctx, collection, wordsToGenerate...)
		if err != nil {
			return "", fmt.Errorf("unable to generate cards: %w", err)
		}

		cards = append(cards, newCards...)

		go func() {
			ctx := context.WithoutCancel(ctx)
			for _, card := range newCards {
				if err := s.cardRepo.Put(ctx, card); err != nil {
					logger.Error("Failed to store a card in DB", "error", err)
				}
			}
		}()
	}

	apkgCollection, err := s.buildAPKG(collection, cards)
	if err != nil {
		return "", fmt.Errorf("unable to create APKG collection: %w", err)
	}

	archive, err := apkg.Build(apkgCollection)
	if err != nil {
		return "", fmt.Errorf("unable to build APKG file: %w", err)
	}

	url, err := s.apkgRepo.StoreAndGetURL(ctx, s.s3Key(collection), archive)
	if err != nil {
		return "", fmt.Errorf("unable to save APKG file to S3: %w", err)
	}

	return url, nil
}

func (s *APKG) buildAPKG(collection *model.Collection, cards []model.Card) (*apkg.Collection, error) {
	if cards == nil || len(cards) == 0 {
		return nil, ErrEmptyCards
	}

	notes := make([]apkg.Note, 0)

	for _, card := range cards {
		notes = append(notes, apkg.Note{
			ID:    card.ID,
			Model: collection.Template,
			Fields: []string{
				card.Word,
				card.Article,
				card.Transcription,
				card.Plural,
				card.AllTranslations(),
				card.Context,
				card.ContextTranslation,
				card.Definition,
			},
		})
	}

	return &apkg.Collection{
		Decks: []apkg.Deck{{
			ID:     collection.ID,
			Name:   collection.Name,
			Config: apkgConfig,
			Notes:  notes,
		}},
	}, nil
}

func (s *APKG) s3Key(collection *model.Collection) string {
	return fmt.Sprintf(
		"apkg/%d/lingo_%d_%s.apkg",
		collection.ID, time.Now().Unix(), uuid.New().String(),
	)
}
