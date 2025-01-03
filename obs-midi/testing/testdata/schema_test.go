package testdata

import (
	"encoding/json"
	"github.com/zalgonoise/x/obs-midi"
	"os"
	"testing"
)

func TestCreateConfig(t *testing.T) {
	c := &obs_midi.Config{
		Base:               Base,
		SkipSecondary:      NoSecondarySources,
		SkipTertiary:       NoTetriarySources,
		SceneMap:           SceneSet,
		HigherScene:        DefaultHigherScene,
		LowerScene:         DefaultLowerScene,
		ModSource:          DefaultSource,
		PrimarySourceMap:   PrimarySourceSet,
		SecondarySourceMap: SecondarySourceSet,
		TertiarySourceMap:  TetriarySourceSet,
		ToggleMap:          TogglesSet,
		TransitionMap:      TransitionsSet,
		FaderMap:           SliderSet,
		ControlMap:         ControlSet,
		ColorSchema:        obs_midi.GreenColorSchema(),
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.Create(`./local.json`)
	if err != nil {
		t.Fatal(err)
	}

	defer f.Close()

	if _, err := f.Write(data); err != nil {
		t.Fatal(err)
	}

	t.Log("OK")
}
