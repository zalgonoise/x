package errs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJoin(t *testing.T) {
	testErr1 := WithDomain("x/errs", "first", "test error")
	testErr2 := Sentinel("second", "test error without domain")
	testErr3 := WithDomain("x/errs", "third", "test error")
	testErr4 := Sentinel("fourth", "test error without domain")
	testErr5 := WithDomain("x/errs", "fifth", "test error")

	for _, testcase := range []struct {
		name        string
		input       []error
		wantsString string
	}{
		{
			name: "NoInput",
		},
		{
			name:        "OneError",
			input:       []error{testErr1},
			wantsString: "x/errs: first test error",
		},
		{
			name:        "TwoErrors",
			input:       []error{testErr1, testErr2},
			wantsString: "first test error: second test error without domain",
		},
		{
			name:        "FiveErrorsWithAndWithoutDomain",
			input:       []error{testErr1, testErr2, testErr3, testErr4, testErr5},
			wantsString: "first test error: second test error without domain: third test error: fourth test error without domain: fifth test error",
		},
		{
			name:        "FiveErrorsButThreeAreNil",
			input:       []error{nil, testErr1, nil, testErr2, nil},
			wantsString: "first test error: second test error without domain",
		},
		{
			name:  "FiveErrorsAllNil",
			input: []error{nil, nil, nil, nil, nil},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			err := Join(testcase.input...)

			for i := range testcase.input {
				if testcase.input[i] != nil {
					require.ErrorIs(t, err, testcase.input[i])
				}
			}

			if err != nil {
				require.Equal(t, testcase.wantsString, err.Error())
			}
		})
	}
}

func TestJoinWith(t *testing.T) {
	testErr1 := WithDomain("x/errs", "first", "test error")
	testErr2 := Sentinel("second", "test error without domain")
	testErr3 := WithDomain("x/errs", "third", "test error")
	testErr4 := Sentinel("fourth", "test error without domain")
	testErr5 := WithDomain("x/errs", "fifth", "test error")

	for _, testcase := range []struct {
		name        string
		input       []error
		sep         string
		wantsString string
	}{
		{
			name: "NoInput",
		},
		{
			name:        "OneError",
			input:       []error{testErr1},
			wantsString: "x/errs: first test error",
		},
		{
			name:        "TwoErrors",
			sep:         " -- ",
			input:       []error{testErr1, testErr2},
			wantsString: "first test error -- second test error without domain",
		},
		{
			name:        "FiveErrorsWithAndWithoutDomain",
			sep:         "; ",
			input:       []error{testErr1, testErr2, testErr3, testErr4, testErr5},
			wantsString: "first test error; second test error without domain; third test error; fourth test error without domain; fifth test error",
		},
		{
			name:        "FiveErrorsWithAndWithoutDomainDefaultSeparator",
			input:       []error{testErr1, testErr2, testErr3, testErr4, testErr5},
			wantsString: "first test error: second test error without domain: third test error: fourth test error without domain: fifth test error",
		},
		{
			name:        "FiveErrorsButThreeAreNil",
			sep:         " -> ",
			input:       []error{nil, testErr1, nil, testErr2, nil},
			wantsString: "first test error -> second test error without domain",
		},
		{
			name:  "FiveErrorsAllNil",
			input: []error{nil, nil, nil, nil, nil},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			err := JoinWith(testcase.sep, testcase.input...)

			for i := range testcase.input {
				if testcase.input[i] != nil {
					require.ErrorIs(t, err, testcase.input[i])
				}
			}

			if err != nil {
				require.Equal(t, testcase.wantsString, err.Error())
			}
		})
	}
}
