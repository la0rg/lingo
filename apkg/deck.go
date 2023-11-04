package apkg

import (
	"time"

	"github.com/la0rg/lingo/apkg/apkgdb"
)

type Deck struct {
	ID     int64
	Name   string
	Config Config
	Notes  []Note
}

func (d *Deck) buildDeck() apkgdb.Deck {
	return apkgdb.Deck{
		ID:               d.ID,
		Name:             d.Name,
		Mod:              time.Now().Unix(),
		Usn:              1,
		Collapsed:        true,
		BrowserCollapsed: true,
		Conf:             d.Config.ID,
	}
}
