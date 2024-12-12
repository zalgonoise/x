package obs_midi

import (
	"errors"
	"slices"
	"strings"
)

var ErrSourceConfigEmpty = errors.New("source config is empty")

type Config struct {
	Base          string   `json:"base"`
	SkipSecondary []string `json:"skip_secondary"`
	SkipTetriary  []string `json:"skip_tetriary"`

	SceneMap           map[string]int `json:"scene_map"`
	HigherScene        string         `json:"higher_scene"`
	LowerScene         string         `json:"lower_scene"`
	ModSource          string         `json:"mod_source"`
	PrimarySourceMap   map[string]int `json:"primary_source_map"`
	SecondarySourceMap map[string]int `json:"secondary_source_map"`
	TetriarySourceMap  map[string]int `json:"tetriary_source_map"`

	ToggleMap     map[string]SourceNote `json:"toggle_map"`
	TransitionMap map[string]int        `json:"transition_map"`
	FaderMap      map[string]int        `json:"fader_map"`
	ControlMap    map[string]SourceNote `json:"control_map"`

	ColorSchema ColorSchema `json:"color_schema"`
}

func (c *Config) Validate() error {
	if c.ModSource == "" {
		return ErrSourceConfigEmpty
	}

	if len(c.SceneMap) > 0 && (c.HigherScene == "" || c.LowerScene == "") {
		scenes := getScenes(c.SceneMap)

		if c.HigherScene == "" {
			c.HigherScene = scenes[len(scenes)-1]
		}

		if c.LowerScene == "" {
			c.HigherScene = scenes[0]
		}
	}

	return nil
}

func getScenes(sceneMap map[string]int) []string {
	scenes := make([]string, 0, len(sceneMap))

	for scene := range sceneMap {
		if scene == "" || strings.HasPrefix(scene, blankPrefix) {
			continue
		}

		scenes = append(scenes, scene)
	}

	slices.Sort(scenes)

	return scenes
}
