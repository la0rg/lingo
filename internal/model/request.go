package model

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/la0rg/lingo/internal/maps"
)

var ErrEmptyWordList = errors.New("word list is empty")

// TooManyWordsError is returned when a request has more words than allowed.
type TooManyWordsError struct{ Max int }

func (e *TooManyWordsError) Error() string {
	return fmt.Sprintf("too many words, the maximum is %d", e.Max)
}

// WordTooLongError is returned when a word is longer than allowed.
type WordTooLongError struct{ Max int }

func (e *WordTooLongError) Error() string {
	return fmt.Sprintf("word is too long, the maximum is %d characters", e.Max)
}

type APKGRequest struct {
	Words []string `json:"words"`
}

// Normalize trims, lowercases and deduplicates the words, dropping empty ones.
func (r *APKGRequest) Normalize() {
	words := make([]string, 0, len(r.Words))
	for _, word := range r.Words {
		if word = strings.ToLower(strings.TrimSpace(word)); word != "" {
			words = append(words, word)
		}
	}
	r.Words = maps.Unique(words)
}

// Validate checks the number of words and the length of each word. A zero limit disables the check.
func (r *APKGRequest) Validate(maxWords, maxWordLength int) error {
	if len(r.Words) == 0 {
		return ErrEmptyWordList
	}

	if maxWords > 0 && len(r.Words) > maxWords {
		return &TooManyWordsError{Max: maxWords}
	}

	for _, word := range r.Words {
		if err := ValidateWord(word, maxWordLength); err != nil {
			return err
		}
	}

	return nil
}

// ValidateWord checks the length of a single word in runes. A zero limit disables the check.
func ValidateWord(word string, maxLength int) error {
	if maxLength > 0 && utf8.RuneCountInString(word) > maxLength {
		return &WordTooLongError{Max: maxLength}
	}
	return nil
}

type APKGResponse struct {
	URL string `json:"url"`
}
