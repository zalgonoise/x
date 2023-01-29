package errors

import (
	"fmt"
	"testing"
)

func BenchmarkError(b *testing.B) {
	var (
		err1  = fmt.Errorf("error one")
		err2  = fmt.Errorf("error two")
		err3  = fmt.Errorf("error three")
		err4  = fmt.Errorf("error four")
		err5  = fmt.Errorf("error five")
		err6  = fmt.Errorf("error six")
		err7  = fmt.Errorf("error seven")
		err8  = fmt.Errorf("error eight")
		err9  = fmt.Errorf("error nine")
		err10 = fmt.Errorf("error ten")
		err11 = fmt.Errorf("error eleven")
		err12 = fmt.Errorf("error twelve")
		err13 = fmt.Errorf("error thirteen")
		err14 = fmt.Errorf("error fourteen")
		err15 = fmt.Errorf("error fifteen")
	)

	jerr := joinError{errs: []error{
		err1, err2, err3, err4, err5, err6, err7, err8,
		err9, err10, err11, err12, err13, err14, err15,
	}}

	var out string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out = jerr.Error()
	}
	_ = out

}

func BenchmarkJoin(b *testing.B) {
	var (
		err1  = fmt.Errorf("error one")
		err2  = fmt.Errorf("error two")
		err3  = fmt.Errorf("error three")
		err4  = fmt.Errorf("error four")
		err5  = fmt.Errorf("error five")
		err6  = fmt.Errorf("error six")
		err7  = fmt.Errorf("error seven")
		err8  = fmt.Errorf("error eight")
		err9  = fmt.Errorf("error nine")
		err10 = fmt.Errorf("error ten")
		err11 = fmt.Errorf("error eleven")
		err12 = fmt.Errorf("error twelve")
		err13 = fmt.Errorf("error thirteen")
		err14 = fmt.Errorf("error fourteen")
		err15 = fmt.Errorf("error fifteen")
	)

	var jerr error
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jerr = Join(
			err1, err2, nil, nil, err3, err4, err5, nil, err6, err7, nil, err8,
			err9, err10, nil, err11, nil, nil, err12, err13, err14, nil, err15,
		)
	}
	_ = jerr

}
