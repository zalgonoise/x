package fts

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIndex_SearchStrings(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		attrs []Attribute[int, string]
		query string
		wants []Attribute[int, string]
		err   error
	}{
		{
			name: "Success/OneResult",
			attrs: []Attribute[int, string]{
				{Key: 1, Value: "some data"},
				{Key: 2, Value: "struck gold"},
				{Key: 3, Value: "some kind of copper"},
				{Key: 4, Value: "probably bronze"},
			},
			query: "gold",
			wants: []Attribute[int, string]{
				{Key: 2, Value: "struck gold"},
			},
		},
		{
			name: "Success/ThreeResults",
			attrs: []Attribute[int, string]{
				{Key: 1, Value: "some data"},
				{Key: 2, Value: "struck gold"},
				{Key: 3, Value: "some kind of copper"},
				{Key: 4, Value: "probably bronze"},
				{Key: 5, Value: "just chips"},
				{Key: 6, Value: "good ol' gold plate"},
				{Key: 7, Value: "gol-- gol-- gold!!"},
			},
			query: "gold",
			wants: []Attribute[int, string]{
				{Key: 2, Value: "struck gold"},
				{Key: 6, Value: "good ol' gold plate"},
				{Key: 7, Value: "gol-- gol-- gold!!"},
			},
		},
		{
			name: "Success/ThreeResultsWithExpression",
			attrs: []Attribute[int, string]{
				{Key: 1, Value: "some data"},
				{Key: 2, Value: "struck gold"},
				{Key: 3, Value: "some kind of copper"},
				{Key: 4, Value: "probably bronze"},
				{Key: 5, Value: "just chips"},
				{Key: 6, Value: "good ol' golden plate"},
				{Key: 7, Value: "gol-- gol-- gold!!"},
			},
			query: "gold*",
			wants: []Attribute[int, string]{
				{Key: 2, Value: "struck gold"},
				{Key: 6, Value: "good ol' golden plate"},
				{Key: 7, Value: "gol-- gol-- gold!!"},
			},
		},
		{
			name: "Fail/NoResults",
			attrs: []Attribute[int, string]{
				{Key: 1, Value: "some data"},
				{Key: 3, Value: "some kind of copper"},
				{Key: 4, Value: "probably bronze"},
				{Key: 5, Value: "just chips"},
			},
			query: "gold",
			err:   ErrNotFoundKeyword,
		},
		{
			name:  "Fail/NoInput",
			attrs: []Attribute[int, string]{},
			query: "gold",
			err:   ErrZeroAttributes,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			index, err := NewIndex("", testcase.attrs...)
			if err != nil {
				require.ErrorIs(t, err, testcase.err)

				return
			}

			res, err := index.Search(context.Background(), testcase.query)
			if err != nil {
				require.ErrorIs(t, err, testcase.err)

				return
			}

			ids := make([]int, 0, len(res))
			for i := range res {
				ids = append(ids, res[i].Key)
			}

			require.NoError(t, index.Delete(context.Background(), ids...))

			require.Equal(t, testcase.wants, res)
			require.NoError(t, index.Shutdown(context.Background()))
		})
	}
}

func TestIndex_SearchBytes(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		attrs []Attribute[int, []byte]
		query []byte
		wants []Attribute[int, []byte]
		err   error
	}{
		{
			name: "Success/OneResult",
			attrs: []Attribute[int, []byte]{
				{Key: 1, Value: []byte("some data")},
				{Key: 2, Value: []byte("struck gold")},
				{Key: 3, Value: []byte("some kind of copper")},
				{Key: 4, Value: []byte("probably bronze")},
			},
			query: []byte("gold"),
			wants: []Attribute[int, []byte]{
				{Key: 2, Value: []byte("struck gold")},
			},
		},
		{
			name: "Success/ThreeResults",
			attrs: []Attribute[int, []byte]{
				{Key: 1, Value: []byte("some data")},
				{Key: 2, Value: []byte("struck gold")},
				{Key: 3, Value: []byte("some kind of copper")},
				{Key: 4, Value: []byte("probably bronze")},
				{Key: 5, Value: []byte("just chips")},
				{Key: 6, Value: []byte("good ol' gold plate")},
				{Key: 7, Value: []byte("gol-- gol-- gold!!")},
			},
			query: []byte("gold"),
			wants: []Attribute[int, []byte]{
				{Key: 2, Value: []byte("struck gold")},
				{Key: 6, Value: []byte("good ol' gold plate")},
				{Key: 7, Value: []byte("gol-- gol-- gold!!")},
			},
		},
		{
			name: "Success/ThreeResultsWithExpression",
			attrs: []Attribute[int, []byte]{
				{Key: 1, Value: []byte("some data")},
				{Key: 2, Value: []byte("struck gold")},
				{Key: 3, Value: []byte("some kind of copper")},
				{Key: 4, Value: []byte("probably bronze")},
				{Key: 5, Value: []byte("just chips")},
				{Key: 6, Value: []byte("good ol' golden plate")},
				{Key: 7, Value: []byte("gol-- gol-- gold!!")},
			},
			query: []byte("gold*"),
			wants: []Attribute[int, []byte]{
				{Key: 2, Value: []byte("struck gold")},
				{Key: 6, Value: []byte("good ol' golden plate")},
				{Key: 7, Value: []byte("gol-- gol-- gold!!")},
			},
		},
		{
			name: "Fail/NoResults",
			attrs: []Attribute[int, []byte]{
				{Key: 1, Value: []byte("some data")},
				{Key: 3, Value: []byte("some kind of copper")},
				{Key: 4, Value: []byte("probably bronze")},
				{Key: 5, Value: []byte("just chips")},
			},
			query: []byte("gold"),
			err:   ErrNotFoundKeyword,
		},
		{
			name:  "Fail/NoInput",
			attrs: []Attribute[int, []byte]{},
			query: []byte("gold"),
			err:   ErrZeroAttributes,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			index, err := NewIndex[int, []byte]("", testcase.attrs...)
			if err != nil {
				require.ErrorIs(t, err, testcase.err)

				return
			}

			res, err := index.Search(context.Background(), testcase.query)
			if err != nil {
				require.ErrorIs(t, err, testcase.err)

				return
			}

			ids := make([]int, 0, len(res))
			for i := range res {
				ids = append(ids, res[i].Key)
			}

			require.NoError(t, index.Delete(context.Background(), ids...))

			require.Equal(t, testcase.wants, res)
			require.NoError(t, index.Shutdown(context.Background()))
		})
	}
}
