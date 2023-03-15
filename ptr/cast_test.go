package ptr

import (
	"reflect"
	"testing"
)

func TestCast(t *testing.T) {
	t.Run(
		"SuccessStringToBytes", func(t *testing.T) {
			input := "test"
			bytes := Cast[[]byte](input)
			if string(bytes) != input {
				t.Errorf("output mismatch error: wanted %s ; got %s", input, string(bytes))
			}
		},
	)

	t.Run(
		"SuccessBytesToIntegers", func(t *testing.T) {
			input := []byte("test")
			ints := Cast[[]uint8](input)

			for idx := range input {
				if ints[idx] != input[idx] {
					t.Errorf("output mismatch error on index %d: wanted %v ; got %v", idx, input[idx], ints[idx])
				}
			}
		},
	)
	t.Run(
		"SuccessTypeToType", func(t *testing.T) {
			type inputType struct {
				id      int8
				num     int32
				name    string
				age     int
				quality float64
				skills  []string
				streak  []float64
			}

			type outputType struct {
				Id      int8      `json:"id,omitempty"`
				Number  int32     `json:"number,omitempty"`
				Name    string    `json:"name,omitempty"`
				Age     int       `json:"age,omitempty"`
				Quality float64   `json:"quality,omitempty"`
				Skills  []string  `json:"skills,omitempty"`
				Streak  []float64 `json:"streak,omitempty"`
			}

			input := inputType{
				id:      8,
				num:     32,
				name:    "jackson",
				age:     99,
				quality: 1.5,
				skills: []string{
					"committed",
					"hard-working",
					"interested",
				},
				streak: []float64{
					0.5, 0.8, 1.3, 1.9, 2.5,
				},
			}

			wants := outputType{
				Id:      8,
				Number:  32,
				Name:    "jackson",
				Age:     99,
				Quality: 1.5,
				Skills: []string{
					"committed",
					"hard-working",
					"interested",
				},
				Streak: []float64{
					0.5, 0.8, 1.3, 1.9, 2.5,
				},
			}

			output := Cast[outputType, inputType](input)
			if !reflect.DeepEqual(wants, output) {
				t.Errorf("output mismatch error: wanted %v ; got %v", wants, output)
			}
		},
	)
}

func TestCastPtr(t *testing.T) {
	t.Run(
		"SuccessStringToBytes", func(t *testing.T) {
			input := "test"
			bytes := CastPtr[[]byte](&input)

			if bytes == nil {
				t.Error("unexpected nil output")
				return
			}

			if string(*bytes) != input {
				t.Errorf("output mismatch error: wanted %s ; got %s", input, string(*bytes))
			}
		},
	)

	t.Run(
		"SuccessBytesToIntegers", func(t *testing.T) {
			input := []byte("test")
			intsPtr := CastPtr[[]uint8](&input)

			if intsPtr == nil {
				t.Error("unexpected nil output")
				return
			}
			ints := *intsPtr

			for idx := range input {
				if ints[idx] != input[idx] {
					t.Errorf("output mismatch error on index %d: wanted %v ; got %v", idx, input[idx], ints[idx])
				}
			}
		},
	)
	t.Run(
		"SuccessTypeToType", func(t *testing.T) {
			type inputType struct {
				id      int8
				num     int32
				name    string
				age     int
				quality float64
				skills  []string
				streak  []float64
			}

			type outputType struct {
				Id      int8      `json:"id,omitempty"`
				Number  int32     `json:"number,omitempty"`
				Name    string    `json:"name,omitempty"`
				Age     int       `json:"age,omitempty"`
				Quality float64   `json:"quality,omitempty"`
				Skills  []string  `json:"skills,omitempty"`
				Streak  []float64 `json:"streak,omitempty"`
			}

			input := inputType{
				id:      8,
				num:     32,
				name:    "jackson",
				age:     99,
				quality: 1.5,
				skills: []string{
					"committed",
					"hard-working",
					"interested",
				},
				streak: []float64{
					0.5, 0.8, 1.3, 1.9, 2.5,
				},
			}

			wants := outputType{
				Id:      8,
				Number:  32,
				Name:    "jackson",
				Age:     99,
				Quality: 1.5,
				Skills: []string{
					"committed",
					"hard-working",
					"interested",
				},
				Streak: []float64{
					0.5, 0.8, 1.3, 1.9, 2.5,
				},
			}

			output := CastPtr[outputType, inputType](&input)
			if !reflect.DeepEqual(wants, *output) {
				t.Errorf("output mismatch error: wanted %v ; got %v", wants, *output)
			}
		},
	)
}
