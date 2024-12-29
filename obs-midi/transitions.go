package obs_midi

import (
	"fmt"
	"strings"
)

func NewTransitionBindings(mapping map[string]int, colors TransitionsColor) []Binding {
	bindings := make([]Binding, 0, len(mapping))

	setTransitions := make([]int, 0, len(mapping))
	for name, note := range mapping {
		if name == TransitionToProgram {
			continue
		}

		if strings.HasPrefix(name, blankPrefix) {
			continue
		}

		setTransitions = append(setTransitions, note)
	}

	for name, note := range mapping {
		switch {
		case name == TransitionToProgram:
			bindings = append(bindings, Binding{
				Actions: []Action{
					{Category: 5, Name: TransitionToProgram, Scene: ValueString{}, Sub: 4, Type: 0},
					{Category: 15, Name: "Enable LED", Sub: 0, Type: 0, Messages: []Message{{
						Channel: channel5,
						Device:  midiDevice,
						Name:    "Enable LED",
						Note:    ValueInt{Higher: 127, Lower: 0, Number: note, State: 0},
						Type:    typeNoteOn,
						Value:   ValueInt{Higher: 127, Lower: 0, Number: int(colors.Transition), State: 0},
					}}},
				},
				Enabled: true,
				Messages: []Message{{
					Channel: channel1,
					Device:  midiDevice,
					Name:    TransitionToProgram,
					Note:    ValueInt{Higher: 127, Lower: 0, Number: note, State: 0},
					Type:    typeNoteOn,
					Value:   ValueInt{Higher: 127, Lower: 0, Number: 127, State: 0},
				}},
				Name:      TransitionToProgram,
				ResetMode: 0,
				Type:      0,
			})
		case strings.HasPrefix(name, blankPrefix):
			bindingName := fmt.Sprintf("Transition - Blank #%d", note)

			bindings = append(bindings, Binding{
				Actions: []Action{{
					Category: 15, Name: bindingName, Messages: []Message{{
						Channel: channel5,
						Device:  midiDevice,
						Name:    bindingName,
						Note:    ValueInt{Higher: 127, Lower: 0, Number: note, State: 0},
						Type:    typeNoteOn,
						Value:   ValueInt{Higher: 127, Lower: 0, Number: int(colors.Blank), State: 0}}}}},
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

		default:
			bindingName := fmt.Sprintf("Set to %s", name)

			binding := Binding{
				Actions: make([]Action, 0, 5),
				Enabled: true,
				Messages: []Message{{
					Channel: channel1,
					Device:  midiDevice,
					Name:    bindingName,
					Note:    ValueInt{Higher: 127, Lower: 0, Number: note, State: 0},
					Type:    typeNoteOn,
					Value:   ValueInt{Higher: 127, Lower: 0, Number: 0, State: 1},
				}},
				Name: bindingName, ResetMode: 0, Type: 0,
			}

			binding.Actions = append(binding.Actions, Action{
				Category: 10, Name: bindingName, Scene: "", Sub: 0, Type: 0, Source: "",
				Num: ValueInt{
					Higher: 20000,
					Lower:  25,
					Number: 400,
					State:  0,
				},
				Transition: name,
			})

			binding.Actions = append(binding.Actions, Action{Category: 15, Name: "LED On", Messages: []Message{{
				Channel: channel5,
				Device:  midiDevice,
				Name:    "LED On",
				Note:    ValueInt{Higher: 127, Lower: 0, Number: note, State: 0},
				Type:    typeNoteOn,
				Value:   ValueInt{Higher: 127, Lower: 0, Number: int(colors.On), State: 0},
			}}})

			for i := range setTransitions {
				if setTransitions[i] == note {
					continue
				}

				binding.Actions = append(binding.Actions, Action{Category: 15, Name: "LED Off", Messages: []Message{{
					Channel: channel5,
					Device:  midiDevice,
					Name:    "LED Off",
					Note:    ValueInt{Higher: 127, Lower: 0, Number: setTransitions[i], State: 0},
					Type:    typeNoteOn,
					Value:   ValueInt{Higher: 127, Lower: 0, Number: int(colors.Off), State: 0},
				}}})
			}

			bindings = append(bindings, binding)
		}
	}

	return bindings
}
