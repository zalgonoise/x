package tracks

import (
	"gopkg.in/yaml.v3"
)

type Track struct {
	ID           uint8
	Name         string  `yaml:"name"`
	District     string  `yaml:"district"`
	IsDriftTrack bool    `yaml:"is_drift_track"`
	CollidesWith []uint8 `yaml:"collides_with"`
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
