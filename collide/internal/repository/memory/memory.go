package memory

import (
	"context"
	"errors"
	"fmt"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"log/slog"
	"slices"

	"go.opentelemetry.io/otel/trace"

	"github.com/zalgonoise/x/collide/internal/tracks"
)

var (
	ErrDistrictNotFound = errors.New("district not found")
	ErrNoDistricts      = errors.New("no districts found")
	ErrNoTracks         = errors.New("no tracks found")
)

type Repository struct {
	tracks *tracks.TrackList

	logger *slog.Logger
	tracer trace.Tracer
}

func New(logger *slog.Logger, tracer trace.Tracer) *Repository {
	return &Repository{logger: logger, tracer: tracer}
}

func (r *Repository) FromBytes(buf []byte) error {
	r.tracks = &tracks.TrackList{}

	if _, err := r.tracks.Read(buf); err != nil {
		return err
	}

	return nil
}

func (r *Repository) ListDistricts(ctx context.Context) ([]string, error) {
	ctx, span := r.tracer.Start(ctx, "Repository.ListDistricts", trace.WithAttributes(
		attribute.Int("track_count", len(r.tracks.Tracks))))
	defer span.End()

	districts, err := tracks.GetDistricts(r.tracks)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelcodes.Error, err.Error())
		span.AddEvent("listing districts", trace.WithAttributes(
			attribute.String("error", err.Error())))

		return nil, err
	}

	if len(districts) == 0 {
		span.RecordError(ErrNoDistricts)
		span.SetStatus(otelcodes.Error, ErrNoDistricts.Error())
		span.AddEvent("listing districts", trace.WithAttributes(
			attribute.String("error", ErrNoDistricts.Error())))

		return nil, ErrNoDistricts
	}

	span.AddEvent("listed districts successfully", trace.WithAttributes(
		attribute.Int("district_count", len(districts))))
	span.SetStatus(otelcodes.Ok, "")

	return districts, nil
}

func (r *Repository) ListAllTracksByDistrict(ctx context.Context, district string) ([]string, error) {
	ctx, span := r.tracer.Start(ctx, "Repository.ListAllTracksByDistrict", trace.WithAttributes(
		attribute.Int("track_count", len(r.tracks.Tracks)),
		attribute.String("district", district),
		attribute.Bool("drift_only", false)))
	defer span.End()

	t, err := tracks.GetTracksByDistrict(r.tracks, district, false)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelcodes.Error, err.Error())
		span.AddEvent("fetching all tracks by district", trace.WithAttributes(
			attribute.String("error", err.Error())))

		return nil, err
	}

	if len(t) == 0 {
		span.AddEvent("fetching all tracks by district yielded zero results", trace.WithAttributes(
			attribute.String("district", district),
			attribute.Int("track_count", 0)))

		districts, err := r.ListDistricts(ctx)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelcodes.Error, err.Error())
			span.AddEvent("failed to fetch district list for verification", trace.WithAttributes(
				attribute.String("district", district),
				attribute.String("error", err.Error())))

			return nil, fmt.Errorf("listing all tracks by district %q: %w", district, err)
		}

		if !slices.Contains(districts, district) {
			span.RecordError(ErrDistrictNotFound)
			span.SetStatus(otelcodes.Error, ErrDistrictNotFound.Error())
			span.AddEvent("district does not exist", trace.WithAttributes(
				attribute.String("target_district", district),
				attribute.StringSlice("districts", districts),
				attribute.String("error", ErrDistrictNotFound.Error())))

			return nil, ErrDistrictNotFound
		}

		span.RecordError(ErrNoTracks)
		span.SetStatus(otelcodes.Error, ErrNoTracks.Error())
		span.AddEvent("district does not contain any tracks", trace.WithAttributes(
			attribute.String("district", district),
			attribute.String("error", ErrNoTracks.Error())))

		return nil, ErrNoTracks
	}

	span.AddEvent("listed tracks in district successfully", trace.WithAttributes(
		attribute.String("district", district),
		attribute.Int("track_count", len(t))))
	span.SetStatus(otelcodes.Ok, "")

	return tracks.GetNamesFromIDs(r.tracks, t)
}

func (r *Repository) ListDriftTracksByDistrict(ctx context.Context, district string) ([]string, error) {
	ctx, span := r.tracer.Start(ctx, "Repository.ListDriftTracksByDistrict", trace.WithAttributes(
		attribute.Int("track_count", len(r.tracks.Tracks)),
		attribute.String("district", district),
		attribute.Bool("drift_only", true)))
	defer span.End()

	t, err := tracks.GetTracksByDistrict(r.tracks, district, true)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelcodes.Error, err.Error())
		span.AddEvent("fetching drift tracks by district", trace.WithAttributes(
			attribute.String("error", err.Error())))

		return nil, err
	}

	if len(t) == 0 {
		span.AddEvent("fetching drift tracks by district yielded zero results", trace.WithAttributes(
			attribute.String("district", district),
			attribute.Int("track_count", 0)))

		districts, err := r.ListDistricts(ctx)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelcodes.Error, err.Error())
			span.AddEvent("failed to fetch district list for verification", trace.WithAttributes(
				attribute.String("district", district),
				attribute.String("error", err.Error())))

			return nil, fmt.Errorf("listing drift tracks by district %q: %w", district, err)
		}

		if !slices.Contains(districts, district) {
			span.RecordError(ErrDistrictNotFound)
			span.SetStatus(otelcodes.Error, ErrDistrictNotFound.Error())
			span.AddEvent("district does not exist", trace.WithAttributes(
				attribute.String("target_district", district),
				attribute.StringSlice("districts", districts),
				attribute.String("error", ErrDistrictNotFound.Error())))

			return nil, ErrDistrictNotFound
		}

		span.AddEvent("listed drift tracks in district successfully", trace.WithAttributes(
			attribute.String("district", district),
			attribute.Int("track_count", len(t))))
		span.SetStatus(otelcodes.Ok, "")

		return nil, ErrNoTracks
	}

	span.AddEvent("listed drift tracks in district successfully", trace.WithAttributes(
		attribute.String("district", district),
		attribute.Int("track_count", len(t))))
	span.SetStatus(otelcodes.Ok, "")

	return tracks.GetNamesFromIDs(r.tracks, t)
}

func (r *Repository) GetAlternativesByDistrictAndTrack(ctx context.Context, district, track string) ([]string, error) {
	ctx, span := r.tracer.Start(ctx, "Repository.GetAlternativesByDistrictAndTrack", trace.WithAttributes(
		attribute.Int("track_count", len(r.tracks.Tracks)),
		attribute.String("district", district),
		attribute.String("track", track)))
	defer span.End()

	trackID, err := tracks.GetIDFromName(r.tracks, track)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelcodes.Error, err.Error())
		span.AddEvent("getting track ID from name", trace.WithAttributes(
			attribute.String("track", track),
			attribute.String("error", err.Error())))

		return nil, err
	}

	span.AddEvent("found track ID", trace.WithAttributes(
		attribute.String("track", track),
		attribute.String("track_id", trackID)))

	t, err := tracks.GetOpenTracks(r.tracks, trackID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelcodes.Error, err.Error())
		span.AddEvent("getting alternative tracks in district", trace.WithAttributes(
			attribute.String("district", district),
			attribute.String("track", track),
			attribute.String("error", err.Error())))

		return nil, err
	}

	if len(t) == 0 {
		span.AddEvent("fetching alternatives by track in district yielded zero results", trace.WithAttributes(
			attribute.String("track", track),
			attribute.String("district", district),
			attribute.Int("track_count", 0)))

		districts, err := r.ListDistricts(ctx)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelcodes.Error, err.Error())
			span.AddEvent("failed to fetch district list for verification", trace.WithAttributes(
				attribute.String("district", district),
				attribute.String("error", err.Error())))

			return nil, fmt.Errorf("listing tracks by district %q: %w", district, err)
		}

		if !slices.Contains(districts, district) {
			span.RecordError(ErrDistrictNotFound)
			span.SetStatus(otelcodes.Error, ErrDistrictNotFound.Error())
			span.AddEvent("district does not exist", trace.WithAttributes(
				attribute.String("target_district", district),
				attribute.StringSlice("districts", districts),
				attribute.String("error", ErrDistrictNotFound.Error())))

			return nil, ErrDistrictNotFound
		}

		return nil, ErrNoTracks
	}

	span.AddEvent("listed alternatives for track in district successfully", trace.WithAttributes(
		attribute.String("district", district),
		attribute.String("track", track),
		attribute.Int("track_count", len(t))))
	span.SetStatus(otelcodes.Ok, "")

	return tracks.GetNamesFromIDs(r.tracks, t)
}

func (r *Repository) GetCollisionsByDistrictAndTrack(ctx context.Context, district, track string) ([]string, error) {
	trackID, err := tracks.GetIDFromName(r.tracks, track)
	if err != nil {
		return nil, err
	}

	t, err := tracks.GetCollisions(r.tracks, trackID)
	if err != nil {
		return nil, err
	}

	if len(t) == 0 {
		districts, err := r.ListDistricts(ctx)
		if err != nil {
			return nil, fmt.Errorf("listing tracks by district %q: %w", district, err)
		}

		if !slices.Contains(districts, district) {
			return nil, ErrDistrictNotFound
		}

		return nil, ErrNoTracks
	}

	return tracks.GetNamesFromIDs(r.tracks, t)
}
