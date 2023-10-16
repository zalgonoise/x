package selector

import (
	"context"
)

type Metrics interface {
	IncSelectorSelectCalls()
	IncSelectorSelectErrors()
}

type SelectorWithMetrics struct {
	s Selector
	m Metrics
}

func (s SelectorWithMetrics) Next(ctx context.Context) error {
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

	if withMetrics, ok := s.(SelectorWithMetrics); ok {
		withMetrics.m = m

		return withMetrics
	}

	return SelectorWithMetrics{
		s: s,
		m: m,
	}
}
