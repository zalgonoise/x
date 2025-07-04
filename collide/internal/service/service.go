package service

import (
	"context"
	"errors"
	"github.com/zalgonoise/x/collide/internal/repository/memory"
	pb "github.com/zalgonoise/x/collide/pkg/api/pb/collide/v1"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log/slog"
	"time"
)

type Repository interface {
	ListDistricts(ctx context.Context) ([]string, error)
	ListAllTracksByDistrict(ctx context.Context, district string) ([]string, error)
	ListDriftTracksByDistrict(ctx context.Context, district string) ([]string, error)
	GetAlternativesByDistrictAndTrack(ctx context.Context, district string, track string) ([]string, error)
	GetCollisionsByDistrictAndTrack(ctx context.Context, district string, track string) ([]string, error)
}

type Metrics interface {
	IncListDistricts()
	IncListDistrictsFailed()
	ObserveListDistrictsLatency(context.Context, time.Duration)
	IncListAllTracksByDistrict(string)
	IncListAllTracksByDistrictFailed(string)
	ObserveListAllTracksByDistrictLatency(context.Context, time.Duration, string)
	IncListDriftTracksByDistrict(string)
	IncListDriftTracksByDistrictFailed(string)
	ObserveListDriftTracksByDistrictLatency(context.Context, time.Duration, string)
	IncGetAlternativesByDistrictAndTrack(string, string)
	IncGetAlternativesByDistrictAndTrackFailed(string, string)
	ObserveGetAlternativesByDistrictAndTrackLatency(context.Context, time.Duration, string, string)
	IncGetCollisionsByDistrictAndTrack(string, string)
	IncGetCollisionsByDistrictAndTrackFailed(string, string)
	ObserveGetCollisionsByDistrictAndTrackLatency(context.Context, time.Duration, string, string)
}

type Service struct {
	pb.UnimplementedCollideServiceServer

	repo Repository

	metrics Metrics
	logger  *slog.Logger
	tracer  trace.Tracer
}

func New(repo Repository, metrics Metrics, logger *slog.Logger, tracer trace.Tracer) *Service {
	return &Service{
		repo:    repo,
		metrics: metrics,
		logger:  logger,
		tracer:  tracer,
	}
}

func (s *Service) ListDistricts(ctx context.Context, _ *pb.ListDistrictsRequest) (*pb.ListDistrictsResponse, error) {
	ctx, span := s.tracer.Start(ctx, "ListDistricts")
	defer span.End()

	s.metrics.IncListDistricts()
	start := time.Now()
	defer func() {
		s.metrics.ObserveListDistrictsLatency(ctx, time.Since(start))
	}()

	districts, err := s.repo.ListDistricts(ctx)

	switch {
	case errors.Is(err, memory.ErrNoDistricts):
		s.metrics.IncListDistrictsFailed()
		s.logger.ErrorContext(ctx, "listing districts got zero results", slog.String("error", err.Error()))
		span.RecordError(err)
		span.SetStatus(otelcodes.Error, err.Error())
		span.AddEvent("listing districts got zero results", trace.WithAttributes(
			attribute.String("error", err.Error())))

		return nil, status.Error(codes.NotFound, err.Error())
	case err != nil:
		s.metrics.IncListDistrictsFailed()
		s.logger.ErrorContext(ctx, "listing districts", slog.String("error", err.Error()))
		span.RecordError(err)
		span.SetStatus(otelcodes.Error, err.Error())
		span.AddEvent("listing districts", trace.WithAttributes(
			attribute.String("error", err.Error())))

		return nil, status.Error(codes.Internal, "internal server error")
	}

	s.logger.DebugContext(ctx, "listed districts successfully")
	span.AddEvent("listed districts successfully")
	span.SetStatus(otelcodes.Ok, "")

	return &pb.ListDistrictsResponse{Districts: districts}, nil
}

func (s *Service) ListAllTracksByDistrict(ctx context.Context, req *pb.ListAllTracksByDistrictRequest) (*pb.ListAllTracksByDistrictResponse, error) {
	ctx, span := s.tracer.Start(ctx, "ListAllTracksByDistrict", trace.WithAttributes(
		attribute.String("district", req.GetDistrict()),
		attribute.String("filter", "all")))
	defer span.End()

	district := req.GetDistrict()

	s.metrics.IncListAllTracksByDistrict(district)
	start := time.Now()
	defer func() {
		s.metrics.ObserveListAllTracksByDistrictLatency(ctx, time.Since(start), district)
	}()

	tracks, err := s.repo.ListAllTracksByDistrict(ctx, district)

	switch {
	case errors.Is(err, memory.ErrNoTracks), errors.Is(err, memory.ErrNoDistricts):
		s.metrics.IncListAllTracksByDistrictFailed(district)
		s.logger.ErrorContext(ctx, "listing all tracks by district got zero results",
			slog.String("error", err.Error()), slog.String("district", district))
		span.RecordError(err)
		span.SetStatus(otelcodes.Error, err.Error())
		span.AddEvent("listing all tracks by district got zero results", trace.WithAttributes(
			attribute.String("error", err.Error()),
			attribute.String("district", district)))

		return nil, status.Error(codes.NotFound, err.Error())
	case errors.Is(err, memory.ErrDistrictNotFound):
		s.metrics.IncListAllTracksByDistrictFailed(district)
		s.logger.ErrorContext(ctx, "listing all tracks in unknown district",
			slog.String("error", err.Error()), slog.String("district", district))
		span.RecordError(err)
		span.SetStatus(otelcodes.Error, err.Error())
		span.AddEvent("listing all tracks in unknown district", trace.WithAttributes(
			attribute.String("error", err.Error()),
			attribute.String("district", district)))

		return nil, status.Error(codes.InvalidArgument, err.Error())
	case err != nil:
		s.metrics.IncListAllTracksByDistrictFailed(district)
		s.logger.ErrorContext(ctx, "listing all tracks in district",
			slog.String("error", err.Error()), slog.String("district", district))
		span.RecordError(err)
		span.SetStatus(otelcodes.Error, err.Error())
		span.AddEvent("listing all tracks in district", trace.WithAttributes(
			attribute.String("error", err.Error()),
			attribute.String("district", district)))

		return nil, status.Error(codes.Internal, "internal server error")
	}

	s.logger.DebugContext(ctx, "listed all tracks by district successfully",
		slog.String("district", district), slog.Int("track_count", len(tracks)))
	span.AddEvent("listed all tracks by district successfully", trace.WithAttributes(
		attribute.String("district", district), attribute.Int("track_count", len(tracks))))
	span.SetStatus(otelcodes.Ok, "")

	return &pb.ListAllTracksByDistrictResponse{Tracks: tracks}, nil
}

func (s *Service) ListDriftTracksByDistrict(ctx context.Context, req *pb.ListDriftTracksByDistrictRequest) (*pb.ListDriftTracksByDistrictResponse, error) {
	ctx, span := s.tracer.Start(ctx, "ListDriftTracksByDistrict", trace.WithAttributes(
		attribute.String("district", req.GetDistrict()),
		attribute.String("filter", "drift")))

	defer span.End()

	district := req.GetDistrict()

	s.metrics.IncListDriftTracksByDistrict(district)
	start := time.Now()
	defer func() {
		s.metrics.ObserveListDriftTracksByDistrictLatency(ctx, time.Since(start), district)
	}()

	tracks, err := s.repo.ListDriftTracksByDistrict(ctx, district)

	switch {
	case errors.Is(err, memory.ErrNoTracks), errors.Is(err, memory.ErrNoDistricts):
		s.metrics.IncListDriftTracksByDistrictFailed(district)
		s.logger.ErrorContext(ctx, "listing drift tracks by district got zero results",
			slog.String("error", err.Error()), slog.String("district", district))
		span.RecordError(err)
		span.SetStatus(otelcodes.Error, err.Error())
		span.AddEvent("listing drift tracks by district got zero results", trace.WithAttributes(
			attribute.String("error", err.Error()),
			attribute.String("district", district)))

		return nil, status.Error(codes.NotFound, err.Error())
	case errors.Is(err, memory.ErrDistrictNotFound):
		s.metrics.IncListDriftTracksByDistrictFailed(district)
		s.logger.ErrorContext(ctx, "listing drift tracks in unknown district",
			slog.String("error", err.Error()), slog.String("district", district))
		span.RecordError(err)
		span.SetStatus(otelcodes.Error, err.Error())
		span.AddEvent("listing drift tracks in unknown district", trace.WithAttributes(
			attribute.String("error", err.Error()),
			attribute.String("district", district)))

		return nil, status.Error(codes.InvalidArgument, err.Error())
	case err != nil:
		s.metrics.IncListDriftTracksByDistrictFailed(district)
		s.logger.ErrorContext(ctx, "listing drift tracks in district",
			slog.String("error", err.Error()), slog.String("district", district))
		span.RecordError(err)
		span.SetStatus(otelcodes.Error, err.Error())
		span.AddEvent("listing drift tracks in district", trace.WithAttributes(
			attribute.String("error", err.Error()),
			attribute.String("district", district)))

		return nil, status.Error(codes.Internal, "internal server error")
	}

	s.logger.DebugContext(ctx, "listed drift tracks by district successfully",
		slog.String("district", district), slog.Int("track_count", len(tracks)))
	span.AddEvent("listed drift tracks by district successfully", trace.WithAttributes(
		attribute.String("district", district), attribute.Int("track_count", len(tracks))))
	span.SetStatus(otelcodes.Ok, "")

	return &pb.ListDriftTracksByDistrictResponse{Tracks: tracks}, nil
}

func (s *Service) GetAlternativesByDistrictAndTrack(ctx context.Context, req *pb.GetAlternativesByDistrictAndTrackRequest) (*pb.GetAlternativesByDistrictAndTrackResponse, error) {
	ctx, span := s.tracer.Start(ctx, "GetAlternativesByDistrictAndTrack", trace.WithAttributes(
		attribute.String("district", req.GetDistrict()),
		attribute.String("track", req.GetTrack())))
	defer span.End()

	district := req.GetDistrict()
	track := req.GetTrack()

	s.metrics.IncGetAlternativesByDistrictAndTrack(district, track)
	start := time.Now()
	defer func() {
		s.metrics.ObserveGetAlternativesByDistrictAndTrackLatency(ctx, time.Since(start), district, track)
	}()

	tracks, err := s.repo.GetAlternativesByDistrictAndTrack(ctx, district, track)

	switch {
	case errors.Is(err, memory.ErrNoTracks), errors.Is(err, memory.ErrNoDistricts):
		s.metrics.IncGetAlternativesByDistrictAndTrackFailed(district, track)
		s.logger.ErrorContext(ctx, "getting alternative tracks by track in district got zero results",
			slog.String("error", err.Error()), slog.String("district", district), slog.String("track", track))
		span.RecordError(err)
		span.SetStatus(otelcodes.Error, err.Error())
		span.AddEvent("getting alternative tracks by track in district got zero results", trace.WithAttributes(
			attribute.String("error", err.Error()),
			attribute.String("district", district),
			attribute.String("track", track)))

		return nil, status.Error(codes.NotFound, err.Error())
	case errors.Is(err, memory.ErrDistrictNotFound):
		s.metrics.IncGetAlternativesByDistrictAndTrackFailed(district, track)
		s.logger.ErrorContext(ctx, "getting alternative tracks by track in unknown district",
			slog.String("error", err.Error()), slog.String("district", district), slog.String("track", track))
		span.RecordError(err)
		span.SetStatus(otelcodes.Error, err.Error())
		span.AddEvent("getting alternative tracks by track in unknown district", trace.WithAttributes(
			attribute.String("error", err.Error()),
			attribute.String("district", district),
			attribute.String("track", track)))

		return nil, status.Error(codes.InvalidArgument, err.Error())
	case err != nil:
		s.metrics.IncGetAlternativesByDistrictAndTrackFailed(district, track)
		s.logger.ErrorContext(ctx, "getting alternative tracks by track in district",
			slog.String("error", err.Error()), slog.String("district", district), slog.String("track", track))
		span.RecordError(err)
		span.SetStatus(otelcodes.Error, err.Error())
		span.AddEvent("getting alternative tracks by track in district", trace.WithAttributes(
			attribute.String("error", err.Error()),
			attribute.String("district", district),
			attribute.String("track", track)))

		return nil, status.Error(codes.Internal, "internal server error")
	}

	s.logger.DebugContext(ctx, "fetched alternative tracks by track in district successfully",
		slog.String("district", district), slog.String("track", track), slog.Int("track_count", len(tracks)))
	span.AddEvent("fetched alternative tracks by track in district successfully", trace.WithAttributes(
		attribute.String("district", district),
		attribute.String("track", track),
		attribute.Int("track_count", len(tracks))))
	span.SetStatus(otelcodes.Ok, "")

	return &pb.GetAlternativesByDistrictAndTrackResponse{Tracks: tracks}, nil
}

func (s *Service) GetCollisionsByDistrictAndTrack(ctx context.Context, req *pb.GetCollisionsByDistrictAndTrackRequest) (*pb.GetCollisionsByDistrictAndTrackResponse, error) {
	ctx, span := s.tracer.Start(ctx, "GetCollisionsByDistrictAndTrack", trace.WithAttributes(
		attribute.String("district", req.GetDistrict()),
		attribute.String("track", req.GetTrack())))
	defer span.End()

	district := req.GetDistrict()
	track := req.GetTrack()

	s.metrics.IncGetCollisionsByDistrictAndTrack(district, track)
	start := time.Now()
	defer func() {
		s.metrics.ObserveGetCollisionsByDistrictAndTrackLatency(ctx, time.Since(start), district, track)
	}()

	tracks, err := s.repo.GetCollisionsByDistrictAndTrack(ctx, district, track)

	switch {
	case errors.Is(err, memory.ErrNoTracks), errors.Is(err, memory.ErrNoDistricts):
		s.metrics.IncGetCollisionsByDistrictAndTrackFailed(district, track)
		s.logger.ErrorContext(ctx, "getting colliding tracks by track in district got zero results",
			slog.String("error", err.Error()), slog.String("district", district), slog.String("track", track))
		span.RecordError(err)
		span.SetStatus(otelcodes.Error, err.Error())
		span.AddEvent("getting colliding tracks by track in district got zero results", trace.WithAttributes(
			attribute.String("error", err.Error()),
			attribute.String("district", district),
			attribute.String("track", track)))

		return nil, status.Error(codes.NotFound, err.Error())
	case errors.Is(err, memory.ErrDistrictNotFound):
		s.metrics.IncGetCollisionsByDistrictAndTrackFailed(district, track)
		s.logger.ErrorContext(ctx, "getting colliding tracks by track in unknown district",
			slog.String("error", err.Error()), slog.String("district", district), slog.String("track", track))
		span.RecordError(err)
		span.SetStatus(otelcodes.Error, err.Error())
		span.AddEvent("getting colliding tracks by track in unknown district", trace.WithAttributes(
			attribute.String("error", err.Error()),
			attribute.String("district", district),
			attribute.String("track", track)))

		return nil, status.Error(codes.InvalidArgument, err.Error())
	case err != nil:
		s.metrics.IncGetCollisionsByDistrictAndTrackFailed(district, track)
		s.logger.ErrorContext(ctx, "getting colliding tracks by track in district",
			slog.String("error", err.Error()), slog.String("district", district), slog.String("track", track))
		span.RecordError(err)
		span.SetStatus(otelcodes.Error, err.Error())
		span.AddEvent("getting colliding tracks by track in district", trace.WithAttributes(
			attribute.String("error", err.Error()),
			attribute.String("district", district),
			attribute.String("track", track)))

		return nil, status.Error(codes.Internal, "internal server error")
	}

	s.logger.DebugContext(ctx, "fetched colliding tracks by track in district successfully",
		slog.String("district", district), slog.String("track", track), slog.Int("track_count", len(tracks)))
	span.AddEvent("fetched colliding tracks by track in district successfully", trace.WithAttributes(
		attribute.String("district", district),
		attribute.String("track", track),
		attribute.Int("track_count", len(tracks))))
	span.SetStatus(otelcodes.Ok, "")

	return &pb.GetCollisionsByDistrictAndTrackResponse{Tracks: tracks}, nil
}
