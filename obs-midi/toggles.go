package obs_midi

import (
	"fmt"
	"strings"
)

func NewTogglesBindings(mapping map[string]SourceNote, sceneNames []string, colors OnOffColor) []Binding {
	bindings := make([]Binding, 0, len(mapping))

	for name, sourceNote := range mapping {
		if strings.HasPrefix(name, blankPrefix) {
			bindingName := fmt.Sprintf("Toggle Blank #%d", sourceNote.NoteOn)
			bindings = append(bindings, Binding{
				Enabled: true,
				Messages: []Message{{
					Channel: channel1,
					Device:  midiDevice,
					Name:    bindingName,
					Note:    ValueInt{Higher: 127, Lower: 0, Number: sourceNote.NoteOn, State: 0},
					Type:    typeNoteOn,
					Value:   ValueInt{Higher: 127, Lower: 0, Number: 127, State: 0},
				}},
				Name:      bindingName,
				ResetMode: 0,
				Type:      0,
			})

			continue
		}

		bindingName := fmt.Sprintf("Toggle %s", name)

		state := toggleStateHide
		if sourceNote.State {
			state = toggleStateShow
		}

		binding := Binding{
			Actions: make([]Action, 0, 30),
			Enabled: true,
			Messages: []Message{{
				Channel: channel1,
				Device:  midiDevice,
				Name:    bindingName,
				Note:    ValueInt{Higher: 127, Lower: 0, Number: sourceNote.NoteOn, State: 0},
				Type:    typeNoteOn,
				Value:   ValueInt{Higher: 127, Lower: 0, Number: 0, State: 1},
			}},
			Name:      bindingName,
			ResetMode: 0,
			Type:      0,
		}

		for i := range sceneNames {
			actionName := fmt.Sprintf("%s %s", bindingName, sceneNames[i])

			binding.Actions = append(binding.Actions, Action{
				Category: 7, Name: actionName, Scene: sceneNames[i], Sub: 1, Type: 0,
				Action: ValueString{Higher: toggleStateHide, Lower: toggleStateShow, State: 0, String: state},
				Json:   JSON{},
				Num1:   ValueFloat{Higher: 1920, Lower: 0, Number: 0, State: 0},
				Num2:   ValueFloat{Higher: 1080, Lower: 0, Number: 0, State: 0},
				Num3:   ValueFloatLower{Higher: 100, Lower: 0.01, Number: 1, State: 0},
				Num4:   ValueFloat{Higher: 7.63651521544067e+218, Lower: 0, Number: 0, State: 0},
				Source: sourceNote.Source,
			})
		}

		ledOnName := fmt.Sprintf("LED On %s", bindingName)

		binding.Actions = append(binding.Actions, Action{Category: 15, Name: ledOnName, Messages: []Message{{
			Channel: channel5,
			Device:  midiDevice,
			Name:    ledOnName,
			Note:    ValueInt{Higher: 127, Lower: 0, Number: sourceNote.NoteOn, State: 0},
			Type:    typeNoteOn,
			Value:   ValueInt{Higher: 127, Lower: 0, Number: int(colors.On), State: 0},
		}}})

		ledOffName := fmt.Sprintf("LED Off %s", bindingName)

		binding.Actions = append(binding.Actions, Action{Category: 15, Name: ledOffName, Messages: []Message{{
			Channel: channel5,
			Device:  midiDevice,
			Name:    ledOffName,
			Note:    ValueInt{Higher: 127, Lower: 0, Number: sourceNote.NoteOff, State: 0},
			Type:    typeNoteOn,
			Value:   ValueInt{Higher: 127, Lower: 0, Number: int(colors.Off), State: 0},
		}}})

		bindings = append(bindings, binding)
	}

	return bindings
}
