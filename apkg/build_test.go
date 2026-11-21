package apkg

import (
	"archive/zip"
	"bytes"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testCollection() *Collection {
	model := Model{
		ID:   555555,
		Name: "lingo_test",
		Css:  ".word { font-size: 28px; }",
		Templates: []Template{
			{Name: "lingo", Qfmt: "<div class=word>{{word}}</div>", Afmt: "{{FrontSide}}<div>{{translation}}</div>"},
			{Name: "lingo (reversed)", Qfmt: "<div>{{translation}}</div>", Afmt: "{{FrontSide}}<div>{{word}}</div>"},
		},
		Fields: []string{"word", "translation"},
	}

	return &Collection{Decks: []Deck{{
		ID:     58098085085,
		Name:   "lingo_test",
		Config: DefaultConfig(6493485093845, "lingo_test"),
		Notes: []Note{
			{ID: 1111111, Model: model, Fields: []string{"Welt", "world"}},
			{ID: 1111112, Model: model, Fields: []string{"rot", "red"}},
		},
	}}}
}

// openDB extracts the Anki database from the archive and opens it.
func openDB(t *testing.T, archive []byte) *sql.DB {
	t.Helper()

	zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	require.NoError(t, err)

	files := map[string][]byte{}
	for _, f := range zr.File {
		rc, err := f.Open()
		require.NoError(t, err)
		files[f.Name], err = io.ReadAll(rc)
		require.NoError(t, err)
		require.NoError(t, rc.Close())
	}
	require.Contains(t, files, databaseFilename)
	assert.Equal(t, "{}", string(files[mediaFilename]))

	path := filepath.Join(t.TempDir(), databaseFilename)
	require.NoError(t, os.WriteFile(path, files[databaseFilename], 0o600))

	db, err := sql.Open("sqlite", path)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestBuild_RoundTrip(t *testing.T) {
	t.Parallel()

	archive, err := Build(testCollection())
	require.NoError(t, err)
	db := openDB(t, archive)

	rows, err := db.Query(`SELECT id, guid, mid, flds, sfld FROM notes ORDER BY id`)
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()

	type note struct {
		id, mid    int64
		guid, flds string
		sfld       string
	}
	var notes []note
	for rows.Next() {
		var n note
		require.NoError(t, rows.Scan(&n.id, &n.guid, &n.mid, &n.flds, &n.sfld))
		notes = append(notes, n)
	}
	require.NoError(t, rows.Err())
	require.Len(t, notes, 2)

	assert.Equal(t, int64(1111111), notes[0].id)
	assert.Equal(t, int64(555555), notes[0].mid)
	assert.Equal(t, "Welt\x1fworld", notes[0].flds, "fields are joined with the unit separator")
	assert.Equal(t, "Welt", notes[0].sfld)

	// The guid identifies a note across imports, so re-importing a word updates the card instead of duplicating it.
	sum := sha1.Sum([]byte("lingo_test" + "Welt"))
	assert.Equal(t, hex.EncodeToString(sum[:]), notes[0].guid)

	var cards, decks int
	require.NoError(t, db.QueryRow(`SELECT count(*) FROM cards WHERE did = 58098085085`).Scan(&cards))
	assert.Equal(t, 4, cards, "one card per note and template")
	require.NoError(t, db.QueryRow(`SELECT count(DISTINCT ord) FROM cards`).Scan(&decks))
	assert.Equal(t, 2, decks)

	var models, deckJSON string
	require.NoError(t, db.QueryRow(`SELECT models, decks FROM col`).Scan(&models, &deckJSON))
	var parsedModels map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(models), &parsedModels))
	assert.Contains(t, parsedModels, "555555")
	assert.Contains(t, deckJSON, "lingo_test")
	assert.True(t, strings.Contains(models, "{{FrontSide}}"), "templates are stored unescaped")
}

func TestBuild_UniqueCardIDs(t *testing.T) {
	t.Parallel()

	archive, err := Build(testCollection())
	require.NoError(t, err)
	db := openDB(t, archive)

	var total, unique int
	require.NoError(t, db.QueryRow(`SELECT count(*), count(DISTINCT id) FROM cards`).Scan(&total, &unique))
	assert.Equal(t, total, unique)
}
