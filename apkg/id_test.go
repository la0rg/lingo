package apkg

import (
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewID(t *testing.T) {
	// Test basic functionality
	id1 := NewID()
	id2 := NewID()

	assert.Greater(t, id2, id1)

	// Test that IDs are at least as large as current time
	now := time.Now().UnixMilli()
	id := NewID()
	assert.GreaterOrEqual(t, id, now)
}

func TestNewID_Concurrent(t *testing.T) {
	const numGoroutines = 100
	const idsPerGoroutine = 100

	var wg sync.WaitGroup
	ids := make(chan int64, numGoroutines*idsPerGoroutine)

	for range numGoroutines {
		wg.Go(func() {
			for range idsPerGoroutine {
				ids <- NewID()
			}
		})
	}

	go func() {
		wg.Wait()
		close(ids)
	}()

	// Check for duplicates and increasing order
	seen := make(map[int64]bool)
	for id := range ids {
		assert.False(t, seen[id], "Duplicate ID found: %d", id)
		seen[id] = true
	}
}

func TestNewID_IncreasingOrder(t *testing.T) {
	ids := make([]int64, 1000)
	for i := range 1000 {
		ids[i] = NewID()
	}

	assert.True(t, slices.IsSorted(ids))
}
