package cron

import "context"

type CronMetrics interface {
	IsUp(bool)
}

type CronWithMetrics struct {
	r Runtime
	m CronMetrics
}

func (c CronWithMetrics) Run(ctx context.Context) {
	c.m.IsUp(true)
	c.r.Run(ctx)
	c.m.IsUp(false)
}

func (c CronWithMetrics) Err() <-chan error {
	return c.r.Err()
}

func cronWithMetrics(r Runtime, m CronMetrics) Runtime {
	if r == nil {
		return noOpRuntime{}
	}

	if m == nil {
		return r
	}

	return CronWithMetrics{
		r: r,
		m: m,
	}
}
