package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/la0rg/lingo/apkg"
	"github.com/la0rg/lingo/internal/logger"
	"github.com/la0rg/lingo/internal/model"
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

func (s *APKG) CreateOnS3(ctx context.Context, collection *model.Collection, words ...string) (*model.CollectionFile, error) {
	logger := logger.WithContext(ctx).With("collection", collection.ID, "words", words)
	logger.Info("Creating an APKG file")

	var cards []model.Card
	var wordsToGenerate []string
	for _, word := range words {
		card, err := s.cardRepo.Get(ctx, collection.ID, word)
		if err != nil {
			return nil, err
		}

		if card == nil {
			wordsToGenerate = append(wordsToGenerate, word)
			continue
		}

		cards = append(cards, *card)
	}

	newCards, err := s.generate(ctx, logger, collection, wordsToGenerate...)
	if err != nil {
		return nil, err
	}
	cards = append(cards, newCards...)

	apkgCollection, err := s.buildAPKG(collection, cards)
	if err != nil {
		return nil, fmt.Errorf("unable to create APKG collection: %w", err)
	}

	archive, err := apkg.Build(apkgCollection)
	if err != nil {
		return nil, fmt.Errorf("unable to build APKG file: %w", err)
	}

	url, err := s.apkgRepo.StoreAndGetURL(ctx, s.s3Key(collection), archive)
	if err != nil {
		return nil, fmt.Errorf("unable to save APKG file to S3: %w", err)
	}

	return &model.CollectionFile{
		URL:            url,
		CardsCount:     len(cards),
		CollectionName: collection.Name,
		MissingWords:   s.missingWords(newCards, wordsToGenerate),
	}, nil
}

func (s *APKG) generate(ctx context.Context, logger *slog.Logger, collection *model.Collection, words ...string) ([]model.Card, error) {
	if len(words) == 0 {
		return nil, nil
	}

	logger.Info("Generating new cards", "new_words", words)

	newCards, err := s.generator.Generate(ctx, collection, words...)
	if err != nil {
		return nil, fmt.Errorf("unable to generate cards: %w", err)
	}

	go func() {
		ctx := context.WithoutCancel(ctx)
		for _, card := range newCards {
			if err := s.cardRepo.Put(ctx, card); err != nil {
				logger.Error("Failed to store a card in DB", "error", err)
			}
		}
	}()

	return newCards, nil
}

func (s *APKG) missingWords(cards []model.Card, words []string) []string {
	var cardsMap = make(map[string]struct{}, len(cards))
	for _, card := range cards {
		cardsMap[strings.ToLower(card.Word)] = struct{}{}
	}

	var missingWords []string
	for _, w := range words {
		if _, ok := cardsMap[strings.ToLower(w)]; !ok {
			missingWords = append(missingWords, w)
		}
	}

	return missingWords
}

func (s *APKG) buildAPKG(collection *model.Collection, cards []model.Card) (*apkg.Collection, error) {
	if len(cards) == 0 {
		return nil, ErrEmptyCards
	}

	notes := make([]apkg.Note, 0)

	for _, card := range cards {
		note := apkg.Note{ID: card.ID, Model: collection.Template}
		for _, field := range collection.Template.Fields {
			note.Fields = append(note.Fields, card.FieldByName(field))
		}
		notes = append(notes, note)
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
