package testdata

import (
	"github.com/zalgonoise/x/obs-midi"
)

const (
	DefaultHigherScene = "Camera Full"
	DefaultLowerScene  = "Fullscreen"

	DefaultSource = "ModBrowser"
)

var (
	//SceneNames = []string{
	//	"Base",
	//	"Banner Full", "Banner BL", "Banner BR", "Banner TR",
	//	"Chrome Full", "Chrome BL", "Chrome BR", "2Way Game + Chrome", "2Way Game + Chrome L", "2Way Game + Chrome R",
	//	"Monitor5 Full", "Mon5 BL", "Mon5 BR",
	//	"Chrome 2 Full", "2Way Game + Firefox", "2Way Game + Firefox L", "2Way Game + Firefox R",
	//	"Discord Full", "2Way Game + Discord", "2Way Game + Discord L", "2Way Game + Discord R",
	//	"2Way Chrome + Discord", "2Way Chrome + Discord L", "2Way Chrome + Discord R",
	//	"Base w/ opacity chrome", "Camera Full",
	//}

	//SceneSet = map[string]int{
	//	"Base":      56,
	//	"Blank #57": 57,
	//	"Banner TR": 58,
	//	"Blank #59": 59,
	//	"Blank #60": 60,
	//	"Blank #61": 61,
	//	"Blank #62": 62,
	//
	//	"Banner Full":             48,
	//	"Banner BL":               49,
	//	"Banner BR":               50,
	//	"2Way Chrome + Discord":   51,
	//	"2Way Chrome + Discord L": 52,
	//	"2Way Chrome + Discord R": 53,
	//	"Blank #54":               54,
	//
	//	"Chrome Full":          40,
	//	"Chrome BL":            41,
	//	"Chrome BR":            42,
	//	"2Way Game + Chrome":   43,
	//	"2Way Game + Chrome L": 44,
	//	"2Way Game + Chrome R": 45,
	//	"Blank #46":            46,
	//
	//	"Monitor5 Full":      32,
	//	"Mon5 BL":            33,
	//	"Mon5 BR":            34,
	//	"2Way Game + Mon5":   35,
	//	"2Way Game + Mon5 L": 36,
	//	"2Way Game + Mon5 R": 37,
	//	"Blank #30":          38,
	//
	//	"Chrome 2 Full":         24,
	//	"Blank #33":             25,
	//	"Blank #34":             26,
	//	"2Way Game + Firefox":   27,
	//	"2Way Game + Firefox L": 28,
	//	"2Way Game + Firefox R": 29,
	//	"Blank #38":             30,
	//
	//	"Discord Full":          16,
	//	"Blank #17":             17,
	//	"Blank #18":             18,
	//	"2Way Game + Discord":   19,
	//	"2Way Game + Discord L": 20,
	//	"2Way Game + Discord R": 21,
	//	"Blank #22":             22,
	//}

	SceneNames = []string{
		"Fullscreen",
		"Corner BL", "Corner BR",
		"2 Way", "2 Way Left", "2 Way Right",
	}

	SceneSet = map[string]int{
		"Fullscreen":  56,
		"Corner BL":   57,
		"Corner BR":   58,
		"2 Way":       59,
		"2 Way Left":  60,
		"2 Way Right": 61,
		"Blank #62":   62,
		"Blank #63":   63,

		"Blank #48":        48,
		"Blank #49":        49,
		"Blank #50":        50,
		"3 Way Horizontal": 51,
		"3 Way Left":       52,
		"3 Way Right":      53,
		"Blank #54":        54,
		"Blank #55":        55,
	}

	NoSecondarySources = []string{"Fullscreen"}

	NoTetriarySources = []string{
		"Fullscreen",
		"Corner BL",
		"Corner BR",
		"2 Way",
		"2 Way Left",
		"2 Way Right",
	}

	Base = "Banner"

	PrimarySources = []string{
		"Banner",
		"Screen",
		"Monitor 3",
		"Monitor 5",
		"Discord Stream",
		"iPad",
		"Camera Full",
		"Camera Desk Full",
	}

	PrimarySourceSet = map[string]int{
		"Banner":           40,
		"Screen":           41,
		"Monitor 3":        42,
		"Monitor 5":        43,
		"Discord Stream":   44,
		"iPad":             45,
		"Camera Full":      46,
		"Camera Desk Full": 47,
	}

	SecondarySources = []string{
		"Banner",
		"Screen B",
		"Monitor 3 B",
		"Monitor 5 B",
		"Discord Stream B",
		"iPad B",
		"Camera Full B",
		"Camera Desk Full B",
	}

	SecondarySourceSet = map[string]int{
		"Banner":             32,
		"Screen B":           33,
		"Monitor 3 B":        34,
		"Monitor 5 B":        35,
		"Discord Stream B":   36,
		"iPad B":             37,
		"Camera Full B":      38,
		"Camera Desk Full B": 39,
	}

	TetriarySourceSet = map[string]int{
		"Banner":             24,
		"Screen C":           25,
		"Monitor 3 C":        26,
		"Monitor 5 C":        27,
		"Discord Stream C":   28,
		"iPad C":             29,
		"Camera Full C":      30,
		"Camera Desk Full C": 31,
	}

	TogglesSet = map[string]obs_midi.SourceNote{
		"Task Manager On":  {State: true, Source: "Task Manager", NoteOn: 8, NoteOff: 0},
		"Task Manager Off": {State: false, Source: "Task Manager", NoteOn: 0, NoteOff: 8},
		"Cam Self On":      {State: true, Source: "DroidCam OBS Self", NoteOn: 9, NoteOff: 1},
		"Cam Self Off":     {State: false, Source: "DroidCam OBS Self", NoteOn: 1, NoteOff: 9},
		"Cam Desk On":      {State: true, Source: "DroidCam OBS Desk", NoteOn: 10, NoteOff: 2},
		"Cam Desk Off":     {State: false, Source: "DroidCam OBS Desk", NoteOn: 2, NoteOff: 10},
		"ModBrowser On":    {State: true, Source: "ModBrowser", NoteOn: 11, NoteOff: 3},
		"ModBrowser Off":   {State: false, Source: "ModBrowser", NoteOn: 3, NoteOff: 11},
		"Vtube Studio On":  {State: true, Source: "vtube studio", NoteOn: 12, NoteOff: 4},
		"Vtube Studio Off": {State: false, Source: "vtube studio", NoteOn: 4, NoteOff: 12},
		"Xenise On":        {State: true, Source: "Xenise", NoteOn: 13, NoteOff: 5},
		"Xenise Off":       {State: false, Source: "Xenise", NoteOn: 5, NoteOff: 13},

		"Blank #6": {NoteOn: 6},

		"Blank #16": {NoteOn: 16},
		"Blank #17": {NoteOn: 17},
		"Blank #18": {NoteOn: 18},
		"Blank #19": {NoteOn: 19},
		"Blank #20": {NoteOn: 20},
		"Blank #21": {NoteOn: 21},
	}

	TransitionsSet = map[string]int{
		obs_midi.TransitionToProgram: 7,
		"Move":                       15,
		"Cut":                        23,
		"Fade":                       14,
		"Fade to Black":              22,
	}

	SliderSet = map[string]int{
		obs_midi.SliderModeMoveX:      48,
		obs_midi.SliderModeMoveY:      49,
		obs_midi.SliderModeScale:      50,
		obs_midi.SliderModeOpacity:    51,
		obs_midi.SliderModeCropTop:    52,
		obs_midi.SliderModeCropBottom: 53,
		obs_midi.SliderModeCropLeft:   54,
		obs_midi.SliderModeCropRight:  55,
		obs_midi.SliderModeTransition: 56,
	}

	ControlSet = map[string]obs_midi.SourceNote{
		obs_midi.ControlModeSaveReplayBuffer: {NoteOn: 119},
		obs_midi.ControlModeStudioModeOn:     {State: true, NoteOn: 114, NoteOff: 115},
		obs_midi.ControlModeStudioModeOff:    {State: false, NoteOn: 115, NoteOff: 114},
		obs_midi.ControlModeLEDOn:            {State: true, NoteOn: 112, NoteOff: 113},
		obs_midi.ControlModeLEDOff:           {State: false, NoteOn: 113, NoteOff: 112},
	}
)
