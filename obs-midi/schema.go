package obs_midi

const (
	minCh5Note = 0
	maxCh5Note = 63
	minCh1Note = 100
	maxCh1Note = 127

	blankPrefix = "Blank"

	defaultTransition    = "Cut"
	defaultOpacityFilter = "Color Correction"

	TransitionToProgram = "Transition to Program"

	toggleStateHide = "Hide"
	toggleStateShow = "Show"

	SliderModeMoveX      = "Move X"
	SliderModeMoveY      = "Move Y"
	SliderModeScale      = "Scale"
	SliderModeOpacity    = "Opacity"
	SliderModeCropTop    = "Crop Top"
	SliderModeCropBottom = "Crop Bottom"
	SliderModeCropLeft   = "Crop Left"
	SliderModeCropRight  = "Crop Right"
	SliderModeTransition = "Transition"

	ControlModeSaveReplayBuffer = "Save Replay Buffer"
	ControlModeStudioModeOn     = "Enable Studio Mode"
	ControlModeStudioModeOff    = "Disable Studio Mode"
	ControlModeLEDOn            = "Enable LEDs"
	ControlModeLEDOff           = "Disable LEDs"
)

var (
	SupportedSliderModes = []string{
		SliderModeMoveX,
		SliderModeMoveY,
		SliderModeScale,
		SliderModeOpacity,
		SliderModeCropTop,
		SliderModeCropBottom,
		SliderModeCropLeft,
		SliderModeCropRight,
		SliderModeTransition,
	}

	SupportedControlModes = []string{
		ControlModeSaveReplayBuffer,
		ControlModeStudioModeOn,
		ControlModeStudioModeOff,
		ControlModeLEDOn,
		ControlModeLEDOff,
	}
)

type SourceNote struct {
	State   bool   `json:"state"`
	Source  string `json:"source"`
	NoteOn  int    `json:"note_on"`
	NoteOff int    `json:"note_off"`
}

var (
	channel1 = ValueInt{Higher: 16, Lower: 1, Number: 1, State: 0}
	channel5 = ValueInt{Higher: 16, Lower: 1, Number: 5, State: 0}

	typeNoteOn = ValueString{Higher: "", Lower: "", State: 0, String: "Note On"}
	typeCC     = ValueString{Higher: "", Lower: "", State: 0, String: "Control Change"}
)

type ConfigMap struct {
	Collections []Collection `json:"collections"`
	Devices     []Devices    `json:"devices"`
	Savedate    string       `json:"savedate"`
	Version     string       `json:"version"`
}

type Collection struct {
	Bindings []Binding `json:"bindings"`
	Name     string    `json:"name"`
}

type Binding struct {
	Actions   []Action  `json:"actions"`
	Enabled   bool      `json:"enabled"`
	Messages  []Message `json:"messages"`
	Name      string    `json:"name"`
	ResetMode int       `json:"reset_mode"`
	Type      int       `json:"type"`
}

type Action struct {
	Category   int             `json:"category"`
	Name       string          `json:"name"`
	Scene      interface{}     `json:"scene,omitempty"`
	Sub        int             `json:"sub"`
	Type       int             `json:"type"`
	Messages   []Message       `json:"messages,omitempty"`
	Action     ValueString     `json:"action,omitempty"`
	Json       JSON            `json:"json,omitempty"`
	Num1       ValueFloat      `json:"num1,omitempty"`
	Num2       ValueFloat      `json:"num2,omitempty"`
	Num3       ValueFloatLower `json:"num3,omitempty"`
	Num4       ValueFloat      `json:"num4,omitempty"`
	Source     string          `json:"source,omitempty"`
	Num        ValueInt        `json:"num,omitempty"`
	Transition string          `json:"transition,omitempty"`
	Filter     string          `json:"filter,omitempty"`
}

type JSON struct {
	Brightness    ValueInt `json:"brightness,omitempty"`
	ColorAdd      Color    `json:"color_add,omitempty"`
	ColorMultiply Color    `json:"color_multiply,omitempty"`
	Contrast      ValueInt `json:"contrast,omitempty"`
	Gamma         ValueInt `json:"gamma,omitempty"`
	HueShift      ValueInt `json:"hue_shift,omitempty"`
	Opacity       ValueInt `json:"opacity,omitempty"`
	Saturation    ValueInt `json:"saturation,omitempty"`
}

type Color struct {
	Color int `json:"color"`
}

type Message struct {
	Channel ValueInt    `json:"channel"`
	Device  string      `json:"device"`
	Name    string      `json:"name"`
	Note    ValueInt    `json:"note"`
	Type    ValueString `json:"type"`
	Value   ValueInt    `json:"value"`
}

type ValueString struct {
	Higher string `json:"higher"`
	Lower  string `json:"lower"`
	State  int    `json:"state"`
	String string `json:"string"`
}

type ValueInt struct {
	Higher int `json:"higher"`
	Lower  int `json:"lower"`
	Number int `json:"number"`
	State  int `json:"state"`
}

type ValueFloat struct {
	Higher float64 `json:"higher"`
	Lower  float64 `json:"lower"`
	Number float64 `json:"number"`
	State  int     `json:"state"`
}

type ValueFloatLower struct {
	Higher float64 `json:"higher"`
	Lower  float64 `json:"lower"`
	Number float64 `json:"number"`
	State  int     `json:"state"`
}

type Devices struct {
	Active int    `json:"active"`
	Name   string `json:"name"`
}
