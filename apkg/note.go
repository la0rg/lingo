package apkg

import (
	"crypto/sha1"
	"encoding/hex"
	"time"

	"github.com/la0rg/lingo/apkg/apkgdb"
)

type Note struct {
	ID     int64
	Model  Model
	Fields []string
}

func (n Note) buildNote() apkgdb.Note {
	return apkgdb.Note{
		ID:     n.ID,
		Guid:   n.guid(),
		Mid:    n.Model.ID,
		Mod:    time.Now().Unix(),
		Usn:    1,
		Fields: n.Fields,
	}
}

func (n Note) buildCards(deckID int64) []apkgdb.Card {
	var cards = make([]apkgdb.Card, 0, len(n.Model.Templates))

	for ord := 0; ord < len(n.Model.Templates); ord++ {
		cards = append(cards, apkgdb.Card{
			ID:   NewID(),
			Nid:  n.ID,
			Did:  deckID,
			Ord:  ord,
			Mod:  time.Now().Unix(),
			Usn:  1,
			Due:  n.ID,
			Left: 1001, //  e.g.: '2004' means 2 reps left today and 4 reps till graduation
		})
	}

	return cards
}

func (n Note) guid() string {
	var key string
	if len(n.Fields) > 0 {
		key = n.Fields[0]
	}

	hash := sha1.Sum([]byte(n.Model.Name + key))
	return hex.EncodeToString(hash[:])
}
