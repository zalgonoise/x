package obs_midi

import "time"

const (
	defaultTimeFormat = "2006-01-02 15-04-05-000"
	configName        = "stream"

	defaultSource = "Chrome Opacity"

	midiDevice = "APC mini mk2"
	midiIn     = "MIDIIN2 (APC mini mk2)"
	midiOut    = "MIDIOUT2 (APC mini mk2)"
)

func NewConfigMap(cfg *Config) *ConfigMap {
	return &ConfigMap{
		Collections: []Collection{{
			Name: configName,
			Bindings: NewBindings(cfg.Base, cfg.SkipSecondary,
				cfg.ControlMap, cfg.ToggleMap, cfg.SceneMap, cfg.PrimarySourceMap, cfg.SecondarySourceMap,
				cfg.TransitionMap, cfg.FaderMap, cfg.HigherScene, cfg.LowerScene, cfg.ModSource,
				cfg.ColorSchema,
			),
		}},
		Devices: []Devices{
			{Active: 3, Name: midiDevice},
			{Active: 2, Name: midiOut},
			{Active: 1, Name: midiIn},
		},
		Savedate: time.Now().Format(defaultTimeFormat),
		Version:  "v3.0.3",
	}
}

func NewBindings(
	base string, skipSecondary []string,
	controlSet, togglesSet map[string]SourceNote,
	sceneSet, primarySourceSet, secondarySourceSet, transitionsSet, sliderSet map[string]int,
	higher, lower, source string,
	colorSchema ColorSchema,
) []Binding {
	bindings := make([]Binding, 0, 256)

	scenes := getScenes(sceneSet)
	bindings = append(bindings, NewControlsBindings(controlSet, togglesSet, sceneSet,
		primarySourceSet, secondarySourceSet, transitionsSet, higher, lower, colorSchema)...)
	bindings = append(bindings, NewFaderBindings(sliderSet, source, scenes)...)
	bindings = append(bindings, NewScenesBindings(sceneSet, higher, lower, colorSchema.Scenes)...)
	bindings = append(bindings, NewSourcesBindings(base, skipSecondary, scenes,
		primarySourceSet, secondarySourceSet, colorSchema.Sources)...)
	bindings = append(bindings, NewTogglesBindings(togglesSet, scenes, colorSchema.Toggles)...)
	bindings = append(bindings, NewTransitionBindings(transitionsSet, colorSchema.Transitions)...)

	return bindings
}
