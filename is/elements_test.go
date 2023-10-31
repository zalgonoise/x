package is

import "testing"

func TestContains(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		for _, testcase := range []struct {
			name  string
			slice []string
			item  string
			fails bool
		}{
			{
				name:  "Success",
				slice: []string{"a", "b", "c"},
				item:  "a",
			},
			{
				name:  "Fail",
				slice: []string{"a", "b", "c"},
				item:  "d",
				fails: true,
			},
		} {
			t.Run(testcase.name, func(t *testing.T) {
				tt := &testT{}

				Contains(tt, testcase.item, testcase.slice)

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
			name  string
			slice []person
			item  person
			fails bool
		}{
			{
				name: "Success",
				slice: []person{
					{"a", 1},
					{"b", 2},
					{"c", 3},
				},
				item: person{"b", 2},
			},
			{
				name: "Fail/Mismatch/All",
				slice: []person{
					{"a", 1},
					{"b", 2},
					{"c", 3},
				},
				item:  person{"d", 4},
				fails: true,
			},
			{
				name: "Fail/Mismatch/ID",
				slice: []person{
					{"a", 1},
					{"b", 2},
					{"c", 3},
				},
				item:  person{"a", 4},
				fails: true,
			},
			{
				name: "Fail/Mismatch/Name",
				slice: []person{
					{"a", 1},
					{"b", 2},
					{"c", 3},
				},
				item:  person{"d", 1},
				fails: true,
			},
		} {
			t.Run(testcase.name, func(t *testing.T) {
				tt := &testT{}

				Contains(tt, testcase.item, testcase.slice)

				if tt.failCount > 0 && !testcase.fails {
					t.Logf(tt.last)
					t.Fail()
				}
			})
		}
	})
}

func TestElementsMatch(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		for _, testcase := range []struct {
			name     string
			actual   []string
			expected []string
			fails    bool
		}{
			{
				name:     "Success/MatchingOrder",
				actual:   []string{"a", "b", "c"},
				expected: []string{"a", "b", "c"},
			},
			{
				name:     "Success/MismatchingOrder",
				actual:   []string{"a", "b", "c"},
				expected: []string{"c", "a", "b"},
			},
			{
				name:     "Fail/MissingItem",
				actual:   []string{"a", "b", "c"},
				expected: []string{"a", "b"},
				fails:    true,
			},
			{
				name:     "Fail/MismatchingItem",
				actual:   []string{"a", "b", "c"},
				expected: []string{"a", "d", "b"},
				fails:    true,
			},
		} {
			t.Run(testcase.name, func(t *testing.T) {
				tt := &testT{}

				ElementsMatch(tt, testcase.expected, testcase.actual)

				if tt.failCount > 0 && !testcase.fails {
					t.Logf(tt.last)
					t.Fail()
				}
			})
		}
	})
}

func TestEqualElements(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		for _, testcase := range []struct {
			name     string
			actual   []string
			expected []string
			fails    bool
		}{
			{
				name:     "Success/MatchingOrder",
				actual:   []string{"a", "b", "c"},
				expected: []string{"a", "b", "c"},
			},
			{
				name:     "Fail/MismatchingOrder",
				actual:   []string{"a", "b", "c"},
				expected: []string{"c", "a", "b"},
				fails:    true,
			},
			{
				name:     "Fail/MissingItem",
				actual:   []string{"a", "b", "c"},
				expected: []string{"a", "b"},
				fails:    true,
			},
			{
				name:     "Fail/MismatchingItem",
				actual:   []string{"a", "b", "c"},
				expected: []string{"a", "d", "b"},
				fails:    true,
			},
		} {
			t.Run(testcase.name, func(t *testing.T) {
				tt := &testT{}

				EqualElements(tt, testcase.expected, testcase.actual)

				if tt.failCount > 0 && !testcase.fails {
					t.Logf(tt.last)
					t.Fail()
				}
			})
		}
	})

	t.Run("ByteSlice", func(t *testing.T) {
		for _, testcase := range []struct {
			name     string
			expected []byte
			actual   []byte
			fails    bool
		}{
			{
				name:     "Success/Matching",
				expected: []byte("some content"),
				actual:   []byte("some content"),
			},
			{
				name:     "Fail/Mismatching",
				expected: []byte("some content"),
				actual:   []byte("other content"),
				fails:    true,
			},
			{
				name:     "Fails/Empty",
				expected: []byte("some content"),
				actual:   []byte{},
				fails:    true,
			},
			{
				name:     "Fails/Nil",
				expected: []byte("some content"),
				actual:   nil,
				fails:    true,
			},
		} {
			t.Run(testcase.name, func(t *testing.T) {
				tt := &testT{}

				EqualElements(tt, testcase.expected, testcase.actual)

				if tt.failCount > 0 && !testcase.fails {
					t.Logf(tt.last)
					t.Fail()
				}
			})
		}
	})
}
