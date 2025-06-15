package memory

import (
	"context"
	"github.com/zalgonoise/x/collide/tracks"
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
	return tracks.GetDistricts(r.tracks)
}

func (r Repository) ListAllTracksByDistrict(ctx context.Context, district string) ([]string, error) {
	return tracks.GetTracksByDistrict(r.tracks, district, false)
}

func (r Repository) ListDriftTracksByDistrict(ctx context.Context, district string) ([]string, error) {
	return tracks.GetTracksByDistrict(r.tracks, district, true)
}

func (r Repository) GetAlternativesByDistrictAndTrack(ctx context.Context, district, trackID string) ([]string, error) {
	return tracks.GetOpenTracks(r.tracks, trackID)
}

func (r Repository) GetCollisionsByDistrictAndTrack(ctx context.Context, district, trackID string) ([]string, error) {
	return tracks.GetCollisions(r.tracks, trackID)
}
