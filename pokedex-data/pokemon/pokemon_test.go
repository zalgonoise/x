package pokemon

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetPokemon(t *testing.T) {
	p, err := getPokemon(context.Background(), 1)
	require.NoError(t, err)

	t.Log(p)
}
