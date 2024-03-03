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

func (noOp) IncServiceRegistries()                                                                  {}
func (noOp) IncServiceRegistryFailed()                                                              {}
func (noOp) ObserveServiceRegistryLatency(context.Context, time.Duration)                           {}
func (noOp) IncServiceCertsFetched(string)                                                          {}
func (noOp) IncServiceCertsFetchFailed(string)                                                      {}
func (noOp) ObserveServiceCertsFetchLatency(context.Context, string, time.Duration)                 {}
func (noOp) IncVerificationRequests(string)                                                         {}
func (noOp) IncVerificationFailed(string)                                                           {}
func (noOp) ObserveVerificationLatency(context.Context, string, time.Duration)                      {}
func (noOp) IncServiceDeletions()                                                                   {}
func (noOp) IncServiceDeletionFailed()                                                              {}
func (noOp) ObserveServiceDeletionLatency(context.Context, time.Duration)                           {}
func (noOp) IncPubKeyRequests()                                                                     {}
func (noOp) IncPubKeyRequestFailed()                                                                {}
func (noOp) ObservePubKeyRequestLatency(context.Context, time.Duration)                             {}
func (noOp) IncServiceLoginRequests(service string)                                                 {}
func (noOp) IncServiceLoginFailed(service string)                                                   {}
func (noOp) ObserveServiceLoginLatency(ctx context.Context, service string, duration time.Duration) {}
func (noOp) IncServiceTokenRequests(service string)                                                 {}
func (noOp) IncServiceTokenFailed(service string)                                                   {}
func (noOp) ObserveServiceTokenLatency(ctx context.Context, service string, duration time.Duration) {}
func (noOp) IncSchedulerNextCalls()                                                                 {}
func (noOp) IncExecutorExecCalls(string)                                                            {}
func (noOp) IncExecutorExecErrors(string)                                                           {}
func (noOp) ObserveExecLatency(context.Context, string, time.Duration)                              {}
func (noOp) IncExecutorNextCalls(string)                                                            {}
func (noOp) IncSelectorSelectCalls()                                                                {}
func (noOp) IncSelectorSelectErrors()                                                               {}
func (noOp) RegisterCollector(prometheus.Collector)                                                 {}
func (noOp) Registry() (*prometheus.Registry, error)                                                { return nil, nil }
