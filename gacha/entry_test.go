package gacha

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOutcome_Validate(t *testing.T) {
	for _, testcase := range []struct {
		name   string
		status Status
		err    error
	}{
		{name: "Completed", status: StatusCompleted},
		{name: "Pending", status: StatusPending},
		{name: "Failed", status: StatusFailed},
		{name: "Released", status: StatusReleased},
		{name: "Processing", status: StatusProcessing, err: ErrInvalidStatus},
		{name: "Zero", status: 0, err: ErrInvalidStatus},
		{name: "OutOfRange", status: StatusReleased + 1, err: ErrInvalidStatus},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			err := Outcome{Status: testcase.status}.Validate()
			if testcase.err != nil {
				require.ErrorIs(t, err, testcase.err)

				return
			}

			require.NoError(t, err)
		})
	}
}
