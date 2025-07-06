package metrics

import (
	"context"
	"time"
)

func NoOp() noOp {
	return noOp{}
}

type noOp struct{}

func (noOp) IncListDistricts(context.Context)                                               {}
func (noOp) IncListDistrictsFailed(context.Context)                                         {}
func (noOp) ObserveListDistrictsLatency(context.Context, time.Duration)                     {}
func (noOp) IncListAllTracksByDistrict(context.Context, string)                             {}
func (noOp) IncListAllTracksByDistrictFailed(context.Context, string)                       {}
func (noOp) ObserveListAllTracksByDistrictLatency(context.Context, time.Duration, string)   {}
func (noOp) IncListDriftTracksByDistrict(context.Context, string)                           {}
func (noOp) IncListDriftTracksByDistrictFailed(context.Context, string)                     {}
func (noOp) ObserveListDriftTracksByDistrictLatency(context.Context, time.Duration, string) {}
func (noOp) IncGetAlternativesByDistrictAndTrack(context.Context, string, string)           {}
func (noOp) IncGetAlternativesByDistrictAndTrackFailed(context.Context, string, string)     {}
func (noOp) ObserveGetAlternativesByDistrictAndTrackLatency(context.Context, time.Duration, string, string) {
}
func (noOp) IncGetCollisionsByDistrictAndTrack(context.Context, string, string)       {}
func (noOp) IncGetCollisionsByDistrictAndTrackFailed(context.Context, string, string) {}
func (noOp) ObserveGetCollisionsByDistrictAndTrackLatency(context.Context, time.Duration, string, string) {
}
