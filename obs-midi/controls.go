package obs_midi

import (
	"fmt"
	"slices"
	"strings"
)

func NewControlsBindings(
	mapping, toggles map[string]SourceNote,
	scenes, transitions map[string]int,
	higher, lower string,
	colorSchema ColorSchema,
) []Binding {
	bindings := make([]Binding, 0, len(mapping))

	for name, notes := range mapping {
		if !slices.Contains(SupportedControlModes, name) {
			continue
		}

		bindingName := fmt.Sprintf("Core - %s", name)

		switch name {
		case ControlModeSaveReplayBuffer:
			bindings = append(bindings, Binding{
				Enabled: true,
				Actions: []Action{
					{Category: 4, Name: bindingName, Sub: 3, Type: 0},
					{Category: 15, Name: "LED Off", Messages: []Message{{
						Channel: channel1,
						Device:  midiDevice,
						Name:    "LED Off",
						Note:    ValueInt{Higher: 127, Lower: 0, Number: notes.NoteOn, State: 0},
						Type:    typeNoteOn,
						Value:   ValueInt{Higher: 127, Lower: 0, Number: 0, State: 0},
					}}},
				},
				Messages: []Message{{
					Channel: channel1,
					Device:  midiDevice,
					Name:    bindingName,
					Note:    ValueInt{Higher: 127, Lower: 0, Number: notes.NoteOn, State: 0},
					Type:    typeNoteOn,
					Value:   ValueInt{Higher: 127, Lower: 0, Number: 127, State: 0},
				}},
				Name: bindingName, ResetMode: 0, Type: 0,
			})
		case ControlModeStudioModeOn, ControlModeStudioModeOff:
			var sub int
			if name == ControlModeStudioModeOff {
				sub = 1
			}

			bindings = append(bindings, Binding{
				Actions: []Action{
					{
						Category: 5, Name: bindingName, Sub: sub, Type: 0,
						Scene: ValueString{Higher: higher, Lower: lower, State: 4, String: lower},
					},
					{
						Category: 15, Name: "LED Off", Messages: []Message{{
							Channel: channel1,
							Device:  midiDevice,
							Name:    "LED Off",
							Note:    ValueInt{Higher: 127, Lower: 0, Number: notes.NoteOff, State: 0},
							Type:    typeNoteOn,
							Value:   ValueInt{Higher: 127, Lower: 0, Number: 0, State: 0},
						}},
					},
					{
						Category: 15, Name: "LED On", Messages: []Message{{
							Channel: channel1,
							Device:  midiDevice,
							Name:    "LED On",
							Note:    ValueInt{Higher: 127, Lower: 0, Number: notes.NoteOff, State: 0},
							Type:    typeNoteOn,
							Value:   ValueInt{Higher: 127, Lower: 0, Number: 1, State: 0},
						}},
					},
				},
				Enabled: true,
				Messages: []Message{{
					Channel: channel1, Device: midiDevice, Name: bindingName,
					Note:  ValueInt{Higher: 127, Lower: 0, Number: notes.NoteOn, State: 0},
					Type:  typeNoteOn,
					Value: ValueInt{Higher: 127, Lower: 0, Number: 127, State: 0},
				}},
				Name: bindingName, ResetMode: 0, Type: 0,
			})

		case ControlModeLEDOff:
			binding := Binding{
				Actions: make([]Action, 0, 128),
				Enabled: true,
				Messages: []Message{{
					Channel: channel1,
					Device:  midiDevice,
					Name:    bindingName,
					Note:    ValueInt{Higher: 127, Lower: 0, Number: notes.NoteOn, State: 0},
					Type:    typeNoteOn,
					Value:   ValueInt{Higher: 127, Lower: 0, Number: 127, State: 0},
				}},
				Name:      bindingName,
				ResetMode: 0,
				Type:      0,
			}

			for i := minCh5Note; i <= maxCh5Note; i++ {
				actionName := fmt.Sprintf("%s #%d", bindingName, i)
				binding.Actions = append(binding.Actions, Action{Category: 15, Name: actionName, Messages: []Message{{
					Channel: channel5,
					Device:  midiDevice,
					Name:    actionName,
					Note:    ValueInt{Higher: 127, Lower: 0, Number: i, State: 0},
					Type:    typeNoteOn,
					Value:   ValueInt{Higher: 127, Lower: 0, Number: 0, State: 0},
				}}})
			}

			for i := minCh1Note; i <= maxCh1Note; i++ {
				actionName := fmt.Sprintf("%s #%d", bindingName, i)
				binding.Actions = append(binding.Actions, Action{Category: 15, Name: actionName, Messages: []Message{{
					Channel: channel1,
					Device:  midiDevice,
					Name:    actionName,
					Note:    ValueInt{Higher: 127, Lower: 0, Number: i, State: 0},
					Type:    typeNoteOn,
					Value:   ValueInt{Higher: 127, Lower: 0, Number: 0, State: 0},
				}}})
			}

			binding.Actions = append(binding.Actions, Action{Category: 15, Name: "Enable LED", Messages: []Message{{
				Channel: channel5,
				Device:  midiDevice,
				Name:    "Enable LED",
				Note:    ValueInt{Higher: 127, Lower: 0, Number: notes.NoteOn, State: 0},
				Type:    typeNoteOn,
				Value:   ValueInt{Higher: 127, Lower: 0, Number: 0, State: 0},
			}}})

			bindings = append(bindings, binding)

		case ControlModeLEDOn:
			binding := Binding{
				Actions: make([]Action, 0, len(scenes)+len(transitions)+len(toggles)+2),
				Enabled: true,
				Messages: []Message{{
					Channel: channel1,
					Device:  midiDevice,
					Name:    bindingName,
					Note:    ValueInt{Higher: 127, Lower: 0, Number: notes.NoteOn, State: 0},
					Type:    typeNoteOn,
					Value:   ValueInt{Higher: 127, Lower: 0, Number: 127, State: 0},
				}},
				Name:      bindingName,
				ResetMode: 0,
				Type:      0,
			}

			for sceneName, note := range scenes {
				color := colorSchema.Scenes.Off
				if strings.HasPrefix(sceneName, blankPrefix) {
					color = colorSchema.Scenes.Blank
				}

				actionName := fmt.Sprintf("%s - %s #%d", name, sceneName, note)

				binding.Actions = append(binding.Actions, Action{Category: 15, Name: actionName, Messages: []Message{{
					Channel: channel5,
					Device:  midiDevice,
					Name:    actionName,
					Note:    ValueInt{Higher: 127, Lower: 0, Number: note, State: 0},
					Type:    typeNoteOn,
					Value:   ValueInt{Higher: 127, Lower: 0, Number: color, State: 0},
				}}})
			}

			for toggleName, toggleNotes := range toggles {
				color := colorSchema.Toggles.Off

				switch {
				case strings.HasPrefix(toggleName, blankPrefix):
					color = colorSchema.Toggles.Blank
				case !toggleNotes.State:
					color = colorSchema.Toggles.On
				}

				actionName := fmt.Sprintf("%s - %s #%d", name, toggleName, toggleNotes.NoteOn)

				binding.Actions = append(binding.Actions, Action{Category: 15, Name: actionName, Messages: []Message{{
					Channel: channel5,
					Device:  midiDevice,
					Name:    actionName,
					Note:    ValueInt{Higher: 127, Lower: 0, Number: toggleNotes.NoteOn, State: 0},
					Type:    typeNoteOn,
					Value:   ValueInt{Higher: 127, Lower: 0, Number: color, State: 0},
				}}})
			}

			for transitionName, note := range transitions {
				actionName := fmt.Sprintf("%s - %s #%d", name, transitionName, note)

				color := colorSchema.Transitions.Off
				switch {
				case transitionName == TransitionToProgram:
					color = colorSchema.Transitions.Transition
				case strings.HasPrefix(transitionName, blankPrefix):
					color = colorSchema.Transitions.Blank
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

			binding.Actions = append(binding.Actions, Action{Category: 15, Name: "Enable LED", Messages: []Message{{
				Channel: channel5,
				Device:  midiDevice,
				Name:    "Enable LED",
				Note:    ValueInt{Higher: 127, Lower: 0, Number: notes.NoteOn, State: 0},
				Type:    typeNoteOn,
				Value:   ValueInt{Higher: 127, Lower: 0, Number: 0, State: 0},
			}}})

			binding.Actions = append(binding.Actions, Action{Category: 15, Name: "Disable LED", Messages: []Message{{
				Channel: channel5,
				Device:  midiDevice,
				Name:    "Disable LED",
				Note:    ValueInt{Higher: 127, Lower: 0, Number: notes.NoteOff, State: 0},
				Type:    typeNoteOn,
				Value:   ValueInt{Higher: 127, Lower: 0, Number: 0, State: 0},
			}}})

			bindings = append(bindings, binding)
		}
	}

	return bindings
}
