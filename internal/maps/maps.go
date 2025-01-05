package maps

import (
	"iter"
	"maps"
	"slices"
	"strconv"

	"golang.org/x/exp/constraints"
)

// KeysToString converts the keys of the input map to strings.
func KeysToString[K constraints.Signed, V any](m map[K]V) map[string]V {
	result := make(map[string]V, len(m))
	for id, v := range m {
		result[strconv.FormatInt(int64(id), 10)] = v
	}
	return result
}

// Set returns a Seq2 iterator for the given slice of keys. The value type is struct{}.
// It is used to create a set (map[K]struct{}) from a slice of keys.
func Set[K comparable](keys []K) iter.Seq2[K, struct{}] {
	return func(yield func(K, struct{}) bool) {
		for _, k := range keys {
			if !yield(k, struct{}{}) {
				return
			}
		}
	}
}

// Unique returns a slice of unique keys from the input slice.
func Unique[K comparable](keys []K) []K {
	return slices.Collect(maps.Keys(maps.Collect(Set(keys))))
}
