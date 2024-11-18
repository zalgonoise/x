package obs_midi

import (
	"fmt"
	"strings"
)

func NewScenesBindings(mapping map[string]int, higher, lower string, colors OnOffColor) []Binding {
	scenes := make([]Binding, 0, 48)

	for sceneName, sceneNote := range mapping {
		bindName := fmt.Sprintf("Prev. %s #%d", sceneName, sceneNote)

		binding := Binding{
			Actions: make([]Action, 0, 30),
			Enabled: true,
			Messages: []Message{{
				Channel: channel1,
				Device:  midiDevice,
				Name:    bindName,
				Note:    ValueInt{Higher: 127, Lower: 0, Number: sceneNote, State: 0},
				Type:    typeNoteOn,
				Value:   ValueInt{Higher: 127, Lower: 0, Number: 127, State: 0},
			}},
			Name:      bindName,
			ResetMode: 0,
			Type:      0,
		}

		binding.Actions = append(binding.Actions, Action{Category: 5, Name: fmt.Sprintf("Prev. %s", sceneName),
			Scene: ValueString{
				Higher: higher,
				Lower:  lower,
				State:  0,
				String: sceneName,
			}, Sub: 3, Type: 0,
		})

		for name, note := range mapping {
			actionName := fmt.Sprintf("LED #%d %s", note, name)

			var color int

			switch {
			case strings.HasPrefix(name, blankPrefix):
				color = colors.Blank
			case name == sceneName:
				color = colors.On
			default:
				color = colors.Off
			}

			binding.Actions = append(binding.Actions, Action{Category: 15, Name: actionName, Messages: []Message{{
				Channel: channel5,
				Device:  midiDevice,
				Name:    actionName,
				Note:    ValueInt{Higher: 127, Lower: 0, Number: note, State: 0},
				Type:    typeNoteOn,
				Value:   ValueInt{Higher: 127, Lower: 0, Number: color, State: 0},
			}}})
		}

		scenes = append(scenes, binding)
	}

	return scenes
}
