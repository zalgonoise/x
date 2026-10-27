package poller_test

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/zalgonoise/x/gacha/poller"
)

func TestConstant(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		d     time.Duration
		wants time.Duration
	}{
		{name: "Simple", d: time.Second, wants: time.Second},
		{name: "Zero", d: 0, wants: 0},
		{name: "NegativeIsZero", d: -time.Second, wants: 0},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			backoff := poller.Constant(testcase.d)

			for retry := range 5 {
				require.Equal(t, testcase.wants, backoff(retry+1))
			}
		})
	}
}

func TestExponential(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		base  time.Duration
		limit time.Duration
		wants []time.Duration
	}{
		{
			name:  "Simple",
			base:  time.Second,
			limit: 10 * time.Second,
			wants: []time.Duration{time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second, 10 * time.Second},
		},
		{
			name:  "LimitBelowBase",
			base:  time.Second,
			limit: time.Millisecond,
			wants: []time.Duration{time.Second, time.Second, time.Second},
		},
		{
			name:  "ZeroBase",
			base:  0,
			limit: time.Second,
			wants: []time.Duration{0, 0, 0},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			backoff := poller.Exponential(testcase.base, testcase.limit)

			for i, wants := range testcase.wants {
				require.Equal(t, wants, backoff(i+1), "retry %d", i+1)
			}
		})
	}

	t.Run("LargeRetryDoesNotOverflow", func(t *testing.T) {
		require.Equal(t, time.Hour, poller.Exponential(time.Second, time.Hour)(1000))
		require.Equal(t, time.Duration(math.MaxInt64), poller.Exponential(time.Second, math.MaxInt64)(1000))
	})
}

func TestJitter(t *testing.T) {
	const d = 10 * time.Second

	for _, testcase := range []struct {
		name     string
		factor   float64
		wantsMin time.Duration
	}{
		{name: "Simple", factor: 0.2, wantsMin: 8 * time.Second},
		{name: "None", factor: 0, wantsMin: d},
		{name: "NegativeIsNone", factor: -1, wantsMin: d},
		{name: "Full", factor: 1, wantsMin: 0},
		{name: "AboveOneIsFull", factor: 3, wantsMin: 0},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			backoff := poller.Jitter(poller.Constant(d), testcase.factor)

			for range 1000 {
				delay := backoff(1)
				require.GreaterOrEqual(t, delay, testcase.wantsMin)
				require.LessOrEqual(t, delay, d)
			}
		})
	}
}
