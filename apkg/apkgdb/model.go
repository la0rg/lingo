package apkgdb

import (
	"crypto/sha1"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

type Collection struct {
	Crt    time.Time
	Mod    time.Time
	Scm    time.Time
	Models Models
	Decks  Decks
	DConf  DConf
}

type Models map[string]Model

func (m Models) Value() (driver.Value, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

type Decks map[string]Deck

func (d Decks) Value() (driver.Value, error) {
	b, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

type DConf map[string]DeckConfig

func (d DConf) Value() (driver.Value, error) {
	b, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

type DeckConfig struct {
	ID       int64            `json:"id"`
	Mod      int64            `json:"mod"`
	Name     string           `json:"name"`
	Usn      int              `json:"usn"`
	MaxTaken int              `json:"maxTaken"`
	Autoplay bool             `json:"autoplay"`
	Timer    int              `json:"timer"`
	Replayq  bool             `json:"replayq"`
	New      NewCardConfig    `json:"new"`
	Rev      ReviewCardConfig `json:"rev"`
	Lapse    LapseCardConfig  `json:"lapse"`
	Dyn      bool             `json:"dyn"`
}

type NewCardConfig struct {
	Bury          bool      `json:"bury"`
	Delays        []float64 `json:"delays"`
	InitialFactor int       `json:"initialFactor"`
	Ints          []int     `json:"ints"`
	Order         int       `json:"order"`
	PerDay        int       `json:"perDay"`
}

type ReviewCardConfig struct {
	Bury   bool    `json:"bury"`
	Ease4  float64 `json:"ease4"`
	IvlFct float64 `json:"ivlFct"`
	MaxIvl int     `json:"maxIvl"`
	PerDay int     `json:"perDay"`
}

type LapseCardConfig struct {
	Delays      []float64 `json:"delays"`
	LeechAction int       `json:"leechAction"`
	LeechFails  int       `json:"leechFails"`
	MinInt      int       `json:"minInt"`
	Mult        float64   `json:"mult"`
}

type Model struct {
	ID        int64           `json:"id"`
	Name      string          `json:"name"`
	Type      int             `json:"type"`
	Mod       int64           `json:"mod"`
	Usn       int             `json:"usn"`
	Sortf     int             `json:"sortf"`
	Did       int64           `json:"did"`
	Templates []Template      `json:"tmpls"`
	Fields    []Field         `json:"flds"`
	Css       UnescapedString `json:"css"`
	LatexPre  string          `json:"latexPre"`
	LatexPost string          `json:"latexPost"`
	Latexsvg  bool            `json:"latexsvg"`
	Tags      []string        `json:"tags"`
}

type Template struct {
	Name string          `json:"name"`
	Ord  int             `json:"ord"`
	Qfmt UnescapedString `json:"qfmt"`
	Afmt UnescapedString `json:"afmt"`
}

type Field struct {
	Name   string `json:"name"`
	Ord    int    `json:"ord"`
	Font   string `json:"font"`
	Size   int    `json:"size"`
	Sticky bool   `json:"sticky"`
	Rtl    bool   `json:"rtl"`
}

type UnescapedString string

func (s UnescapedString) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, s)), nil
}

type Deck struct {
	ID               int64  `json:"id"`
	Mod              int64  `json:"mod"`
	Name             string `json:"name"`
	Usn              int    `json:"usn"`
	LrnToday         [2]int `json:"lrnToday"`
	RevToday         [2]int `json:"revToday"`
	NewToday         [2]int `json:"newToday"`
	TimeToday        [2]int `json:"timeToday"`
	Collapsed        bool   `json:"collapsed"`
	BrowserCollapsed bool   `json:"browserCollapsed"`
	Desc             string `json:"desc"`
	Dyn              int    `json:"dyn"`
	Conf             int64  `json:"conf"`
}

type Note struct {
	ID     int64
	Guid   string
	Mid    int64
	Mod    int64
	Usn    int
	Tags   string
	Fields []string
}

func (n Note) SField() string {
	if len(n.Fields) == 0 {
		return ""
	}

	return n.Fields[0]
}

func (n Note) Csum() string {
	if len(n.Fields) == 0 {
		return ""
	}

	sha := sha1.Sum([]byte(n.Fields[0]))
	s := hex.EncodeToString(sha[:])
	return s[:8]
}

type Card struct {
	ID     int64
	Nid    int64
	Did    int64
	Ord    int
	Mod    int64
	Usn    int
	Type   int
	Queue  int
	Due    int64
	Ivl    int64
	Factor int
	Reps   int
	Lapses int
	Left   int
	ODue   int
	ODid   int
	Flags  int
}
