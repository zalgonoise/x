package service

import (
	"context"
	"time"
)

type Metrics interface {
	IncRequestsReceived(ctx context.Context)
	IncRequestsFailed(ctx context.Context)
	ObserveHandlingLatency(ctx context.Context, dur time.Duration)
}

var _ Service = HandlerWithMetrics{}

type HandlerWithMetrics struct {
	s       Service
	metrics Metrics
}

func (h HandlerWithMetrics) Handle(ctx context.Context, value int) (err error) {
	start := time.Now()
	h.metrics.IncRequestsReceived(ctx)

	if err = h.s.Handle(ctx, value); err != nil {
		h.metrics.IncRequestsFailed(ctx)
	}

	h.metrics.ObserveHandlingLatency(ctx, time.Since(start))

	return err
}

func WithMetrics(s Service, metrics Metrics) HandlerWithMetrics {
	return HandlerWithMetrics{
		s:       s,
		metrics: metrics,
	}
}
