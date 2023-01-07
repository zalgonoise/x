package ptr_test

import (
	"errors"
	"reflect"
	"testing"

	. "github.com/zalgonoise/x/ptr"
)

func BenchmarkToArray(b *testing.B) {
	b.Run("PointerConversion256Elems", func(b *testing.B) {
		input := []int{
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
		}
		wants := [256]int{
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
		}

		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			res, err := ToArray(input)
			if err != nil {
				b.Errorf("unexpected error: %v", err)
				break
			}

			// validate results
			b.StopTimer()
			if !reflect.DeepEqual(wants, res) {
				b.Errorf("unexpected output error: wanted %v %T; got %v %T", wants, wants, res, res)
				return
			}
			b.StartTimer()
		}
	})

	b.Run("CopyElementsToArray256Elem", func(b *testing.B) {
		input := []int{
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
		}
		wants := [256]int{
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
		}

		var conv = func(source []int) (any, error) {
			const size = 256
			dest := [size]int{}
			copy(dest[:], source)
			return dest, nil
		}

		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			res, err := conv(input)
			if err != nil {
				b.Errorf("unexpected error: %v", err)
				break
			}

			// validate results
			b.StopTimer()
			if !reflect.DeepEqual(wants, res) {
				b.Errorf("unexpected output error: wanted %v %T; got %v %T", wants, wants, res, res)
				return
			}
			b.StartTimer()
		}
	})
}

func TestToArrayInts(t *testing.T) {
	t.Run("Cap1", func(t *testing.T) {
		input := []int{1}
		wants := [1]int{1}

		res, err := ToArray(input)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if !reflect.DeepEqual(wants, res) {
			t.Errorf("unexpected output error: wanted %v %T; got %v %T", wants, wants, res, res)
			return
		}
	})
	t.Run("Cap2", func(t *testing.T) {
		input := []int{1, 2}
		wants := [2]int{1, 2}

		res, err := ToArray(input)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if !reflect.DeepEqual(wants, res) {
			t.Errorf("unexpected output error: wanted %v %T; got %v %T", wants, wants, res, res)
			return
		}
	})
	t.Run("Cap4", func(t *testing.T) {
		input := []int{1, 2, 3, 4}
		wants := [4]int{1, 2, 3, 4}

		res, err := ToArray(input)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if !reflect.DeepEqual(wants, res) {
			t.Errorf("unexpected output error: wanted %v %T; got %v %T", wants, wants, res, res)
			return
		}
	})
	t.Run("Cap8", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5, 6, 7, 8}
		wants := [8]int{1, 2, 3, 4, 5, 6, 7, 8}

		res, err := ToArray(input)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if !reflect.DeepEqual(wants, res) {
			t.Errorf("unexpected output error: wanted %v %T; got %v %T", wants, wants, res, res)
			return
		}
	})

	t.Run("Error", func(t *testing.T) {
		input := []int{1, 2, 3}

		_, err := ToArray(input)
		if err == nil {
			t.Errorf("unexpectedly nil error")
			return
		}
		if !errors.Is(err, ErrInvalidCap) {
			t.Errorf("unexpected output error: wanted %v; got %v", ErrInvalidCap, err)
			return
		}
	})

}

func TestToArrayStrings(t *testing.T) {
	t.Run("Cap1", func(t *testing.T) {
		input := []string{"1"}
		wants := [1]string{"1"}

		res, err := ToArray(input)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if !reflect.DeepEqual(wants, res) {
			t.Errorf("unexpected output error: wanted %v %T; got %v %T", wants, wants, res, res)
			return
		}
	})
	t.Run("Cap2", func(t *testing.T) {
		input := []string{"1", "2"}
		wants := [2]string{"1", "2"}

		res, err := ToArray(input)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if !reflect.DeepEqual(wants, res) {
			t.Errorf("unexpected output error: wanted %v %T; got %v %T", wants, wants, res, res)
			return
		}
	})
	t.Run("Cap4", func(t *testing.T) {
		input := []string{"1", "2", "3", "4"}
		wants := [4]string{"1", "2", "3", "4"}

		res, err := ToArray(input)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if !reflect.DeepEqual(wants, res) {
			t.Errorf("unexpected output error: wanted %v %T; got %v %T", wants, wants, res, res)
			return
		}
	})
	t.Run("Cap8", func(t *testing.T) {
		input := []string{"1", "2", "3", "4", "5", "6", "7", "8"}
		wants := [8]string{"1", "2", "3", "4", "5", "6", "7", "8"}

		res, err := ToArray(input)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if !reflect.DeepEqual(wants, res) {
			t.Errorf("unexpected output error: wanted %v %T; got %v %T", wants, wants, res, res)
			return
		}
	})

	t.Run("Error", func(t *testing.T) {
		input := []string{"1", "2", "3"}

		_, err := ToArray(input)
		if err == nil {
			t.Errorf("unexpectedly nil error")
			return
		}
		if !errors.Is(err, ErrInvalidCap) {
			t.Errorf("unexpected output error: wanted %v; got %v", ErrInvalidCap, err)
			return
		}
	})

}
