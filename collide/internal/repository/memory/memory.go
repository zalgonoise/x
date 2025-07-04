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
	t, err := tracks.GetTracksByDistrict(r.tracks, district, false)
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

func (r *Repository) ListDriftTracksByDistrict(ctx context.Context, district string) ([]string, error) {
	t, err := tracks.GetTracksByDistrict(r.tracks, district, true)
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

func (r *Repository) GetAlternativesByDistrictAndTrack(ctx context.Context, district, track string) ([]string, error) {
	trackID, err := tracks.GetIDFromName(r.tracks, track)
	if err != nil {
		return nil, err
	}

	t, err := tracks.GetOpenTracks(r.tracks, trackID)
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
