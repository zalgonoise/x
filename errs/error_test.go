package errs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSentinel(t *testing.T) {
	for _, testcase := range []struct {
		name   string
		domain Domain
		kind   Kind
		entity Entity
		wants  string
	}{
		{
			name: "Nil",
		},
		{
			name:   "DomainOnly",
			domain: "x/errs",
			wants:  "x/errs",
		},
		{
			name:  "KindOnly",
			kind:  "first",
			wants: "first",
		},
		{
			name:   "EntityOnly",
			entity: "test error",
			wants:  "test error",
		},
		{
			name:   "Sentinel",
			kind:   "first",
			entity: "test error",
			wants:  "first test error",
		},
		{
			name:   "SentinelWithDomain",
			domain: "x/errs",
			kind:   "first",
			entity: "test error",
			wants:  "x/errs: first test error",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			err := WithDomain(testcase.domain, testcase.kind, testcase.entity)

			if err == nil {
				require.Empty(t, testcase.wants)

				return
			}

			require.Equal(t, testcase.wants, err.Error())
		})
	}
}
