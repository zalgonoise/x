package level

import "testing"

func TestLevelString(t *testing.T) {
	tc := []struct {
		input Level
		wants string
	}{
		{
			input: Trace,
			wants: "trace",
		}, {
			input: lDebug,
			wants: "debug",
		}, {
			input: lv(2),
			wants: "info",
		}, {
			input: lv(3),
			wants: "warn",
		}, {
			input: lv(4),
			wants: "error",
		}, {
			input: lv(5),
			wants: "fatal",
		}, {
			input: lv(99),
			wants: "",
		},
	}

	for _, tt := range tc {
		if tt.input.String() != tt.wants {
			t.Errorf("unexpected output error: wanted %s ; got %s", tt.wants, tt.input.String())
		}
	}
}
func TestLevelInt(t *testing.T) {
	tc := []struct {
		input Level
		wants int
	}{
		{
			input: Trace,
			wants: 0,
		}, {
			input: lDebug,
			wants: 1,
		}, {
			input: lv(2),
			wants: 2,
		}, {
			input: lv(3),
			wants: 3,
		}, {
			input: lv(4),
			wants: 4,
		}, {
			input: lv(5),
			wants: 5,
		}, {
			input: lv(99),
			wants: 99,
		},
	}

	for _, tt := range tc {
		if tt.input.Int() != tt.wants {
			t.Errorf("unexpected output error: wanted %v ; got %v", tt.wants, tt.input.Int())
		}
	}
}
func TestAsLevel(t *testing.T) {
	tc := []struct {
		input string
		wants Level
	}{
		{
			input: "trace",
			wants: Trace,
		}, {
			input: "debug",
			wants: Debug,
		}, {
			input: "info",
			wants: Info,
		}, {
			input: "warn",
			wants: Warn,
		}, {
			input: "error",
			wants: Error,
		}, {
			input: "fatal",
			wants: Fatal,
		}, {
			input: "",
			wants: nil,
		},
	}

	for _, tt := range tc {
		if AsLevel(tt.input) != tt.wants {
			t.Errorf("unexpected output error: wanted %v ; got %v", tt.wants, AsLevel(tt.input))
		}
	}
}
