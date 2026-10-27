package gacha

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type permanence bool

func (p permanence) Error() string   { return fmt.Sprintf("permanence(%t)", bool(p)) }
func (p permanence) Permanent() bool { return bool(p) }

func TestPermanent(t *testing.T) {
	errBase := errors.New("base")

	for _, testcase := range []struct {
		name      string
		err       error
		permanent bool
		wantsText string
		wantsIs   error
	}{
		{name: "Nil", err: nil},
		{name: "Plain", err: errBase, wantsText: "base", wantsIs: errBase},
		{name: "Permanent", err: Permanent(errBase), permanent: true, wantsText: "base", wantsIs: errBase},
		{
			name:      "WrappedPermanent",
			err:       fmt.Errorf("handling: %w", Permanent(errBase)),
			permanent: true,
			wantsText: "handling: base",
			wantsIs:   errBase,
		},
		{
			name:      "JoinedPermanent",
			err:       errors.Join(errors.New("other"), Permanent(errBase)),
			permanent: true,
			wantsText: "other\nbase",
			wantsIs:   errBase,
		},
		{name: "CustomPermanent", err: permanence(true), permanent: true, wantsText: "permanence(true)"},
		{name: "CustomNotPermanent", err: permanence(false), wantsText: "permanence(false)"},
		{
			name:      "OutermostWins",
			err:       fmt.Errorf("%w: %w", permanence(false), Permanent(errBase)),
			wantsText: "permanence(false): base",
			wantsIs:   errBase,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			require.Equal(t, testcase.permanent, IsPermanent(testcase.err))

			if testcase.err == nil {
				require.NoError(t, Permanent(nil))

				return
			}

			require.Equal(t, testcase.wantsText, testcase.err.Error())

			if testcase.wantsIs != nil {
				require.ErrorIs(t, testcase.err, testcase.wantsIs)
			}

			// Marking any error as permanent makes it permanent, keeping its message.
			marked := Permanent(testcase.err)
			require.True(t, IsPermanent(marked))
			require.Equal(t, testcase.wantsText, marked.Error())
			require.ErrorIs(t, marked, testcase.err)
		})
	}
}
