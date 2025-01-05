package strings

import (
	"strings"
)

// SplitAny splits a string into substrings using the characters in chars as separators.
func SplitAny(s string, chars string) []string {
	splitFunc := func(r rune) bool {
		return strings.ContainsRune(chars, r)
	}

	return strings.FieldsFunc(s, splitFunc)
}
