package errs

import (
	"errors"
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	errNotFound := errors.New("can't find user gopher in the database")
	errDatabase := errors.New("query returned zero results")

	errDomain := Domain("x/errs/testing")
	errInvalid := Kind("invalid")
	errEmpty := Kind("empty")
	errSubject := Entity("subject")
	errObject := Entity("object")
	errInvalidSubject := New("", errInvalid, errSubject)
	errEmptyObjectWithDomain := New(errDomain, errEmpty, errObject)

	for _, testcase := range []struct {
		name        string
		domain      Domain
		kind        Kind
		entity      Entity
		args        []any
		wants       error
		innerString string
		wantsString string
	}{
		{
			name:   "Success/Simple",
			kind:   "invalid",
			entity: "user",
			wants: Error{
				kind:   "invalid",
				entity: "user",
			},
			wantsString: "invalid user",
		},
		{
			name:   "Success/Simple/StringArgs",
			kind:   "invalid",
			entity: "user",
			args:   []any{errNotFound.Error(), errDatabase.Error()},
			wants: Error{
				kind:   "invalid",
				entity: "user",
			},
			innerString: fmt.Sprintf("%s: %s", errNotFound.Error(), errDatabase.Error()),
			wantsString: "invalid user: can't find user gopher in the database: query returned zero results",
		},
		{
			name:   "Success/Simple/RegularErrorArgs",
			kind:   "invalid",
			entity: "user",
			args:   []any{errNotFound, errDatabase},
			wants: Error{
				kind:   "invalid",
				entity: "user",
			},
			innerString: fmt.Sprintf("%s: %s", errNotFound.Error(), errDatabase.Error()),
			wantsString: "invalid user: can't find user gopher in the database: query returned zero results",
		},
		{
			name:   "Success/Simple/ErrorArgs",
			kind:   "invalid",
			entity: "user",
			args:   []any{errInvalidSubject, errEmptyObjectWithDomain},
			wants: Error{
				kind:   "invalid",
				entity: "user",
			},
			innerString: fmt.Sprintf("%s %s: %s: %s %s", errInvalid, errSubject, errDomain, errEmpty, errObject),
			wantsString: "invalid user: invalid subject: x/errs/testing: empty object",
		},
		{
			name:   "Success/Simple/ErrorArgsMixed",
			kind:   "invalid",
			entity: "user",
			args: []any{
				"fatal error", errInvalidSubject,
				errEmptyObjectWithDomain,
				errEmptyObjectWithDomain,
				nil, nil, errInvalidSubject,
			},
			wants: Error{
				kind:   "invalid",
				entity: "user",
			},
			innerString: fmt.Sprintf("%s: %s %s: %s: %s %s: %s: %s %s: %s %s",
				"fatal error",
				errInvalid, errSubject,
				errDomain, errEmpty, errObject,
				errDomain, errEmpty, errObject,
				errInvalid, errSubject,
			),
			wantsString: "invalid user: fatal error: invalid subject: x/errs/testing: empty object: x/errs/testing: empty object: invalid subject",
		},
		{
			name:   "Success/WithDomain",
			domain: "x/errs",
			kind:   "invalid",
			entity: "user",
			wants: ErrorWithDomain{
				domain: "x/errs",
				err: Error{
					kind:   "invalid",
					entity: "user",
				},
			},
			wantsString: "x/errs: invalid user",
		},
		{
			name:   "Success/WithDomain/StringArgs",
			domain: "x/errs",
			kind:   "invalid",
			entity: "user",
			args:   []any{errNotFound.Error(), errDatabase.Error()},
			wants: ErrorWithDomain{
				domain: "x/errs",
				err: Error{
					kind:   "invalid",
					entity: "user",
				},
			},
			innerString: fmt.Sprintf("%s: %s", errNotFound.Error(), errDatabase.Error()),
			wantsString: "x/errs: invalid user: can't find user gopher in the database: query returned zero results",
		},
		{
			name:   "Success/WithDomain/RegularErrorArgs",
			domain: "x/errs",
			kind:   "invalid",
			entity: "user",
			args:   []any{errNotFound, errDatabase},
			wants: ErrorWithDomain{
				domain: "x/errs",
				err: Error{
					kind:   "invalid",
					entity: "user",
				},
			},
			innerString: fmt.Sprintf("%s: %s", errNotFound.Error(), errDatabase.Error()),
			wantsString: "x/errs: invalid user: can't find user gopher in the database: query returned zero results",
		},
		{
			name:   "Success/WithDomain/ErrorArgs",
			domain: "x/errs",
			kind:   "invalid",
			entity: "user",
			args:   []any{errInvalidSubject, errEmptyObjectWithDomain},
			wants: ErrorWithDomain{
				domain: "x/errs",
				err: Error{
					kind:   "invalid",
					entity: "user",
				},
			},
			innerString: fmt.Sprintf("%s %s: %s: %s %s", errInvalid, errSubject, errDomain, errEmpty, errObject),
			wantsString: "x/errs: invalid user: invalid subject: x/errs/testing: empty object",
		},
		{
			name:   "Fail/NoKindOrEntity",
			domain: "x/errs",
		},
		{
			name:        "Fail/OnlyKind",
			kind:        "invalid",
			wantsString: errInvalid.Error(),
		},
		{
			name:        "Fail/OnlyEntity",
			entity:      "subject",
			wantsString: errSubject.Error(),
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			err := New(testcase.domain, testcase.kind, testcase.entity, testcase.args...)

			switch e := err.(type) {
			case nil:
			case ErrorWithDomain:
				we, ok := testcase.wants.(ErrorWithDomain)
				mustMatch(t, true, ok)
				mustMatch(t, we.domain, e.domain)
				mustMatch(t, we.err.kind, e.err.kind)
				mustMatch(t, we.err.entity, e.err.entity)

				var innerString string
				if e.Unwrap() != nil {
					innerString = e.Unwrap().Error()
				}

				mustMatch(t, testcase.innerString, innerString)

			case Error:
				we, ok := testcase.wants.(Error)
				mustMatch(t, true, ok)
				mustMatch(t, we.kind, e.kind)
				mustMatch(t, we.entity, e.entity)

				var innerString string
				if e.Unwrap() != nil {
					innerString = e.Unwrap().Error()
				}

				mustMatch(t, testcase.innerString, innerString)
			}

			if err != nil {
				mustMatch(t, testcase.wantsString, err.Error())
			} else {
				mustMatch(t, nil, testcase.wants)
			}
		})
	}
}

func TestFormat(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		format    string
		domain    Domain
		kind      Kind
		entity    Entity
		unchanged bool
	}{

		{
			name:   "Simple/NoFormat",
			kind:   "invalid",
			entity: "subject",
		},
		{
			name:   "Simple/WithFormat",
			format: "[%s] %s %s",
			kind:   "invalid",
			entity: "subject",
		},
		{
			name:   "WithDomain/NoFormat",
			domain: "x/errs",
			kind:   "invalid",
			entity: "subject",
		},
		{
			name:   "WithDomain/WithFormat",
			domain: "x/errs",
			format: "[%s] %s %s",
			kind:   "invalid",
			entity: "subject",
		},
		{
			name:      "Invalid/NoKindOrEntity",
			domain:    "x/errs",
			format:    "[%s] %s %s",
			unchanged: true,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			err := Format(testcase.domain, testcase.kind, testcase.entity, testcase.format, nil)

			switch e := err.(type) {
			case Error:
				mustMatch(t, testcase.format, e.format)
			case ErrorWithDomain:
				mustMatch(t, testcase.format, e.err.format)
			default:
				mustMatch(t, true, testcase.unchanged)
			}
		})
	}
}

// mustMatch is an over-simplification of a testify/require.Equal() call, or a
// reflect.DeepEqual call; but leverages the generics in Go and the comparable type constraint.
//
// It is able to evaluate the types defined in the testConfig data structure, and should be replaced
// only in case it is no longer suitable. For the moment it evaluates the entire data structure as a
// drop-in replacement of testify/require.Equal; but it could also be used to evaluate on a
// field-by-field approach.
func mustMatch[T comparable](t *testing.T, wants, got T) {
	if wants != got {
		t.Errorf("output mismatch error: wanted %v -- got %v", wants, got)

		return
	}

	t.Logf("item matched value %v", wants)
}
