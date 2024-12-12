package obs_midi

import (
	"fmt"
	"slices"
	"strings"
)

func NewSourcesBindings(
	base string, skipSecondary, skipTetriary, sceneNames []string,
	primarySources, secondarySources, tetriarySources map[string]int,
	colors OnOffColor,
) []Binding {
	bindings := make([]Binding, 0, len(primarySources)+len(secondarySources))

	for name, note := range primarySources {
		// skip blank
		if strings.HasPrefix(name, blankPrefix) {
			bindingName := fmt.Sprintf("Toggle Blank #%d", note)
			bindings = append(bindings, Binding{
				Enabled: true,
				Messages: []Message{{
					Channel: channel1,
					Device:  midiDevice,
					Name:    bindingName,
					Note:    ValueInt{Higher: 127, Lower: 0, Number: note, State: 0},
					Type:    typeNoteOn,
					Value:   ValueInt{Higher: 127, Lower: 0, Number: 127, State: 0},
				}},
				Name:      bindingName,
				ResetMode: 0,
				Type:      0,
			})

			continue
		}

		bindingName := fmt.Sprintf("Enable Source %s", name)

		// set note-on binding
		binding := Binding{
			Actions: make([]Action, 0, 30),
			Enabled: true,
			Messages: []Message{{
				Channel: channel1,
				Device:  midiDevice,
				Name:    bindingName,
				Note:    ValueInt{Higher: 127, Lower: 0, Number: note, State: 0},
				Type:    typeNoteOn,
				Value:   ValueInt{Higher: 127, Lower: 0, Number: 0, State: 1},
			}},
			Name:      bindingName,
			ResetMode: 0,
			Type:      0,
		}

		// LED On
		ledOnName := fmt.Sprintf("LED On %s", bindingName)

		binding.Actions = append(binding.Actions, Action{Category: 15, Name: ledOnName, Messages: []Message{{
			Channel: channel5,
			Device:  midiDevice,
			Name:    ledOnName,
			Note:    ValueInt{Higher: 127, Lower: 0, Number: note, State: 0},
			Type:    typeNoteOn,
			Value:   ValueInt{Higher: 127, Lower: 0, Number: colors.On, State: 0},
		}}})

		// apply to all scenes
		for i := range sceneNames {
			actionName := fmt.Sprintf("%s %s", bindingName, sceneNames[i])

			// show this source
			binding.Actions = append(binding.Actions, Action{
				Category: 7, Name: actionName, Scene: sceneNames[i], Sub: 1, Type: 0,
				Action: ValueString{Higher: toggleStateHide, Lower: toggleStateShow, State: 0, String: toggleStateShow},
				Json:   JSON{},
				Num1:   ValueFloat{Higher: 1920, Lower: 0, Number: 0, State: 0},
				Num2:   ValueFloat{Higher: 1080, Lower: 0, Number: 0, State: 0},
				Num3:   ValueFloatLower{Higher: 100, Lower: 0.01, Number: 1, State: 0},
				Num4:   ValueFloat{Higher: 7.63651521544067e+218, Lower: 0, Number: 0, State: 0},
				Source: name,
			})

			ledOffName := fmt.Sprintf("LED Off %s", bindingName)

			for inner, innerNote := range primarySources {
				if strings.HasPrefix(inner, blankPrefix) || inner == name {
					continue
				}

				// hide source
				if inner != base {
					binding.Actions = append(binding.Actions, Action{
						Category: 7, Name: actionName, Scene: sceneNames[i], Sub: 1, Type: 0,
						Action: ValueString{Higher: toggleStateHide, Lower: toggleStateShow, State: 0, String: toggleStateHide},
						Json:   JSON{},
						Num1:   ValueFloat{Higher: 1920, Lower: 0, Number: 0, State: 0},
						Num2:   ValueFloat{Higher: 1080, Lower: 0, Number: 0, State: 0},
						Num3:   ValueFloatLower{Higher: 100, Lower: 0.01, Number: 1, State: 0},
						Num4:   ValueFloat{Higher: 7.63651521544067e+218, Lower: 0, Number: 0, State: 0},
						Source: inner,
					})
				}

				// LED Off
				binding.Actions = append(binding.Actions, Action{Category: 15, Name: ledOffName, Messages: []Message{{
					Channel: channel5,
					Device:  midiDevice,
					Name:    ledOffName,
					Note:    ValueInt{Higher: 127, Lower: 0, Number: innerNote, State: 0},
					Type:    typeNoteOn,
					Value:   ValueInt{Higher: 127, Lower: 0, Number: colors.Off, State: 0},
				}}})
			}
		}

		bindings = append(bindings, binding)
	}

	for name, note := range secondarySources {
		// skip blank
		if strings.HasPrefix(name, blankPrefix) {
			bindingName := fmt.Sprintf("Toggle Blank #%d", note)
			bindings = append(bindings, Binding{
				Enabled: true,
				Messages: []Message{{
					Channel: channel1,
					Device:  midiDevice,
					Name:    bindingName,
					Note:    ValueInt{Higher: 127, Lower: 0, Number: note, State: 0},
					Type:    typeNoteOn,
					Value:   ValueInt{Higher: 127, Lower: 0, Number: 127, State: 0},
				}},
				Name:      bindingName,
				ResetMode: 0,
				Type:      0,
			})

			continue
		}

		bindingName := fmt.Sprintf("Enable Secondary Source %s", name)

		// set note-on binding
		binding := Binding{
			Actions: make([]Action, 0, 30),
			Enabled: true,
			Messages: []Message{{
				Channel: channel1,
				Device:  midiDevice,
				Name:    bindingName,
				Note:    ValueInt{Higher: 127, Lower: 0, Number: note, State: 0},
				Type:    typeNoteOn,
				Value:   ValueInt{Higher: 127, Lower: 0, Number: 0, State: 1},
			}},
			Name:      bindingName,
			ResetMode: 0,
			Type:      0,
		}

		ledOnName := fmt.Sprintf("LED On %s", bindingName)

		// LED On
		binding.Actions = append(binding.Actions, Action{Category: 15, Name: ledOnName, Messages: []Message{{
			Channel: channel5,
			Device:  midiDevice,
			Name:    ledOnName,
			Note:    ValueInt{Higher: 127, Lower: 0, Number: note, State: 0},
			Type:    typeNoteOn,
			Value:   ValueInt{Higher: 127, Lower: 0, Number: colors.On, State: 0},
		}}})

		// apply to all scenes
		for i := range sceneNames {
			// skip secondary
			if slices.Contains(skipSecondary, sceneNames[i]) {
				continue
			}

			actionName := fmt.Sprintf("%s %s", bindingName, sceneNames[i])

			// show this source
			binding.Actions = append(binding.Actions, Action{
				Category: 7, Name: actionName, Scene: sceneNames[i], Sub: 1, Type: 0,
				Action: ValueString{Higher: toggleStateHide, Lower: toggleStateShow, State: 0, String: toggleStateShow},
				Json:   JSON{},
				Num1:   ValueFloat{Higher: 1920, Lower: 0, Number: 0, State: 0},
				Num2:   ValueFloat{Higher: 1080, Lower: 0, Number: 0, State: 0},
				Num3:   ValueFloatLower{Higher: 100, Lower: 0.01, Number: 1, State: 0},
				Num4:   ValueFloat{Higher: 7.63651521544067e+218, Lower: 0, Number: 0, State: 0},
				Source: name,
			})

			ledOffName := fmt.Sprintf("LED Off %s", bindingName)

			for inner, innerNote := range secondarySources {
				if strings.HasPrefix(inner, blankPrefix) || inner == name {
					continue
				}

				// hide source
				if inner != base {
					binding.Actions = append(binding.Actions, Action{
						Category: 7, Name: actionName, Scene: sceneNames[i], Sub: 1, Type: 0,
						Action: ValueString{Higher: toggleStateHide, Lower: toggleStateShow, State: 0, String: toggleStateHide},
						Json:   JSON{},
						Num1:   ValueFloat{Higher: 1920, Lower: 0, Number: 0, State: 0},
						Num2:   ValueFloat{Higher: 1080, Lower: 0, Number: 0, State: 0},
						Num3:   ValueFloatLower{Higher: 100, Lower: 0.01, Number: 1, State: 0},
						Num4:   ValueFloat{Higher: 7.63651521544067e+218, Lower: 0, Number: 0, State: 0},
						Source: inner,
					})
				}

				// LED Off
				binding.Actions = append(binding.Actions, Action{Category: 15, Name: ledOffName, Messages: []Message{{
					Channel: channel5,
					Device:  midiDevice,
					Name:    ledOffName,
					Note:    ValueInt{Higher: 127, Lower: 0, Number: innerNote, State: 0},
					Type:    typeNoteOn,
					Value:   ValueInt{Higher: 127, Lower: 0, Number: colors.Off, State: 0},
				}}})
			}
		}

		bindings = append(bindings, binding)
	}

	for name, note := range tetriarySources {
		// skip blank
		if strings.HasPrefix(name, blankPrefix) {
			bindingName := fmt.Sprintf("Toggle Blank #%d", note)
			bindings = append(bindings, Binding{
				Enabled: true,
				Messages: []Message{{
					Channel: channel1,
					Device:  midiDevice,
					Name:    bindingName,
					Note:    ValueInt{Higher: 127, Lower: 0, Number: note, State: 0},
					Type:    typeNoteOn,
					Value:   ValueInt{Higher: 127, Lower: 0, Number: 127, State: 0},
				}},
				Name:      bindingName,
				ResetMode: 0,
				Type:      0,
			})

			continue
		}

		bindingName := fmt.Sprintf("Enable Secondary Source %s", name)

		// set note-on binding
		binding := Binding{
			Actions: make([]Action, 0, 30),
			Enabled: true,
			Messages: []Message{{
				Channel: channel1,
				Device:  midiDevice,
				Name:    bindingName,
				Note:    ValueInt{Higher: 127, Lower: 0, Number: note, State: 0},
				Type:    typeNoteOn,
				Value:   ValueInt{Higher: 127, Lower: 0, Number: 0, State: 1},
			}},
			Name:      bindingName,
			ResetMode: 0,
			Type:      0,
		}

		ledOnName := fmt.Sprintf("LED On %s", bindingName)

		// LED On
		binding.Actions = append(binding.Actions, Action{Category: 15, Name: ledOnName, Messages: []Message{{
			Channel: channel5,
			Device:  midiDevice,
			Name:    ledOnName,
			Note:    ValueInt{Higher: 127, Lower: 0, Number: note, State: 0},
			Type:    typeNoteOn,
			Value:   ValueInt{Higher: 127, Lower: 0, Number: colors.On, State: 0},
		}}})

		// apply to all scenes
		for i := range sceneNames {
			// skip tetriary
			if slices.Contains(skipTetriary, sceneNames[i]) {
				continue
			}

			actionName := fmt.Sprintf("%s %s", bindingName, sceneNames[i])

			// show this source
			binding.Actions = append(binding.Actions, Action{
				Category: 7, Name: actionName, Scene: sceneNames[i], Sub: 1, Type: 0,
				Action: ValueString{Higher: toggleStateHide, Lower: toggleStateShow, State: 0, String: toggleStateShow},
				Json:   JSON{},
				Num1:   ValueFloat{Higher: 1920, Lower: 0, Number: 0, State: 0},
				Num2:   ValueFloat{Higher: 1080, Lower: 0, Number: 0, State: 0},
				Num3:   ValueFloatLower{Higher: 100, Lower: 0.01, Number: 1, State: 0},
				Num4:   ValueFloat{Higher: 7.63651521544067e+218, Lower: 0, Number: 0, State: 0},
				Source: name,
			})

			ledOffName := fmt.Sprintf("LED Off %s", bindingName)

			for inner, innerNote := range tetriarySources {
				if strings.HasPrefix(inner, blankPrefix) || inner == name {
					continue
				}

				// hide source
				if inner != base {
					binding.Actions = append(binding.Actions, Action{
						Category: 7, Name: actionName, Scene: sceneNames[i], Sub: 1, Type: 0,
						Action: ValueString{Higher: toggleStateHide, Lower: toggleStateShow, State: 0, String: toggleStateHide},
						Json:   JSON{},
						Num1:   ValueFloat{Higher: 1920, Lower: 0, Number: 0, State: 0},
						Num2:   ValueFloat{Higher: 1080, Lower: 0, Number: 0, State: 0},
						Num3:   ValueFloatLower{Higher: 100, Lower: 0.01, Number: 1, State: 0},
						Num4:   ValueFloat{Higher: 7.63651521544067e+218, Lower: 0, Number: 0, State: 0},
						Source: inner,
					})
				}

				// LED Off
				binding.Actions = append(binding.Actions, Action{Category: 15, Name: ledOffName, Messages: []Message{{
					Channel: channel5,
					Device:  midiDevice,
					Name:    ledOffName,
					Note:    ValueInt{Higher: 127, Lower: 0, Number: innerNote, State: 0},
					Type:    typeNoteOn,
					Value:   ValueInt{Higher: 127, Lower: 0, Number: colors.Off, State: 0},
				}}})
			}
		}

		bindings = append(bindings, binding)
	}

	return bindings
}
