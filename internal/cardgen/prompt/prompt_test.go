package prompt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSystem(t *testing.T) {
	t.Parallel()

	system, err := System()

	require.NoError(t, err)
	assert.Contains(t, system, "You are a dictionary API")
	assert.Contains(t, system, `"context_translation"`)
}
