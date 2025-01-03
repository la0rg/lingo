package maps

import (
	"iter"
	"maps"
	"slices"
	"strconv"

	"golang.org/x/exp/constraints"
)

func KeysToString[K constraints.Signed, V any](m map[K]V) map[string]V {
	result := make(map[string]V, len(m))
	for id, v := range m {
		result[strconv.FormatInt(int64(id), 10)] = v
	}
	return result
}

func Set[K comparable](keys []K) iter.Seq2[K, struct{}] {
	return func(yield func(K, struct{}) bool) {
		for _, k := range keys {
			if !yield(k, struct{}{}) {
				return
			}
		}
	}
}

func Unique[K comparable](keys []K) []K {
	return slices.Collect(maps.Keys(maps.Collect(Set(keys))))
}
