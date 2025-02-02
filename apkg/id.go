package apkg

import (
	"sync/atomic"
	"time"
)

var (
	// lastTimestamp tracks the last used timestamp to avoid duplicates
	lastTimestamp atomic.Int64
)

// NewID generates a unique ID based on current timestamp and sequence number.
//
// Anki DB uses epoch milliseconds as a unique identifier.
// Sequence is used to make sure unique card IDs are generated within single process.
func NewID() int64 {
	for {
		new := time.Now().UnixMilli()
		last := lastTimestamp.Load()

		if new <= last {
			new = last + 1
		}

		if lastTimestamp.CompareAndSwap(last, new) {
			return new
		}

		// CAS failed, retry
	}
}
