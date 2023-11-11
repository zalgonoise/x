package ptr_test

import (
	"errors"
	"io"
	"log/slog"
	"reflect"
	"testing"

	"github.com/zalgonoise/x/ptr"
)

type testDoer interface {
	Do() error
}

type testMetrics interface{}

type testService struct{}

func (s testService) Do() error { return nil }

type testServiceWithLogs struct {
	s      testDoer
	logger *slog.Logger
}

func (s testServiceWithLogs) Do() error { return nil }

type testServiceWithMetrics struct {
	s testDoer
	m testMetrics
}

func (s testServiceWithMetrics) Do() error { return nil }

func TestUnwrap(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		input testDoer
		wants testDoer
		err   error
	}{
		{
			name:  "TestService",
			input: testService{},
			wants: testService{},
		},
		{
			name: "WrappedOnce",
			input: testServiceWithLogs{
				s:      testService{},
				logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
			},
			wants: testService{},
		},
		{
			name: "WrappedTwice",
			input: testServiceWithMetrics{
				s: testServiceWithLogs{
					s:      testService{},
					logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
				},
				m: struct{}{},
			},
			wants: testService{},
		},
		{
			name:  "NilInput",
			input: nil,
			wants: testService{},
			err:   ptr.ErrNilInput,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			doer, err := ptr.Unwrap[testService](testcase.input)
			if !errors.Is(err, testcase.err) {
				t.Errorf("unexpected error: wanted %v ; got %v", testcase.err, err)
			}

			if !reflect.DeepEqual(testcase.wants, doer) {
				t.Errorf("output mismatch error: wanted %v ; got %v", testcase.wants, doer)
			}
		})
	}
}
