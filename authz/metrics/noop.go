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

func (noOp) IncServiceRegistries()                                                         {}
func (noOp) IncServiceRegistryFailed()                                                     {}
func (noOp) ObserveServiceRegistryLatency(context.Context, time.Duration)                  {}
func (noOp) IncServiceDeletions()                                                          {}
func (noOp) IncServiceDeletionFailed()                                                     {}
func (noOp) ObserveServiceDeletionLatency(context.Context, time.Duration)                  {}
func (noOp) IncCertificatesCreated(string)                                                 {}
func (noOp) IncCertificatesCreateFailed(string)                                            {}
func (noOp) ObserveCertificatesCreateLatency(context.Context, string, time.Duration)       {}
func (noOp) IncCertificatesListed(string)                                                  {}
func (noOp) IncCertificatesListFailed(string)                                              {}
func (noOp) ObserveCertificatesListLatency(context.Context, string, time.Duration)         {}
func (noOp) IncCertificatesDeleted(string)                                                 {}
func (noOp) IncCertificatesDeleteFailed(string)                                            {}
func (noOp) ObserveCertificatesDeleteLatency(context.Context, string, time.Duration)       {}
func (noOp) IncCertificatesVerified(string)                                                {}
func (noOp) IncCertificateVerificationFailed(string)                                       {}
func (noOp) ObserveCertificateVerificationLatency(context.Context, string, time.Duration)  {}
func (noOp) IncRootCertificateRequests()                                                   {}
func (noOp) IncRootCertificateRequestFailed()                                              {}
func (noOp) ObserveRootCertificateRequestLatency(context.Context, time.Duration)           {}
func (noOp) IncServiceLoginRequests(string)                                                {}
func (noOp) IncServiceLoginFailed(string)                                                  {}
func (noOp) ObserveServiceLoginLatency(context.Context, string, time.Duration)             {}
func (noOp) IncServiceTokenRequests(string)                                                {}
func (noOp) IncServiceTokenFailed(string)                                                  {}
func (noOp) ObserveServiceTokenLatency(context.Context, string, time.Duration)             {}
func (noOp) IncServiceTokenVerifications(string)                                           {}
func (noOp) IncServiceTokenVerificationFailed(string)                                      {}
func (noOp) ObserveServiceTokenVerificationLatency(context.Context, string, time.Duration) {}
func (noOp) IncSchedulerNextCalls()                                                        {}
func (noOp) IncExecutorExecCalls(string)                                                   {}
func (noOp) IncExecutorExecErrors(string)                                                  {}
func (noOp) ObserveExecLatency(context.Context, string, time.Duration)                     {}
func (noOp) IncExecutorNextCalls(string)                                                   {}
func (noOp) IncSelectorSelectCalls()                                                       {}
func (noOp) IncSelectorSelectErrors()                                                      {}
func (noOp) RegisterCollector(prometheus.Collector)                                        {}
func (noOp) Registry() (*prometheus.Registry, error)                                       { return prometheus.NewRegistry(), nil }
