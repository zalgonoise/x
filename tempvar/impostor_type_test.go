package tempvar

import (
	"errors"
	"testing"

	"github.com/zalgonoise/x/is"
)

func TestImpostor(t *testing.T) {
	type user struct {
		name string
		id   int
	}

	u := user{
		name: "Gopher",
		id:   1,
	}

	for _, testcase := range []struct {
		name   string
		max    uint
		thresh uint
		err    error
		wants  *user
	}{
		{
			name:   "Success/MostLikelyPass",
			max:    1_000_000_000,
			thresh: 1,
			wants:  &u,
		},
		{
			name:   "Success/MostLikelyFail",
			max:    1_000_000_000,
			thresh: 999_999_999,
		},
		{
			name:   "Fail/MaxIsZero",
			max:    0,
			thresh: 1,
			err:    ErrMaxMustNotBeZero,
		},
		{
			name:   "Fail/ThresholdOverflow",
			max:    1,
			thresh: 1,
			err:    ErrThresholdOverflow,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			v, err := NewImpostorType(u, testcase.max, testcase.thresh)
			if err != nil {
				is.True(t, errors.Is(err, testcase.err))

				return
			}

			is.NilError(t, err)
			is.True(t, v != nil)

			value := v.Value()

			if testcase.wants != nil {
				v := (*user)(value)
				is.EqualValue(t, v, testcase.wants)
			}
		})
	}
}

func TestImpostorFunc(t *testing.T) {
	for _, testcase := range []struct {
		name string
	}{
		{name: "Coverage"},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			_ = impostorFunc()
		})
	}
}

func TestImpostorInterface_Value(t *testing.T) {
	for _, testcase := range []struct {
		name string
	}{
		{name: "Coverage"},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			_ = impostorInterface.Value()
		})
	}
}
