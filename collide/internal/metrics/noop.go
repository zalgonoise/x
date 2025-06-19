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

func (noOp) IncListDistricts()                                                              {}
func (noOp) IncListDistrictsFailed()                                                        {}
func (noOp) ObserveListDistrictsLatency(context.Context, time.Duration)                     {}
func (noOp) IncListAllTracksByDistrict(string)                                              {}
func (noOp) IncListAllTracksByDistrictFailed(string)                                        {}
func (noOp) ObserveListAllTracksByDistrictLatency(context.Context, time.Duration, string)   {}
func (noOp) IncListDriftTracksByDistrict(string)                                            {}
func (noOp) IncListDriftTracksByDistrictFailed(string)                                      {}
func (noOp) ObserveListDriftTracksByDistrictLatency(context.Context, time.Duration, string) {}
func (noOp) IncGetAlternativesByDistrictAndTrack(string, string)                            {}
func (noOp) IncGetAlternativesByDistrictAndTrackFailed(string, string)                      {}
func (noOp) ObserveGetAlternativesByDistrictAndTrackLatency(context.Context, time.Duration, string, string) {
}
func (noOp) IncGetCollisionsByDistrictAndTrack(string, string)       {}
func (noOp) IncGetCollisionsByDistrictAndTrackFailed(string, string) {}
func (noOp) ObserveGetCollisionsByDistrictAndTrackLatency(context.Context, time.Duration, string, string) {
}
func (noOp) RegisterCollector(prometheus.Collector)  {}
func (noOp) Registry() (*prometheus.Registry, error) { return prometheus.NewRegistry(), nil }
