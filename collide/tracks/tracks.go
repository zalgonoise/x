package tracks

import (
	"errors"
	"fmt"
	"slices"

	"gopkg.in/yaml.v3"
)

var (
	ErrNotFound = errors.New("track not found")
)

type Track struct {
	ID           string   `yaml:"id"`
	Name         string   `yaml:"name"`
	District     string   `yaml:"district"`
	IsDriftTrack bool     `yaml:"is_drift_track"`
	CollidesWith []string `yaml:"collides_with"`
}

type TrackList struct {
	Tracks []Track `yaml:"tracks"`
}

func (t *TrackList) Read(b []byte) (n int, err error) {
	if err = yaml.Unmarshal(b, t); err != nil {
		return 0, err
	}

	return len(b), nil
}

func GetCollisions(list *TrackList, trackID string) ([]string, error) {
	i := slices.IndexFunc(list.Tracks, func(t Track) bool {
		return t.ID == trackID
	})

	if i < 0 {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, trackID)
	}

	return list.Tracks[i].CollidesWith, nil
}

func GetOpenTracks(list *TrackList, trackID string) ([]string, error) {
	i := slices.IndexFunc(list.Tracks, func(t Track) bool {
		return t.ID == trackID
	})

	if i < 0 {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, trackID)
	}

	return slices.Collect(func(yield func(t string) bool) {
		for idx := range list.Tracks {
			if i != idx && !slices.Contains(list.Tracks[i].CollidesWith, list.Tracks[idx].ID) {
				yield(list.Tracks[idx].ID)
			}
		}
	}), nil
}
