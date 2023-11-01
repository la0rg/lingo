package apkgdb

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"strings"

	_ "modernc.org/sqlite"
)

const (
	version         = 11
	fieldsDelimiter = "\x1f"
)

// Database is a helper for creating a new Anki apkgdb in v11 format.
type Database struct {
	db      *sql.DB
	tmpFile *os.File
}

func NewDatabase() (*Database, error) {
	tmpFile, err := os.CreateTemp("", "collection.anki21")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	db, err := sql.Open("sqlite", tmpFile.Name())
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(v11)
	if err != nil {
		return nil, err
	}

	return &Database{db: db, tmpFile: tmpFile}, nil
}

// Write writes the apkgdb to the given writer.
func (d *Database) Write(w io.Writer) error {
	file, err := os.Open(d.tmpFile.Name())
	if err != nil {
		return fmt.Errorf("unable to open temp file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(w, file)
	if err != nil {
		return fmt.Errorf("unable to copy temp file to writer: %w", err)
	}

	return err
}

// Close frees all resources associated with the apkgdb.
// The apkgdb is no longer usable after this call.
func (d *Database) Close() error {
	err := d.db.Close()
	if err != nil {
		return fmt.Errorf("failed to close apkgdb: %w", err)
	}

	if err := d.tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	if err := os.Remove(d.tmpFile.Name()); err != nil {
		return fmt.Errorf("failed to remove temp file: %w", err)
	}

	return nil
}

func (d *Database) AddCollection(c Collection) error {
	_, err := d.db.Exec(
		`INSERT INTO col (crt, mod, scm, ver, dty, usn, ls, conf, models, decks, dconf, tags)
		VALUES (?, ?, ?, ?, 0, 0, 0, '{}', ?, ?, ?, '{}')`,
		c.Crt.Unix(), c.Mod.UnixMilli(), c.Scm.UnixMilli(), version, c.Models, c.Decks, c.DConf,
	)
	return err
}

func (d *Database) AddNote(n Note) error {
	_, err := d.db.Exec(
		`INSERT INTO notes (id, guid, mid, mod, usn, tags, flds, sfld, csum, flags, data)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0, '')`,
		n.ID,
		n.Guid,
		n.Mid,
		n.Mod,
		n.Usn,
		n.Tags,
		strings.Join(n.Fields, fieldsDelimiter),
		n.SField(),
		n.Csum(),
	)
	return err
}

func (d *Database) AddCard(c Card) error {
	_, err := d.db.Exec(
		`INSERT INTO cards (
        	id, nid, did, ord, mod, usn, type, queue, due, ivl, factor, reps, lapses, "left", odue, odid, flags, data
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, '{}')`,
		c.ID, c.Nid, c.Did, c.Ord, c.Mod, c.Usn, c.Type, c.Queue, c.Due, c.Ivl, c.Factor, c.Reps, c.Lapses, c.Left,
		c.ODue, c.ODid, c.Flags,
	)
	return err
}
