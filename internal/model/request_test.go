package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPKGRequest_Normalize(t *testing.T) {
	t.Parallel()

	r := APKGRequest{Words: []string{" Apfel ", "", "APFEL", "  ", "Haus", "der Hund"}}
	r.Normalize()
	assert.ElementsMatch(t, []string{"apfel", "haus", "der hund"}, r.Words)

	empty := APKGRequest{}
	empty.Normalize()
	assert.Empty(t, empty.Words)
}

func TestAPKGRequest_Validate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		words   []string
		max     int
		maxLen  int
		wantErr error
	}{
		{"valid", []string{"мир", "abc"}, 2, 3, nil},
		{"no words", nil, 2, 3, ErrEmptyWordList},
		{"too many words", []string{"a", "b", "c"}, 2, 3, &TooManyWordsError{Max: 2}},
		{"word too long", []string{"abcd"}, 2, 3, &WordTooLongError{Max: 3}},
		{"no limits", []string{"a", "b", "c", "a very long word indeed"}, 0, 0, nil},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			r := APKGRequest{Words: c.words}
			err := r.Validate(c.max, c.maxLen)

			if c.wantErr == nil {
				require.NoError(t, err)
				return
			}
			assert.Equal(t, c.wantErr, err)
		})
	}
}

func TestValidateWord(t *testing.T) {
	t.Parallel()

	assert.NoError(t, ValidateWord("Äpfel", 5), "length is counted in runes")
	assert.Equal(t, &WordTooLongError{Max: 4}, ValidateWord("Äpfel", 4))
	assert.NoError(t, ValidateWord("anything", 0))
}

func TestErrorMessages(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "too many words, the maximum is 5", (&TooManyWordsError{Max: 5}).Error())
	assert.Equal(t, "word is too long, the maximum is 9 characters", (&WordTooLongError{Max: 9}).Error())
}
