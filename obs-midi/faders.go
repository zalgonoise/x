package obs_midi

import (
	"fmt"
	"slices"
)

func NewFaderBindings(mapping map[string]int, source string, sceneNames []string) []Binding {
	bindings := make([]Binding, 0, len(mapping))

	for name, note := range mapping {
		if !slices.Contains(SupportedSliderModes, name) {
			continue
		}

		bindingName := fmt.Sprintf("CC %s", name)

		switch name {
		case SliderModeTransition:
			bindings = append(bindings, Binding{
				Actions: []Action{{Category: 10, Name: bindingName, Scene: "", Sub: 3, Type: 0,
					Json:   JSON{},
					Num:    ValueInt{Higher: 1024, Lower: 0, Number: 0, State: 1},
					Source: "", Transition: defaultTransition, Filter: "",
				}},
				Enabled: true,
				Messages: []Message{{
					Channel: channel1,
					Device:  midiDevice,
					Name:    bindingName,
					Note:    ValueInt{Higher: 127, Lower: 0, Number: note, State: 0},
					Type:    typeCC,
					Value:   ValueInt{Higher: 127, Lower: 0, Number: 0, State: 1},
				}},
				Name:      bindingName,
				ResetMode: 0,
				Type:      0,
			})
		case SliderModeMoveX, SliderModeMoveY:
			num1State := 1
			num2State := 3

			if name == SliderModeMoveY {
				num1State = 3
				num2State = 1
			}

			binding := Binding{
				Actions: make([]Action, 0, 32),
				Enabled: true,
				Messages: []Message{{
					Channel: channel1,
					Device:  midiDevice,
					Name:    bindingName,
					Note:    ValueInt{Higher: 127, Lower: 0, Number: note, State: 0},
					Type:    typeCC,
					Value:   ValueInt{Higher: 127, Lower: 0, Number: 0, State: 2},
				}},
				Name:      bindingName,
				ResetMode: 0,
				Type:      0,
			}

			for i := range sceneNames {
				binding.Actions = append(binding.Actions, Action{
					Category: 7, Name: fmt.Sprintf("CC %s %s", name, sceneNames[i]),
					Scene: sceneNames[i], Sub: 0, Type: 0,
					Action: ValueString{}, Json: JSON{},
					Num1:   ValueFloat{Higher: 1920, Lower: 0, Number: 0, State: num1State},
					Num2:   ValueFloat{Higher: 1080, Lower: 0, Number: 357.16534423828125, State: num2State},
					Num3:   ValueFloatLower{Higher: 100, Lower: 0.01, Number: 2, State: 0},
					Num4:   ValueFloat{},
					Source: source,
				})
			}

			bindings = append(bindings, binding)

		case SliderModeScale:
			binding := Binding{
				Actions: make([]Action, 0, 32),
				Enabled: true,
				Messages: []Message{{
					Channel: channel1,
					Device:  midiDevice,
					Name:    bindingName,
					Note:    ValueInt{Higher: 127, Lower: 0, Number: note, State: 0},
					Type:    typeCC,
					Value:   ValueInt{Higher: 127, Lower: 0, Number: 0, State: 2},
				}},
				Name:      bindingName,
				ResetMode: 0,
				Type:      0,
			}

			for i := range sceneNames {
				binding.Actions = append(binding.Actions, Action{
					Category: 7, Name: fmt.Sprintf("CC %s %s", name, sceneNames[i]),
					Scene: sceneNames[i], Sub: 5, Type: 0,
					Action: ValueString{}, Json: JSON{},
					Num1:   ValueFloat{Higher: 100, Lower: 0, Number: 100, State: 1},
					Num2:   ValueFloat{Higher: 100, Lower: 0, Number: 100, State: 1},
					Num3:   ValueFloatLower{Higher: 100, Lower: 0.01, Number: 2, State: 0},
					Num4:   ValueFloat{},
					Source: source,
				})
			}

			bindings = append(bindings, binding)

		case SliderModeOpacity:
			bindings = append(bindings, Binding{
				Actions: []Action{{Category: 11, Name: bindingName, Sub: 4, Type: 0,
					Action: ValueString{},
					Json: JSON{
						Brightness:    ValueInt{Higher: 1, Lower: -1, Number: 0, State: 3},
						ColorAdd:      Color{},
						ColorMultiply: Color{Color: 16777215},
						Contrast:      ValueInt{Higher: 4, Lower: -4, Number: 0, State: 3},
						Gamma:         ValueInt{Higher: 3, Lower: -3, Number: 0, State: 3},
						HueShift:      ValueInt{Higher: 100, Lower: -100, Number: 0, State: 3},
						Opacity:       ValueInt{Higher: 1, Lower: 0, Number: 0, State: 2},
						Saturation:    ValueInt{Higher: 5, Lower: -1, Number: 0, State: 3},
					},
					Source: source,
					Num:    ValueInt{},
					Filter: defaultOpacityFilter,
				}},
				Enabled: true,
				Messages: []Message{{
					Channel: channel1,
					Device:  midiDevice,
					Name:    bindingName,
					Note:    ValueInt{Higher: 127, Lower: 0, Number: note, State: 0},
					Type:    typeCC,
					Value:   ValueInt{Higher: 127, Lower: 0, Number: 0, State: 1},
				}},
				Name:      bindingName,
				ResetMode: 0,
				Type:      0,
			})
		case SliderModeCropTop, SliderModeCropBottom, SliderModeCropLeft, SliderModeCropRight:
			var (
				num1State = 3
				num2State = 3
				num3State = 3
				num4State = 3
			)
			switch name {
			case SliderModeCropTop:
				num1State = 1
			case SliderModeCropBottom:
				num3State = 1
			case SliderModeCropLeft:
				num4State = 1
			default:
				num2State = 1
			}

			binding := Binding{
				Actions: make([]Action, 0, 32),
				Enabled: true,
				Messages: []Message{{
					Channel: channel1,
					Device:  midiDevice,
					Name:    bindingName,
					Note:    ValueInt{Higher: 127, Lower: 0, Number: note, State: 0},
					Type:    typeCC,
					Value:   ValueInt{Higher: 127, Lower: 0, Number: 0, State: 1},
				}},
				Name:      bindingName,
				ResetMode: 0,
				Type:      0,
			}

			for i := range sceneNames {
				binding.Actions = append(binding.Actions, Action{
					Category: 7, Name: fmt.Sprintf("CC %s %s", name, sceneNames[i]),
					Scene: sceneNames[i], Sub: 3, Type: 0,
					Action: ValueString{}, Json: JSON{},
					Num1:   ValueFloat{Higher: 1440, Lower: 0, Number: 0, State: num1State},
					Num2:   ValueFloat{Higher: 2560, Lower: 0, Number: 0, State: num2State},
					Num3:   ValueFloatLower{Higher: 1440, Lower: 0, Number: 0, State: num3State},
					Num4:   ValueFloat{Higher: 2560, Lower: 0, Number: 0, State: num4State},
					Source: source,
				})
			}

			bindings = append(bindings, binding)
		}
	}

	return bindings
}
