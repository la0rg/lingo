//go:build live

package cardgen

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/la0rg/lingo/internal/model"
)

// TestLive calls the real OpenAI API: OPENAIKEY=sk-... go test -tags live -v ./internal/cardgen/
// Set OPENAIMODEL to try another model. Review the printed cards (IPA, articles, gender) by hand.
func TestLive(t *testing.T) {
	key := os.Getenv("OPENAIKEY")
	if key == "" {
		t.Skip("OPENAIKEY is not set")
	}

	var opts []Option
	if m := os.Getenv("OPENAIMODEL"); m != "" {
		opts = append(opts, WithModel(m))
	}

	client := openai.NewClient(option.WithAPIKey(key))
	g, err := NewOpenAI(&client.Chat.Completions, opts...)
	require.NoError(t, err)

	cases := []struct {
		source, target, word string
		valid                bool
	}{
		{"de", "en", "apfel", true},
		{"de", "ru", "Welt", true},
		{"ru", "en", "мир", true},
		{"en", "de", "book", true},
		{"de", "en", "qwrtzxv", false},
	}

	for _, c := range cases {
		t.Run(c.source+"-"+c.target+"/"+c.word, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()

			start := time.Now()
			card, err := g.Generate(ctx, &model.Collection{SourceLanguage: c.source, TargetLanguage: c.target}, c.word)
			require.NoError(t, err)
			t.Logf("latency %v, card %+v", time.Since(start), card)

			if !c.valid {
				assert.Nil(t, card)
				return
			}
			require.NotNil(t, card)
			assert.NotEmpty(t, card.Word)
			assert.NotEmpty(t, card.Translations)
			assert.NotEmpty(t, card.Transcription)
			assert.NotEmpty(t, card.Context)
		})
	}
}
