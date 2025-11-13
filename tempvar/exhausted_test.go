package tempvar

import (
	"testing"

	"github.com/zalgonoise/x/is"
)

func TestExhausted_Value(t *testing.T) {
	type user struct {
		name string
		id   int
	}

	for _, testcase := range []struct {
		name  string
		data  user
		limit uint64
	}{
		{
			name: "ValueAndExpiry",
			data: user{
				name: "Gopher",
				id:   1,
			},
			limit: 3,
		},
		{
			name: "InvalidLimit",
			data: user{
				name: "Go",
				id:   2,
			},
			limit: 0,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			v := NewExhaustedVar(testcase.data, testcase.limit)

			limit := testcase.limit
			if limit < minLimit {
				limit = minLimit
			}

			values := make([]*user, 0, int(limit))
			for i := 0; i < int(limit); i++ {
				values = append(values, v.Value())
			}

			expired := v.Value()

			for i := range values {
				is.Equal(t, *values[i], testcase.data)
			}

			is.Empty(t, expired)
		})
	}
}
