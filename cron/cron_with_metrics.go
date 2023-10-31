package cron

import "context"

type Metrics interface {
	IsUp(bool)
}

type withMetrics struct {
	r Runtime
	m Metrics
}

func (c withMetrics) Run(ctx context.Context) {
	c.m.IsUp(true)
	c.r.Run(ctx)
	c.m.IsUp(false)
}

func (c withMetrics) Err() <-chan error {
	return c.r.Err()
}

func cronWithMetrics(r Runtime, m Metrics) Runtime {
	if r == nil {
		return noOpRuntime{}
	}

	if m == nil {
		return r
	}

	if metrics, ok := r.(withMetrics); ok {
		metrics.m = m

		return metrics
	}

	return withMetrics{
		r: r,
		m: m,
	}
}
