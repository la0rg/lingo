package apkg

import (
	"time"

	"github.com/la0rg/lingo/apkg/apkgdb"
)

type Config apkgdb.DeckConfig

func DefaultConfig(id int64, name string) Config {
	return Config{
		ID:       id,
		Mod:      time.Date(2024, 05, 15, 11, 10, 0, 0, time.UTC).Unix(),
		Name:     name,
		Usn:      1,
		MaxTaken: 60,
		New: apkgdb.NewCardConfig{
			Delays:        []float64{1, 10},
			InitialFactor: 2500,
			Ints:          []int{1, 4, 0},
			PerDay:        10,
		},
		Rev: apkgdb.ReviewCardConfig{
			Ease4:  1.3,
			IvlFct: 1.0,
			MaxIvl: 36500,
			PerDay: 50,
		},
		Lapse: apkgdb.LapseCardConfig{
			Delays:      []float64{10},
			LeechAction: 1,
			LeechFails:  10,
			MinInt:      1,
		},
	}
}
