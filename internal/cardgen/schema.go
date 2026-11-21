package cardgen

// cardSchema is the Structured Outputs schema of a generated card. It mirrors the json tags of model.Card.
// Strict mode requires every property to be required, so inapplicable fields are returned as empty strings.
var cardSchema = map[string]any{
	"type":                 "object",
	"additionalProperties": false,
	"required": []string{
		"word", "plural", "translations", "article", "gender",
		"transcription", "context", "context_translation", "definition",
	},
	"properties": map[string]any{
		"word":   map[string]any{"type": "string"},
		"plural": map[string]any{"type": "string"},
		"translations": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"required":             []string{"pos", "translations"},
				"properties": map[string]any{
					"pos":          map[string]any{"type": "string"},
					"translations": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
				},
			},
		},
		"article":             map[string]any{"type": "string"},
		"gender":              map[string]any{"type": "string"},
		"transcription":       map[string]any{"type": "string"},
		"context":             map[string]any{"type": "string"},
		"context_translation": map[string]any{"type": "string"},
		"definition":          map[string]any{"type": "string"},
	},
}
