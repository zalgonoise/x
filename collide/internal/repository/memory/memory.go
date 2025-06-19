package memory

import (
	"context"
	"errors"
	"fmt"
	"github.com/zalgonoise/x/collide/internal/tracks"
	"slices"
)

var (
	ErrDistrictNotFound = errors.New("district not found")
	ErrNoDistricts      = errors.New("no districts found")
	ErrNoTracks         = errors.New("no tracks found")
)

type Repository struct {
	tracks *tracks.TrackList
}

func New(tracks *tracks.TrackList) *Repository {
	return &Repository{tracks: tracks}
}

func FromBytes(buf []byte) (*Repository, error) {
	list := &tracks.TrackList{}

	if _, err := list.Read(buf); err != nil {
		return nil, err
	}

	return &Repository{tracks: list}, nil
}

func (r Repository) ListDistricts(ctx context.Context) ([]string, error) {
	districts, err := tracks.GetDistricts(r.tracks)
	if err != nil {
		return nil, err
	}

	if len(districts) == 0 {
		return nil, ErrNoDistricts
	}

	return districts, nil
}

func (r Repository) ListAllTracksByDistrict(ctx context.Context, district string) ([]string, error) {
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

func (r Repository) ListDriftTracksByDistrict(ctx context.Context, district string) ([]string, error) {
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

func (r Repository) GetAlternativesByDistrictAndTrack(ctx context.Context, district, track string) ([]string, error) {
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

func (r Repository) GetCollisionsByDistrictAndTrack(ctx context.Context, district, track string) ([]string, error) {
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
