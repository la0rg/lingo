package strings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitAny(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   string
		want []string
	}{
		{"", []string{}},
		{"a", []string{"a"}},
		{"a,b\nc", []string{"a", "b", "c"}},
		{",,a,\n\nb,", []string{"a", "b"}},
		{"der Hund, die Katze", []string{"der Hund", " die Katze"}},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, c.want, SplitAny(c.in, "\n,"))
		})
	}
}
