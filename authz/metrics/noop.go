package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func NoOp() noOp {
	return noOp{}
}

type noOp struct{}

func (noOp) IncServiceRegistries()                                                  {}
func (noOp) IncServiceRegistryFailed()                                              {}
func (noOp) ObserveServiceRegistryLatency(context.Context, time.Duration)           {}
func (noOp) IncServiceCertsFetched(string)                                          {}
func (noOp) IncServiceCertsFetchFailed(string)                                      {}
func (noOp) ObserveServiceCertsFetchLatency(context.Context, string, time.Duration) {}
func (noOp) IncServiceDeletions()                                                   {}
func (noOp) IncServiceDeletionFailed()                                              {}
func (noOp) ObserveServiceDeletionLatency(context.Context, time.Duration)           {}
func (noOp) IncPubKeyRequests()                                                     {}
func (noOp) IncPubKeyRequestFailed()                                                {}
func (noOp) ObservePubKeyRequestLatency(context.Context, time.Duration)             {}
func (noOp) RegisterCollector(prometheus.Collector)                                 {}
func (noOp) Registry() (*prometheus.Registry, error)                                { return nil, nil }