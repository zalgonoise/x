package mapping

import (
	"cmp"
	"testing"

	"github.com/zalgonoise/cfg"
)

func TestFieldSet(t *testing.T) {
	key1 := "alpha"
	value1 := "A-alpha"
	zero := "zero"

	for _, testcase := range []struct {
		name string
		zero *string
	}{
		{
			name: "WithZero",
			zero: &zero,
		},
		{
			name: "NilZero",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			t.Run("Table", func(t *testing.T) {
				m := make(map[string]*string, 2)

				table := NewTable(m, WithZero[string](testcase.zero))

				value, ok := table.Get(key1)

				isEqual(t, false, ok)
				isEqual(t, testcase.zero, value)

				added := table.Set(key1, &value1)

				isEqual(t, true, added)

				value, ok = table.Get(key1)

				isEqual(t, true, ok)
				isEqual(t, &value1, value)
			})

			t.Run("Index", func(t *testing.T) {
				m := make(map[string]*string, 2)

				index := NewIndex(m,
					WithZero[string](testcase.zero),
					WithIndex[*string](cmp.Compare[string]),
				)

				value, ok := index.Get(key1)

				isEqual(t, false, ok)
				isEqual(t, testcase.zero, value)

				added := index.Set(key1, &value1)

				isEqual(t, true, added)

				value, ok = index.Get(key1)

				isEqual(t, true, ok)
				isEqual(t, &value1, value)
			})
		})
	}
}

func TestFieldGet(t *testing.T) {
	key1 := "alpha"
	value1 := "A-alpha"

	key2 := "beta"
	value2 := "B-beta"

	key3 := "gamma"

	zero := "zero"

	m := map[string]*string{
		key1: &value1,
		key2: &value2,
	}

	for _, testcase := range []struct {
		name    string
		m       map[string]*string
		zero    *string
		indexed bool
		cmpFunc func(a, b string) int
		input   string
		wants   *string
		ok      bool
	}{
		{
			name:  "WithValue/NilZero",
			m:     m,
			zero:  nil,
			input: key1,
			ok:    true,
			wants: &value1,
		},
		{
			name:  "WithValue/SetZero",
			m:     m,
			zero:  &zero,
			input: key2,
			ok:    true,
			wants: &value2,
		},
		{
			name:  "WithoutValue/NilZero",
			m:     m,
			zero:  nil,
			input: key3,
			ok:    false,
			wants: nil,
		},
		{
			name:  "WithoutValue/SetZero",
			m:     m,
			zero:  &zero,
			input: key3,
			ok:    false,
			wants: &zero,
		},
		{
			name:  "EmptyKey/NilZero",
			m:     m,
			zero:  nil,
			input: "",
			ok:    false,
			wants: nil,
		},
		{
			name:  "EmptyKey/SetZero",
			m:     m,
			zero:  &zero,
			input: "",
			ok:    false,
			wants: &zero,
		},
		{
			name:  "EmptyMap/NilZero",
			m:     map[string]*string{},
			zero:  nil,
			input: "",
			ok:    false,
			wants: nil,
		},
		{
			name:  "EmptyMap/SetZero",
			m:     map[string]*string{},
			zero:  &zero,
			input: "",
			ok:    false,
			wants: &zero,
		},
		{
			name:    "Indexed/Unordered/NilZero",
			m:       m,
			zero:    nil,
			indexed: true,
			input:   key1,
			ok:      true,
			wants:   &value1,
		},
		{
			name:    "Indexed/Ordered/NilZero",
			m:       m,
			zero:    nil,
			indexed: true,
			cmpFunc: cmp.Compare[string],
			input:   key1,
			ok:      true,
			wants:   &value1,
		},
		{
			name:    "Indexed/Unordered/SetZero",
			m:       m,
			zero:    &zero,
			indexed: true,
			input:   key1,
			ok:      true,
			wants:   &value1,
		},
		{
			name:    "Indexed/Ordered/SetZero",
			m:       m,
			zero:    &zero,
			indexed: true,
			cmpFunc: cmp.Compare[string],
			input:   key1,
			ok:      true,
			wants:   &value1,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			opts := []cfg.Option[Config[string, *string]]{
				WithZero[string](testcase.zero),
			}

			if testcase.indexed {
				opts = append(opts, WithIndex[*string](testcase.cmpFunc))
			}

			field := New(testcase.m, opts...)

			result, ok := field.Get(testcase.input)
			isEqual(t, testcase.ok, ok)
			isEqual(t, testcase.wants, result)
		})
	}
}

func TestEmbeded(t *testing.T) {
	zero := ""
	name := "gopher"
	id := "#0000"

	namePriorities := map[string]*string{
		"main": nil,
		"sec":  &name,
	}

	idPriorities := map[string]*string{
		"main": nil,
		"sec":  &zero,
		"tetr": &id,
	}

	nameIndex := NewIndex(namePriorities,
		WithIndex[*string](cmp.Compare[string]),
	)

	idIndex := NewIndex(idPriorities,
		WithZero[string](&zero),
		WithIndex[*string](cmp.Compare[string]),
	)

	fields := map[string]*Index[string, *string]{
		"name": nameIndex,
		"id":   idIndex,
	}

	index := New(fields)

	// access name
	indexedName, ok := index.Get("name")

	isEqual(t, true, ok)

	for _, key := range indexedName.Keys {
		value, ok := indexedName.Get(key)

		if ok && value != nil {
			isEqual(t, "sec", key)
			isEqual(t, name, *value)
		}
	}

	// access id
	indexedID, ok := index.Get("id")

	isEqual(t, true, ok)

	for _, key := range indexedID.Keys {
		value, ok := indexedID.Get(key)

		if ok && value != nil && *value != "" {
			isEqual(t, "tetr", key)
			isEqual(t, id, *value)
		}
	}
}

func isEqual[T comparable](t *testing.T, wants, got T) {
	if got != wants {
		t.Errorf("output mismatch error: wanted %v ; got %v", wants, got)
		t.Fail()

		return
	}

	t.Logf("output matched expected value: %v", wants)
}
