package strings

import "strings"

func SplitAny(s string, chars string) []string {
	splitFunc := func(r rune) bool {
		return strings.ContainsRune(chars, r)
	}

	return strings.FieldsFunc(s, splitFunc)
}
