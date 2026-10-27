package gacha

import (
	"context"
	"testing"
	"uuid"

	"github.com/stretchr/testify/require"
)

func TestIDFromContext(t *testing.T) {
	id := uuid.NewV7()

	for _, testcase := range []struct {
		name    string
		ctx     context.Context
		wantsID uuid.UUID
		wantsOK bool
	}{
		{name: "Simple", ctx: ContextWithID(t.Context(), id), wantsID: id, wantsOK: true},
		{name: "NilID", ctx: ContextWithID(t.Context(), uuid.Nil()), wantsID: uuid.Nil(), wantsOK: true},
		{name: "Missing", ctx: t.Context()},
		{
			name:    "Overwritten",
			ctx:     ContextWithID(ContextWithID(t.Context(), uuid.Max()), id),
			wantsID: id,
			wantsOK: true,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			got, ok := IDFromContext(testcase.ctx)
			require.Equal(t, testcase.wantsOK, ok)
			require.Equal(t, testcase.wantsID, got)
		})
	}
}

func TestSinkFunc(t *testing.T) {
	var got []string

	sink := SinkFunc[string](func(_ context.Context, data ...string) error {
		got = append(got, data...)

		return nil
	})

	require.NoError(t, sink.Write(t.Context(), "a", "b"))
	require.NoError(t, sink.Write(t.Context()))
	require.Equal(t, []string{"a", "b"}, got)
}
