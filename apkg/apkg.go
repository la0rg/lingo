package apkg

import (
	"archive/zip"
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/la0rg/lingo/apkg/apkgdb"
)

const (
	databaseFilename = "collection.anki21"
	mediaFilename    = "media"
)

type Collection struct {
	Decks []Deck
}

type Deck struct {
	ID     int64
	Name   string
	Config Config
	Notes  []Note
}

type Config apkgdb.DeckConfig

type Note struct {
	ID     int64
	Model  *Model
	Fields []string
	Cards  []Card
}

type Card struct {
	ID int64
}

func (n *Note) Guid() string {
	var key string
	if len(n.Fields) > 0 {
		key = n.Fields[0]
	}

	hash := sha1.Sum([]byte(n.Model.Name + key))
	return hex.EncodeToString(hash[:])
}

type Model struct {
	ID        int64
	Name      string
	Css       string
	Templates []Template
	Fields    []string
}

func (m Model) buildTemplates() []apkgdb.Template {
	var templates = make([]apkgdb.Template, 0, len(m.Templates))

	for _, template := range m.Templates {
		templates = append(templates, apkgdb.Template{
			Name: template.Name,
			Ord:  len(templates),
			Qfmt: apkgdb.UnescapedString(template.Qfmt),
			Afmt: apkgdb.UnescapedString(template.Afmt),
		})
	}
	return templates
}

func (m Model) buildFields() []apkgdb.Field {
	var fields = make([]apkgdb.Field, 0, len(m.Fields))

	for _, name := range m.Fields {
		fields = append(fields, apkgdb.Field{
			Name: name,
			Ord:  len(fields),
			Font: "Arial",
			Size: 20,
		})
	}
	return fields
}

type Template struct {
	Name string
	Qfmt string
	Afmt string
}

func (a Collection) buildNotes() []apkgdb.Note {
	var notes = make([]apkgdb.Note, 0)

	for _, deck := range a.Decks {
		for _, note := range deck.Notes {
			notes = append(notes, apkgdb.Note{
				ID:     note.ID,
				Guid:   note.Guid(),
				Mid:    note.Model.ID,
				Mod:    time.Now().Unix(),
				Usn:    1,
				Fields: note.Fields,
			})
		}
	}
	return notes
}

func (a Collection) buildCards() []apkgdb.Card {
	var cards = make([]apkgdb.Card, 0)

	for _, deck := range a.Decks {
		for _, note := range deck.Notes {
			for ord, card := range note.Cards {
				if ord >= len(note.Model.Templates) {
					break
				}

				cards = append(cards, apkgdb.Card{
					ID:   card.ID,
					Nid:  note.ID,
					Did:  deck.ID,
					Ord:  ord,
					Mod:  time.Now().Unix(),
					Usn:  1,
					Due:  note.ID,
					Left: 1001, //  e.g.: '2004' means 2 reps left today and 4 reps till graduation
				})
			}
		}
	}

	return cards
}

func (a Collection) Build() ([]byte, error) {
	apkg := new(bytes.Buffer)

	archive := zip.NewWriter(apkg)

	db, err := a.buildDB()
	if err != nil {
		return nil, err
	}

	w, err := archive.Create(databaseFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s: %w", databaseFilename, err)
	}

	if _, err := w.Write(db); err != nil {
		return nil, fmt.Errorf("failed to write %s: %w", databaseFilename, err)
	}

	w, err = archive.Create(mediaFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s: %w", mediaFilename, err)
	}

	if _, err := w.Write([]byte("{}")); err != nil {
		return nil, fmt.Errorf("failed to write %s: %w", mediaFilename, err)
	}

	if err := archive.Close(); err != nil {
		return nil, fmt.Errorf("failed to close archive: %w", err)
	}

	return apkg.Bytes(), nil
}

func (a Collection) buildDB() ([]byte, error) {
	db, err := apkgdb.NewDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to create apkgdb: %w", err)
	}

	err = db.AddCollection(a.buildCollection())
	if err != nil {
		return nil, fmt.Errorf("failed to add collection: %w", err)
	}

	for _, note := range a.buildNotes() {
		err = db.AddNote(note)
		if err != nil {
			return nil, fmt.Errorf("failed to add note: %w", err)
		}
	}

	for _, card := range a.buildCards() {
		err = db.AddCard(card)
		if err != nil {
			return nil, fmt.Errorf("failed to add card: %w", err)
		}
	}

	var buf bytes.Buffer
	if err := db.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write apkgdb: %w", err)
	}

	if err := db.Close(); err != nil {
		return nil, fmt.Errorf("failed to close apkgdb: %w", err)
	}

	return buf.Bytes(), nil
}

func (a Collection) buildCollection() apkgdb.Collection {
	now := time.Now()

	// models
	var models = make(map[int64]apkgdb.Model)
	var decks = make(map[int64]apkgdb.Deck)
	var configs = make(map[int64]apkgdb.DeckConfig)

	for _, deck := range a.Decks {
		configs[deck.Config.ID] = apkgdb.DeckConfig(deck.Config)

		decks[deck.ID] = apkgdb.Deck{
			ID:               deck.ID,
			Name:             deck.Name,
			Mod:              time.Now().Unix(),
			Usn:              1,
			Collapsed:        true,
			BrowserCollapsed: true,
			Conf:             deck.Config.ID,
		}

		for _, note := range deck.Notes {
			models[note.Model.ID] = apkgdb.Model{
				ID:        note.Model.ID,
				Name:      note.Model.Name,
				Mod:       time.Now().Unix(),
				Usn:       1,
				Sortf:     0,
				Did:       deck.ID,
				Css:       apkgdb.UnescapedString(note.Model.Css),
				Templates: note.Model.buildTemplates(),
				Fields:    note.Model.buildFields(),
				Tags:      []string{},
			}
		}
	}

	return apkgdb.Collection{
		Crt:    now,
		Mod:    now,
		Scm:    now,
		Models: mapIntToStr(models),
		Decks:  mapIntToStr(decks),
		DConf:  mapIntToStr(configs),
	}
}

func DefaultConfig(id int64, name string) Config {
	return Config{
		ID:       id,
		Mod:      time.Now().Unix(),
		Name:     name,
		Usn:      1,
		MaxTaken: 60,
		New: apkgdb.NewCardConfig{
			Delays:        []float64{1, 10},
			InitialFactor: 2500,
			Ints:          []int{1, 4, 0},
			PerDay:        20,
		},
		Rev: apkgdb.ReviewCardConfig{
			Ease4:  1.3,
			IvlFct: 1.0,
			MaxIvl: 36500,
			PerDay: 100,
		},
		Lapse: apkgdb.LapseCardConfig{
			Delays:      []float64{10},
			LeechAction: 1,
			LeechFails:  10,
			MinInt:      1,
		},
	}
}

func mapIntToStr[T any](m map[int64]T) map[string]T {
	result := make(map[string]T, len(m))
	for id, v := range m {
		result[strconv.FormatInt(id, 10)] = v
	}
	return result
}
