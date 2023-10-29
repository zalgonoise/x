package validation

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testConfig struct{}

func TestNew(t *testing.T) {
	var (
		testNoOp = func(testConfig) error { return nil }
	)

	for _, testcase := range []struct {
		name     string
		input    []func(testConfig) error
		isNoOp   bool
		wantsLen int
	}{
		{
			name:   "None",
			isNoOp: true,
		},
		{
			name: "One",
			input: []func(config testConfig) error{
				testNoOp,
			},
		},
		{
			name: "Three",
			input: []func(config testConfig) error{
				testNoOp, testNoOp, testNoOp,
			},
			wantsLen: 3,
		},
		{
			name: "WithNilsOneIsValid",
			input: []func(config testConfig) error{
				nil, nil, testNoOp, nil, nil,
			},
		},
		{
			name: "AllNil",
			input: []func(config testConfig) error{
				nil, nil, nil, nil, nil,
			},
			isNoOp: true,
		},
		{
			name: "ThreeWithNils",
			input: []func(config testConfig) error{
				testNoOp, nil, testNoOp, nil, testNoOp, nil,
			},
			wantsLen: 3,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			validator := New(testcase.input...)

			_ = validator.Validate(testConfig{})

			switch v := validator.(type) {
			case noOpValidator[testConfig]:
				require.True(t, testcase.isNoOp)
				return
			case multiValidator[testConfig]:
				require.Equal(t, testcase.wantsLen, len(v.validators))
				return
			}

			require.False(t, testcase.isNoOp)
		})
	}
}

func TestJoin(t *testing.T) {
	var testNoOp = Func[testConfig](func(testConfig) error { return nil })

	for _, testcase := range []struct {
		name     string
		input    []Validator[testConfig]
		isNoOp   bool
		wantsLen int
	}{
		{
			name:   "None",
			isNoOp: true,
		},
		{
			name: "One",
			input: []Validator[testConfig]{
				testNoOp,
			},
		},
		{
			name: "Three",
			input: []Validator[testConfig]{
				testNoOp, testNoOp, testNoOp,
			},
			wantsLen: 3,
		},
		{
			name: "WithNilsOneIsValid",
			input: []Validator[testConfig]{
				nil, nil, testNoOp, nil, nil,
			},
		},
		{
			name: "AllNil",
			input: []Validator[testConfig]{
				nil, nil, nil, nil, nil,
			},
			isNoOp: true,
		},
		{
			name: "ThreeWithNils",
			input: []Validator[testConfig]{
				testNoOp, nil, testNoOp, nil, testNoOp, nil,
			},
			wantsLen: 3,
		},
		{
			name: "ThreeWithAMultiValidator",
			input: []Validator[testConfig]{
				Join[testConfig](testNoOp, testNoOp, testNoOp),
				nil, testNoOp, nil, testNoOp, nil,
			},
			wantsLen: 5,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			validator := Join(testcase.input...)

			_ = validator.Validate(testConfig{})

			switch v := validator.(type) {
			case noOpValidator[testConfig]:
				require.True(t, testcase.isNoOp)
				return
			case multiValidator[testConfig]:
				require.Equal(t, testcase.wantsLen, len(v.validators))
				return
			}

			require.False(t, testcase.isNoOp)
		})
	}
}
