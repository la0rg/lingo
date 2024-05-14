package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/la0rg/lingo/apkg"
	"github.com/la0rg/lingo/logger"
	"github.com/la0rg/lingo/model"
)

var de_en_model = apkg.Model{
	ID:   555555,
	Name: "lingo_de_en",
	Css:  `.word { \n\tfont-size: 28px; \nfont-style: italic; \n}`,
	Templates: []apkg.Template{
		{
			Name: "lingo",
			Qfmt: `<div class=word>{{word}}</div>`,
			Afmt: `{{FrontSide}}\n\n<div class=word>{{translation}}</div>`,
		},
		{
			Name: "lingo (reversed)",
			Qfmt: `<div class=word>{{translation}}</div>`,
			Afmt: `{{FrontSide}}\n\n<div class=word>{{word}}</div>`,
		},
	},
	Fields: []string{"word", "translation"},
}

var apkgConfig = apkg.DefaultConfig(6493485093845, "lingo_config")

var ErrEmptyCards = fmt.Errorf("empty cards")

type APKG struct {
	generator CardGenerator
	repo      APKGRepository
}

func NewAPKG(generator CardGenerator, repo APKGRepository) *APKG {
	return &APKG{generator: generator, repo: repo}
}

func (s *APKG) CreateOnS3(ctx context.Context, collection *model.Collection, words ...string) (string, error) {
	logger := logger.WithContext(ctx).With("collection", collection.ID, "words", words)
	logger.Info("Creating an APKG file")

	cards, err := s.generator.Generate(ctx, collection, words...)
	if err != nil {
		return "", fmt.Errorf("unable to generate cards: %w", err)
	}

	apkgCollection, err := s.buildAPKG(cards)
	if err != nil {
		return "", fmt.Errorf("unable to create APKG collection: %w", err)
	}

	archive, err := apkg.Build(apkgCollection)
	if err != nil {
		return "", fmt.Errorf("unable to build APKG file: %w", err)
	}

	url, err := s.repo.StoreAndGetURL(ctx, s.s3Key(collection), archive)
	if err != nil {
		return "", fmt.Errorf("unable to save APKG file to S3: %w", err)
	}

	return url, nil
}

func (s *APKG) buildAPKG(cards []model.Card) (*apkg.Collection, error) {
	if cards == nil || len(cards) == 0 {
		return nil, ErrEmptyCards
	}

	notes := make([]apkg.Note, 0)
	id := time.Now().UnixMilli()

	for _, card := range cards {
		translations := make([]string, 0)
		for _, translation := range card.Translations {
			translations = append(translations, translation.Translations...)
		}

		notes = append(notes, apkg.Note{
			ID:     id, // TODO: save ID in DB
			Model:  de_en_model,
			Fields: []string{card.Word, strings.Join(translations, ", ")},
		})

		id++
	}

	return &apkg.Collection{
		Decks: []apkg.Deck{{
			ID:     58098085085,
			Name:   "lingo_de_en",
			Config: apkgConfig,
			Notes:  notes,
		}},
	}, nil
}

func (s *APKG) s3Key(collection *model.Collection) string {
	return fmt.Sprintf("apkg/%s/lingo_%s.apkg", uuid.New().String(), collection.ID)
}
