package mock

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/openai/openai-go/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/la0rg/lingo/internal/model"
)

func TestGPTClient_EchoesWord(t *testing.T) {
	t.Parallel()

	resp, err := GPTClient{}.New(context.Background(), openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("system"),
			openai.UserMessage(`{"source_language":"de","target_language":"en","word":"apfel"}`),
		},
	})
	require.NoError(t, err)
	require.Len(t, resp.Choices, 1)
	assert.Equal(t, "stop", resp.Choices[0].FinishReason)

	var card model.Card
	require.NoError(t, json.Unmarshal([]byte(resp.Choices[0].Message.Content), &card))
	assert.Equal(t, "apfel", card.Word)
	assert.NotEmpty(t, card.Translations)
}
