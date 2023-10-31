package selector

import (
	"context"
)

type Metrics interface {
	IncSelectorSelectCalls()
	IncSelectorSelectErrors()
}

type withMetrics struct {
	s Selector
	m Metrics
}

func (s withMetrics) Next(ctx context.Context) error {
	s.m.IncSelectorSelectCalls()

	if err := s.s.Next(ctx); err != nil {
		s.m.IncSelectorSelectErrors()

		return err
	}

	return nil
}

func selectorWithMetrics(s Selector, m Metrics) Selector {
	if s == nil {
		return noOpSelector{}
	}

	if m == nil {
		return s
	}

	if metrics, ok := s.(withMetrics); ok {
		metrics.m = m

		return metrics
	}

	return withMetrics{
		s: s,
		m: m,
	}
}
