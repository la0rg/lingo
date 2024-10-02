package maps

import (
	"strconv"
)

func KeysToString[V any](m map[int64]V) map[string]V {
	result := make(map[string]V, len(m))
	for id, v := range m {
		result[strconv.FormatInt(id, 10)] = v
	}
	return result
}
