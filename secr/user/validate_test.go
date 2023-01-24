package user_test

import (
	"errors"
	"testing"

	. "github.com/zalgonoise/x/secr/user"
)

func BenchmarkValidatePassword(b *testing.B) {
	input := "L0ng_4nD-C0mP!3X^P@$$W0RD+!?L0ng_4nD-C0mP!3X^P@$$W0RD+!?"
	var err error

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = ValidatePassword(input)
	}
	_ = err
}

func TestValidatePassword(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Simple", func(t *testing.T) {
			input := "SecretPassword"
			err := ValidatePassword(input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
		})
	})

	t.Run("Success", func(t *testing.T) {
		t.Run("Special", func(t *testing.T) {
			input := "Secret0!["
			err := ValidatePassword(input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
		})
	})

	t.Run("Fail", func(t *testing.T) {
		t.Run("Empty", func(t *testing.T) {
			input := ""
			err := ValidatePassword(input)
			if !errors.Is(ErrEmptyPassword, err) {
				t.Errorf("unexpected error: wanted %v ; got %v", ErrEmptyPassword, err)
				return
			}
		})
		t.Run("TooShort", func(t *testing.T) {
			input := "x"
			err := ValidatePassword(input)
			if !errors.Is(ErrShortPassword, err) {
				t.Errorf("unexpected error: wanted %v ; got %v", ErrShortPassword, err)
				return
			}
		})
		t.Run("TooLong", func(t *testing.T) {
			// 301 chars
			input := "1111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111"
			err := ValidatePassword(input)
			if !errors.Is(ErrLongPassword, err) {
				t.Errorf("unexpected error: wanted %v ; got %v", ErrLongPassword, err)
				return
			}
		})
		t.Run("InvalidSpecial", func(t *testing.T) {
			input := "Special Password"
			err := ValidatePassword(input)
			if !errors.Is(ErrInvalidPassword, err) {
				t.Errorf("unexpected error: wanted %v ; got %v", ErrInvalidPassword, err)
				return
			}
		})

		t.Run("InvalidRepeat", func(t *testing.T) {
			input := "Special0000"
			err := ValidatePassword(input)
			if !errors.Is(ErrInvalidPassword, err) {
				t.Errorf("unexpected error: wanted %v ; got %v", ErrInvalidPassword, err)
				return
			}
		})
	})
}
