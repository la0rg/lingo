package apkg

import (
	"archive/zip"
	"bytes"
	"fmt"

	"github.com/la0rg/lingo/apkg/apkgdb"
)

const (
	databaseFilename = "collection.anki21"
	mediaFilename    = "media"
)

// Build builds an Anki apkg file from the given collection.
func Build(c Collection) ([]byte, error) {
	apkg := new(bytes.Buffer)
	archive := zip.NewWriter(apkg)

	db, err := buildDB(c)
	if err != nil {
		return nil, err
	}

	if err := c.writeFile(archive, databaseFilename, db); err != nil {
		return nil, err
	}

	if err := c.writeFile(archive, mediaFilename, []byte("{}")); err != nil {
		return nil, err
	}

	if err := archive.Close(); err != nil {
		return nil, fmt.Errorf("failed to close archive: %v", err)
	}

	return apkg.Bytes(), nil
}

func buildDB(c Collection) ([]byte, error) {
	db, err := apkgdb.NewDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to create apkgdb: %w", err)
	}

	err = db.AddCollection(c.buildCollection())
	if err != nil {
		return nil, fmt.Errorf("failed to add collection: %w", err)
	}

	for _, note := range c.buildNotes() {
		err = db.AddNote(note)
		if err != nil {
			return nil, fmt.Errorf("failed to add note: %w", err)
		}
	}
	for _, card := range c.buildCards() {
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
