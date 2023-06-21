package ptr

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type _stringerImpl struct{ value string }

func (s _stringerImpl) String() string {
	return s.value
}

func BenchmarkIsNil(b *testing.B) {
	b.Run("nil interface checking", func(b *testing.B) {
		b.Run("StandardLibrary", func(b *testing.B) {
			var (
				ok    bool
				iface fmt.Stringer
			)

			for i := 0; i < b.N; i++ {
				ok = iface == nil
			}
			_ = ok
		})

		b.Run("PtrLibrary", func(b *testing.B) {
			var (
				ok    bool
				iface fmt.Stringer
			)

			for i := 0; i < b.N; i++ {
				ok = IsNil(iface)
			}
			_ = ok
		})

	})

	b.Run("interface w/ nil value checking", func(b *testing.B) {
		b.Run("StandardLibrary", func(b *testing.B) {
			var (
				ok    bool
				iface = fmt.Stringer(nil)
			)

			for i := 0; i < b.N; i++ {
				ok = iface == nil
			}
			_ = ok
		})

		b.Run("PtrLibrary", func(b *testing.B) {
			var (
				ok    bool
				iface = fmt.Stringer(nil)
			)

			for i := 0; i < b.N; i++ {
				ok = IsNil(iface)
			}
			_ = ok
		})

	})
}

func TestIsNil(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		value any
		isNil bool
	}{
		{
			name:  "WithValue/String",
			value: "test",
		},
		{
			name:  "WithValue/Float32",
			value: (float32)(1.5),
		},
		{
			name:  "WithValue/map[string]string",
			value: map[string]string{},
		},
		{
			name:  "Nil/map[string]string",
			value: (map[string]string)(nil),
			isNil: true,
		},
		{
			name:  "WithValue/chan string",
			value: make(chan string),
		},
		{
			name:  "Nil/chan string",
			value: (chan string)(nil),
			isNil: true,
		},
		{
			name:  "WithValue/slice string",
			value: make([]string, 0),
		},
		// test fails for slices since they are composed
		// with a different structure
		//{
		//	name:  "Nil/slice string",
		//	value: ([]string)(nil),
		//	isNil: true,
		//},
		{
			name:  "WithValue/custom type",
			value: struct{ name string }{},
		},
		{
			name:  "WithValue/custom type pointer",
			value: &struct{ name string }{},
		},
		{
			name:  "Nil/custom type pointer",
			value: (*_stringerImpl)(nil),
			isNil: true,
		},
		{
			name:  "WithValue/interface",
			value: fmt.Stringer(_stringerImpl{value: "OK"}),
		},
		{
			name:  "Nil/interface",
			value: fmt.Stringer(nil),
			isNil: true,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			require.Equal(t, testcase.isNil, IsNil(testcase.value))
		})
	}
}

func TestGetInterface(t *testing.T) {
	t.Run("fmt.Stringer", func(t *testing.T) {
		// use iface below as a breakpoint to explore the underlying data types
		iface := GetInterface(fmt.Stringer(_stringerImpl{value: "OK"}))

		require.NotEmpty(t, iface.Itab.Hash)
	})
}

type Stringer interface {
	String() string
}

func TestIsEqual(t *testing.T) {
	impl := &_stringerImpl{value: "OK"}

	s1 := fmt.Stringer(impl)
	s2 := Stringer(impl)

	require.True(t, IsEqual(s1, s2))
}

func TestMatch(t *testing.T) {
	impl := &_stringerImpl{value: "OK"}

	s1 := fmt.Stringer(impl)
	s2 := Stringer(impl)

	require.True(t, Match(s1, s2))
}

func TestGetUncommon(t *testing.T) {
	impl := &_stringerImpl{value: "OK"}

	s1 := Stringer(impl)
	require.NotNil(t, GetUncommon(s1))
}
