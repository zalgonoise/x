package ptr_test

import (
	"errors"
	"reflect"
	"testing"

	. "github.com/zalgonoise/x/ptr"
)

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
