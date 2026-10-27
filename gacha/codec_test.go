package gacha

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

type order struct {
	ID    int      `json:"id"`
	Items []string `json:"items"`
}

func TestJSON(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		item  order
		wants string
	}{
		{name: "Simple", item: order{ID: 1, Items: []string{"a", "b"}}, wants: `{"id":1,"items":["a","b"]}`},
		{name: "ZeroValue", item: order{}, wants: `{"id":0,"items":null}`},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			codec := JSON[order]()

			data, err := codec.Marshal(testcase.item)
			require.NoError(t, err)
			require.JSONEq(t, testcase.wants, string(data))

			item, err := codec.Unmarshal(data)
			require.NoError(t, err)
			require.Equal(t, testcase.item, item)
		})
	}

	t.Run("UnmarshalError", func(t *testing.T) {
		_, err := JSON[order]().Unmarshal([]byte("{"))
		require.Error(t, err)
	})

	t.Run("MarshalError", func(t *testing.T) {
		_, err := JSON[chan int]().Marshal(make(chan int))
		require.Error(t, err)
	})
}

func TestBytes(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		item  []byte
		wants []byte
	}{
		{name: "Simple", item: []byte("gacha"), wants: []byte("gacha")},
		{name: "Empty", item: []byte{}, wants: []byte{}},
		{name: "NilBecomesEmpty", item: nil, wants: []byte{}},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			codec := Bytes()

			data, err := codec.Marshal(testcase.item)
			require.NoError(t, err)
			require.NotNil(t, data)
			require.Equal(t, testcase.wants, data)

			if len(testcase.item) > 0 {
				// Marshal copies: mutating the input leaves the marshaled data untouched.
				testcase.item[0] = 'x'
				require.Equal(t, testcase.wants, data)
			}

			item, err := codec.Unmarshal(data)
			require.NoError(t, err)
			require.Equal(t, testcase.wants, item)
		})
	}
}

func TestString(t *testing.T) {
	for _, testcase := range []struct {
		name string
		item string
	}{
		{name: "Simple", item: "gacha"},
		{name: "Empty", item: ""},
		{name: "Unicode", item: "ガチャ"},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			codec := String()

			data, err := codec.Marshal(testcase.item)
			require.NoError(t, err)
			require.NotNil(t, data)
			require.Equal(t, []byte(testcase.item), data)

			item, err := codec.Unmarshal(data)
			require.NoError(t, err)
			require.Equal(t, testcase.item, item)
		})
	}
}

func TestIdentity(t *testing.T) {
	item := order{ID: 7, Items: []string{"x"}}
	codec := Identity[order]()

	data, err := codec.Marshal(item)
	require.NoError(t, err)
	require.Equal(t, item, data)

	out, err := codec.Unmarshal(data)
	require.NoError(t, err)
	require.Equal(t, item, out)
}

func TestNewCodec(t *testing.T) {
	errParse := errors.New("parse failure")

	codec := NewCodec[int, string](
		MarshalerFunc[int, string](func(item int) (string, error) {
			return strconv.Itoa(item), nil
		}),
		UnmarshalerFunc[int, string](func(data string) (int, error) {
			n, err := strconv.Atoi(data)
			if err != nil {
				return 0, errors.Join(errParse, err)
			}

			return n, nil
		}),
	)

	for _, testcase := range []struct {
		name  string
		data  string
		wants int
		err   error
	}{
		{name: "Simple", data: "42", wants: 42},
		{name: "Negative", data: "-1", wants: -1},
		{name: "Invalid", data: "x", err: errParse},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			item, err := codec.Unmarshal(testcase.data)
			if testcase.err != nil {
				require.ErrorIs(t, err, testcase.err)

				return
			}

			require.NoError(t, err)
			require.Equal(t, testcase.wants, item)

			data, err := codec.Marshal(item)
			require.NoError(t, err)
			require.Equal(t, testcase.data, data)
		})
	}
}

func BenchmarkJSON(b *testing.B) {
	codec := JSON[order]()
	item := order{ID: 1, Items: []string{"a", "b", "c"}}

	b.ReportAllocs()

	for b.Loop() {
		data, err := codec.Marshal(item)
		if err != nil {
			b.Fatal(err)
		}

		if _, err = codec.Unmarshal(data); err != nil {
			b.Fatal(err)
		}
	}
}
