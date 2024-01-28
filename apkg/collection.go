package apkg

import (
	"archive/zip"
	"fmt"
	"time"

	"github.com/la0rg/lingo/apkg/apkgdb"
	"github.com/la0rg/lingo/maps"
)

type Collection struct {
	Decks []Deck
}

func (c Collection) buildNotes() []apkgdb.Note {
	var notes = make([]apkgdb.Note, 0)

	for _, deck := range c.Decks {
		for _, note := range deck.Notes {
			notes = append(notes, note.buildNote())
		}
	}
	return notes
}

func (c Collection) buildCards() []apkgdb.Card {
	var cards = make([]apkgdb.Card, 0)

	for _, deck := range c.Decks {
		for _, note := range deck.Notes {
			cards = append(cards, note.buildCards(deck.ID)...)
		}
	}

	return cards
}

func (c Collection) buildCollection() apkgdb.Collection {
	now := time.Now()

	var models = make(map[int64]apkgdb.Model)
	var decks = make(map[int64]apkgdb.Deck)
	var configs = make(map[int64]apkgdb.DeckConfig)

	for _, deck := range c.Decks {
		configs[deck.Config.ID] = apkgdb.DeckConfig(deck.Config)
		decks[deck.ID] = deck.buildDeck()

		for _, note := range deck.Notes {
			if _, ok := models[note.Model.ID]; !ok {
				models[note.Model.ID] = note.Model.buildModel(deck.ID)
			}
		}
	}

	return apkgdb.Collection{
		Crt:    now,
		Mod:    now,
		Scm:    now,
		Models: maps.KeysToString(models),
		Decks:  maps.KeysToString(decks),
		DConf:  maps.KeysToString(configs),
	}
}

func (c Collection) writeFile(archive *zip.Writer, filename string, data []byte) error {
	w, err := archive.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create %s: %v", filename, err)
	}

	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("failed to write %s: %v", filename, err)
	}

	return nil
}
