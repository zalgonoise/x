package cfg_test

import (
	"testing"
	"time"

	"github.com/zalgonoise/x/cfg"
)

type testConfig struct {
	target string
	dur    time.Duration
	ratio  float64
}

func withTarget(target string) cfg.Option[testConfig] {
	return cfg.Register(func(cfg testConfig) testConfig {
		cfg.target = target

		return cfg
	})
}

func withDuration(dur time.Duration) cfg.Option[testConfig] {
	return cfg.Register(func(cfg testConfig) testConfig {
		cfg.dur = dur

		return cfg
	})
}

func withRatio(ratio float64) cfg.Option[testConfig] {
	return cfg.Register(func(cfg testConfig) testConfig {
		cfg.ratio = ratio

		return cfg
	})
}

func TestNew(t *testing.T) {
	for _, testcase := range []struct {
		name    string
		options []cfg.Option[testConfig]
		wants   testConfig
	}{
		{
			name:  "ZeroOptions",
			wants: testConfig{},
		},
		{
			name: "OneOption",
			options: []cfg.Option[testConfig]{
				withDuration(time.Minute),
			},
			wants: testConfig{
				target: "",
				dur:    time.Minute,
				ratio:  0,
			},
		},
		{
			name: "AllOptions",
			options: []cfg.Option[testConfig]{
				withTarget("someTarget"),
				withDuration(time.Hour),
				withRatio(0.5),
			},
			wants: testConfig{
				target: "someTarget",
				dur:    time.Hour,
				ratio:  0.5,
			},
		},
		{
			name: "WithOverrides",
			options: []cfg.Option[testConfig]{
				withTarget("someTarget"),
				withDuration(time.Hour),
				withRatio(0.5),
				withTarget("otherTarget"),
				withRatio(0.7),
				withTarget("lastTarget"),
			},
			wants: testConfig{
				target: "lastTarget",
				dur:    time.Hour,
				ratio:  0.7,
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			config := cfg.New[testConfig](testcase.options...)

			mustMatch(t, testcase.wants, config)
		})
	}
}

func TestNoOp(t *testing.T) {
	config := cfg.New[testConfig](cfg.NoOp[testConfig]{})

	mustMatch(t, testConfig{}, config)
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
