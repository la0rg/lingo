package apkg

import (
	"math"
	"math/rand"
	"time"
)

func NewID() int64 {
	return time.Now().UnixMilli() + rand.Int63n(math.MaxInt64)
}
