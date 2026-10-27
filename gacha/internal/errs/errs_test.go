package errs

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestError(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		domain    Domain
		kind      Kind
		entity    Entity
		wantsText string
		matchErr  error
		mismatch  error
	}{
		{
			name:      "Simple/CustomDomainKindEntity",
			domain:    Domain("gacha/memory"),
			kind:      Kind("invalid"),
			entity:    Entity("lease"),
			wantsText: "gacha/memory: invalid lease",
			matchErr:  Kind("invalid"),
			mismatch:  Kind("other"),
		},
		{
			name:      "EmptyDomainUsesDefault",
			domain:    Domain(""),
			kind:      Kind("negative"),
			entity:    Entity("offset"),
			wantsText: "gacha: negative offset",
			matchErr:  Entity("offset"),
			mismatch:  Entity("other"),
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			err := New(testcase.domain, testcase.kind, testcase.entity)
			require.Equal(t, testcase.wantsText, err.Error())
			require.ErrorIs(t, err, err)
			require.ErrorIs(t, err, testcase.matchErr)
			require.NotErrorIs(t, err, testcase.mismatch)

			var typedErr Error
			require.True(t, errors.As(err, &typedErr))
			require.NotNil(t, typedErr.Unwrap())
			require.NotEmpty(t, typedErr.domain.Error())
			require.Equal(t, string(testcase.kind), typedErr.kind.Error())
			require.Equal(t, string(testcase.entity), typedErr.entity.Error())
		})
	}
}
