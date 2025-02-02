package apkg

import (
	"math/rand"
	"time"
)

const randLength = 10_000

// NewID creates an ID based on the current time and a random number.
// This is useful for creating ANKI DB identifiers.
func NewID() int64 {
	return time.Now().UnixMilli() + rand.Int63n(randLength)
}
