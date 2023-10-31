package is

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"testing"
)

type testT struct {
	failCount int
	last      string
}

func (t *testT) Logf(format string, args ...any) {
	t.last = fmt.Sprintf(format, args...)
}

func (t *testT) Fail() {
	t.failCount++
}

func (t *testT) Helper() {}

func TestEqual(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		for _, testcase := range []struct {
			name     string
			expected string
			actual   string
			fails    bool
		}{
			{
				name:     "Success/Matching",
				expected: "some content",
				actual:   "some content",
			},
			{
				name:     "Fail/Mismatching",
				expected: "some content",
				actual:   "other content",
				fails:    true,
			},
			{
				name:     "Fails/Empty",
				expected: "some content",
				actual:   "",
				fails:    true,
			},
		} {
			t.Run(testcase.name, func(t *testing.T) {
				tt := &testT{}

				Equal(tt, testcase.expected, testcase.actual)

				if tt.failCount > 0 && !testcase.fails {
					t.Logf(tt.last)
					t.Fail()
				}
			})
		}
	})

	t.Run("Int", func(t *testing.T) {
		for _, testcase := range []struct {
			name     string
			expected int
			actual   int
			fails    bool
		}{
			{
				name:     "Success/Matching",
				expected: 58,
				actual:   58,
			},
			{
				name:     "Fail/Mismatching",
				expected: 58,
				actual:   45,
				fails:    true,
			},
			{
				name:     "Fails/Empty",
				expected: 58,
				actual:   0,
				fails:    true,
			},
		} {
			t.Run(testcase.name, func(t *testing.T) {
				tt := &testT{}

				Equal(tt, testcase.expected, testcase.actual)

				if tt.failCount > 0 && !testcase.fails {
					t.Logf(tt.last)
					t.Fail()
				}
			})
		}
	})

	t.Run("Float64", func(t *testing.T) {
		for _, testcase := range []struct {
			name     string
			expected float64
			actual   float64
			fails    bool
		}{
			{
				name:     "Success/Matching",
				expected: 58.0,
				actual:   58.0,
			},
			{
				name:     "Fail/Mismatching",
				expected: 58.0,
				actual:   45.0,
				fails:    true,
			},
			{
				name:     "Fails/Empty",
				expected: 58.0,
				actual:   0.0,
				fails:    true,
			},
			{
				name:     "Fails/NaN",
				expected: math.NaN(),
				actual:   math.NaN(),
				fails:    true,
			},
			{
				name:     "Fails/+Inf",
				expected: math.Inf(1),
				actual:   math.Inf(1),
				fails:    true,
			},
			{
				name:     "Fails/-Inf",
				expected: math.Inf(-1),
				actual:   math.Inf(-1),
				fails:    true,
			},
		} {
			t.Run(testcase.name, func(t *testing.T) {
				tt := &testT{}

				Equal(tt, testcase.expected, testcase.actual)

				if tt.failCount > 0 && !testcase.fails {
					t.Logf(tt.last)
					t.Fail()
				}
			})
		}
	})

	t.Run("CustomType", func(t *testing.T) {
		type person struct {
			name string
			id   int
		}

		for _, testcase := range []struct {
			name     string
			expected person
			actual   person
			fails    bool
		}{
			{
				name:     "Success/Matching",
				expected: person{"gopher", 10},
				actual:   person{"gopher", 10},
			},
			{
				name:     "Fail/MismatchingName",
				expected: person{"gopher", 10},
				actual:   person{"go", 10},
				fails:    true,
			},
			{
				name:     "Fail/MismatchingID",
				expected: person{"gopher", 10},
				actual:   person{"gopher", 100},
				fails:    true,
			},
			{
				name:     "Fails/Empty",
				expected: person{"gopher", 10},
				actual:   person{"", 0},
				fails:    true,
			},
		} {
			t.Run(testcase.name, func(t *testing.T) {
				tt := &testT{}

				Equal(tt, testcase.expected, testcase.actual)

				if tt.failCount > 0 && !testcase.fails {
					t.Logf(tt.last)
					t.Fail()
				}
			})
		}
	})

	t.Run("Pointer", func(t *testing.T) {
		type person struct {
			name string
			id   int
		}

		for _, testcase := range []struct {
			name     string
			expected *person
			actual   *person
			fails    bool
		}{
			{
				name:     "Success/Matching",
				expected: &person{"gopher", 10},
				actual:   &person{"gopher", 10},
			},
			{
				name:     "Fail/MismatchingName",
				expected: &person{"gopher", 10},
				actual:   &person{"go", 10},
				fails:    true,
			},
			{
				name:     "Fail/MismatchingID",
				expected: &person{"gopher", 10},
				actual:   &person{"gopher", 100},
				fails:    true,
			},
			{
				name:     "Fails/Empty",
				expected: &person{"gopher", 10},
				actual:   &person{"", 0},
				fails:    true,
			},
			{
				name:     "Fails/Nil",
				expected: &person{"gopher", 10},
				actual:   nil,
				fails:    true,
			},
			{
				name:     "Fails/BothNil",
				expected: nil,
				actual:   nil,
			},
		} {
			t.Run(testcase.name, func(t *testing.T) {
				tt := &testT{}

				EqualValue(tt, testcase.expected, testcase.actual)

				if tt.failCount > 0 && !testcase.fails {
					t.Logf(tt.last)
					t.Fail()
				}
			})
		}
	})
}

func TestEmpty(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		for _, testcase := range []struct {
			name  string
			input string
			fails bool
		}{
			{
				name:  "NotEmpty",
				input: "some string",
				fails: true,
			},
			{
				name:  "Empty",
				input: "",
			},
		} {
			t.Run(testcase.name, func(t *testing.T) {
				tt := &testT{}

				Empty(tt, testcase.input)

				if tt.failCount > 0 && !testcase.fails {
					t.Logf(tt.last)
					t.Fail()
				}
			})
		}
	})
	t.Run("Int", func(t *testing.T) {
		for _, testcase := range []struct {
			name  string
			input int
			fails bool
		}{
			{
				name:  "NotEmpty/Positive",
				input: 10,
				fails: true,
			},
			{
				name:  "NotEmpty/Negative",
				input: -10,
				fails: true,
			},
			{
				name:  "Empty",
				input: 0,
			},
		} {
			t.Run(testcase.name, func(t *testing.T) {
				tt := &testT{}

				Empty(tt, testcase.input)

				if tt.failCount > 0 && !testcase.fails {
					t.Logf(tt.last)
					t.Fail()
				}
			})
		}
	})
	t.Run("Float", func(t *testing.T) {
		for _, testcase := range []struct {
			name  string
			input float64
			fails bool
		}{
			{
				name:  "NotEmpty/Positive",
				input: 10.0,
				fails: true,
			},
			{
				name:  "NotEmpty/Negative",
				input: -10.0,
				fails: true,
			},
			{
				name:  "NotEmpty/NaN",
				input: math.NaN(),
				fails: true,
			},
			{
				name:  "NotEmpty/+Inf",
				input: math.Inf(1),
				fails: true,
			},
			{
				name:  "NotEmpty/-Inf",
				input: math.Inf(-1),
				fails: true,
			},
			{
				name:  "Empty",
				input: 0.0,
			},
		} {
			t.Run(testcase.name, func(t *testing.T) {
				tt := &testT{}

				Empty(tt, testcase.input)

				if tt.failCount > 0 && !testcase.fails {
					t.Logf(tt.last)
					t.Fail()
				}
			})
		}
	})
	t.Run("Error", func(t *testing.T) {
		for _, testcase := range []struct {
			name  string
			input error
			fails bool
		}{
			{
				name:  "Static/io.EOF",
				input: io.EOF,
				fails: true,
			},
			{
				name: "Dynamic/json.UnmarshalError",
				input: &json.MarshalerError{
					Err: io.EOF,
				},
				fails: true,
			},
			{
				name:  "Static/Custom",
				input: errors.New("test error"),
				fails: true,
			},
			{
				name:  "Static/Custom",
				input: nil,
			},
		} {
			t.Run(testcase.name, func(t *testing.T) {
				tt := &testT{}

				Empty(tt, testcase.input)

				if tt.failCount > 0 && !testcase.fails {
					t.Logf(tt.last)
					t.Fail()
				}
			})
		}
	})

	t.Run("Pointer", func(t *testing.T) {
		type person struct {
			name string
			id   int
		}

		for _, testcase := range []struct {
			name  string
			input *person
			fails bool
		}{
			{
				name:  "Fails/Populated",
				input: &person{"gopher", 10},
				fails: true,
			},
			{
				name:  "Success/Empty",
				input: &person{"", 0},
			},
			{
				name:  "Success/Nil",
				input: nil,
			},
		} {
			t.Run(testcase.name, func(t *testing.T) {
				tt := &testT{}

				EmptyValue(tt, testcase.input)

				if tt.failCount > 0 && !testcase.fails {
					t.Logf(tt.last)
					t.Fail()
				}
			})
		}
	})
}

// TestNilError covers NilError, however more in-depth tests for Empty can be found above.
func TestNilError(t *testing.T) {
	t.Run("WithError", func(t *testing.T) {
		tt := &testT{}

		NilError(tt, io.EOF)

		if tt.failCount == 0 {
			t.Logf(tt.last)
			t.Fail()
		}
	})

	t.Run("NilError", func(t *testing.T) {
		tt := &testT{}

		NilError(tt, nil)

		if tt.failCount != 0 {
			t.Logf(tt.last)
			t.Fail()
		}
	})
}

// TestTrueFalse covers True and False, however more in-depth tests for Equal can be found above.
func TestTrueFalse(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		for _, testcase := range []struct {
			name  string
			input bool
			wants int
		}{
			{
				name:  "Success",
				input: true,
				wants: 0,
			},
			{
				name:  "Fail",
				input: false,
				wants: 1,
			},
		} {
			t.Run(testcase.name, func(t *testing.T) {
				tt := &testT{}

				True(tt, testcase.input)

				if tt.failCount != testcase.wants {
					t.Logf(tt.last)
					t.Fail()
				}
			})
		}
	})
	t.Run("False", func(t *testing.T) {
		for _, testcase := range []struct {
			name  string
			input bool
			wants int
		}{
			{
				name:  "Success",
				input: false,
				wants: 0,
			},
			{
				name:  "Fail",
				input: true,
				wants: 1,
			},
		} {
			t.Run(testcase.name, func(t *testing.T) {
				tt := &testT{}

				False(tt, testcase.input)

				if tt.failCount != testcase.wants {
					t.Logf(tt.last)
					t.Fail()
				}
			})
		}
	})
}
