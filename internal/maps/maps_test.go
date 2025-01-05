package maps

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeysToString(t *testing.T) {
	m := map[int64]string{
		1: "a",
		2: "b",
		3: "c",
	}

	result := KeysToString(m)
	expected := map[string]string{
		"1": "a",
		"2": "b",
		"3": "c",
	}
	assert.Equal(t, expected, result)
}
