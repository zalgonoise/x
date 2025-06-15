package tracks

import (
	"errors"
	"fmt"
	"slices"

	"gopkg.in/yaml.v3"
)

var (
	ErrNotFound = errors.New("track not found")
	ErrNilList  = errors.New("list cannot be nil")
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
	if list == nil {
		return nil, ErrNilList
	}

	i := slices.IndexFunc(list.Tracks, func(t Track) bool {
		return t.ID == trackID
	})

	if i < 0 {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, trackID)
	}

	return list.Tracks[i].CollidesWith, nil
}

func GetOpenTracks(list *TrackList, trackID string) ([]string, error) {
	if list == nil {
		return nil, ErrNilList
	}

	i := slices.IndexFunc(list.Tracks, func(t Track) bool {
		return t.ID == trackID
	})

	if i < 0 {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, trackID)
	}

	return slices.Collect(func(yield func(t string) bool) {
		for idx := range list.Tracks {
			if i != idx &&
				list.Tracks[idx].District == list.Tracks[i].District &&
				!slices.Contains(list.Tracks[i].CollidesWith, list.Tracks[idx].ID) {
				yield(list.Tracks[idx].ID)
			}
		}
	}), nil
}

func GetDriftTracks(list *TrackList) ([]string, error) {
	if list == nil {
		return nil, ErrNilList
	}

	return slices.Collect(func(yield func(t string) bool) {
		for idx := range list.Tracks {
			if list.Tracks[idx].IsDriftTrack {
				yield(list.Tracks[idx].ID)
			}
		}
	}), nil
}

func GetTracksByDistrict(list *TrackList, district string, driftOnly bool) ([]string, error) {
	if list == nil {
		return nil, ErrNilList
	}

	return slices.Collect(func(yield func(t string) bool) {
		for idx := range list.Tracks {
			if list.Tracks[idx].District == district {
				if driftOnly && !list.Tracks[idx].IsDriftTrack {
					continue
				}

				yield(list.Tracks[idx].ID)
			}
		}
	}), nil
}

func GetDistricts(list *TrackList) ([]string, error) {
	if list == nil {
		return nil, ErrNilList
	}

	return slices.Collect(func(yield func(t string) bool) {
		cache := map[string]struct{}{}

		for idx := range list.Tracks {
			if _, ok := cache[list.Tracks[idx].District]; !ok {
				cache[list.Tracks[idx].District] = struct{}{}

				yield(list.Tracks[idx].District)
			}
		}
	}), nil
}

func GetNamesFromIDs(list *TrackList, ids []string) ([]string, error) {
	if list == nil {
		return nil, ErrNilList
	}

	return slices.Collect(func(yield func(name string) bool) {
		for i := range ids {
			idx := slices.IndexFunc(list.Tracks, func(track Track) bool {
				return track.ID == ids[i]
			})

			if idx >= 0 {
				yield(list.Tracks[idx].Name)
			}
		}
	}), nil
}
