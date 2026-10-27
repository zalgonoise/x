package gacha

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStatus(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		status    Status
		wantsText string
		valid     bool
	}{
		{name: "Pending", status: StatusPending, wantsText: "pending", valid: true},
		{name: "Processing", status: StatusProcessing, wantsText: "processing", valid: true},
		{name: "Completed", status: StatusCompleted, wantsText: "completed", valid: true},
		{name: "Failed", status: StatusFailed, wantsText: "failed", valid: true},
		{name: "Released", status: StatusReleased, wantsText: "released", valid: true},
		{name: "Invalid/Zero", status: 0, wantsText: "Status(0)"},
		{name: "Invalid/OutOfRange", status: StatusReleased + 1, wantsText: "Status(6)"},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			require.Equal(t, testcase.wantsText, testcase.status.String())
			require.Equal(t, testcase.valid, testcase.status.IsValid())

			text, err := testcase.status.MarshalText()
			if !testcase.valid {
				require.ErrorIs(t, err, ErrInvalidStatus)
				require.Nil(t, text)

				_, err = ParseStatus(testcase.wantsText)
				require.ErrorIs(t, err, ErrInvalidStatus)

				return
			}

			require.NoError(t, err)
			require.Equal(t, testcase.wantsText, string(text))

			parsed, err := ParseStatus(testcase.wantsText)
			require.NoError(t, err)
			require.Equal(t, testcase.status, parsed)

			var unmarshaled Status

			require.NoError(t, unmarshaled.UnmarshalText(text))
			require.Equal(t, testcase.status, unmarshaled)
		})
	}
}

func TestStatus_JSON(t *testing.T) {
	type wrapper struct {
		Status Status `json:"status"`
	}

	for _, testcase := range []struct {
		name  string
		input string
		wants Status
		err   error
	}{
		{name: "Simple/Completed", input: `{"status":"completed"}`, wants: StatusCompleted},
		{name: "Unknown", input: `{"status":"done"}`, err: ErrInvalidStatus},
		{name: "Empty", input: `{"status":""}`, err: ErrInvalidStatus},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			var w wrapper

			err := json.Unmarshal([]byte(testcase.input), &w)
			if testcase.err != nil {
				require.ErrorIs(t, err, testcase.err)

				return
			}

			require.NoError(t, err)
			require.Equal(t, testcase.wants, w.Status)

			out, err := json.Marshal(w)
			require.NoError(t, err)
			require.JSONEq(t, testcase.input, string(out))
		})
	}
}
