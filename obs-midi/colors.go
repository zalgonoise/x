package obs_midi

type ColorSchema struct {
	Scenes      OnOffColor       `json:"scenes"`
	Sources     OnOffColor       `json:"sources"`
	Toggles     OnOffColor       `json:"toggles"`
	Transitions TransitionsColor `json:"transitions"`
}

type TransitionsColor struct {
	On         int `json:"on"`
	Off        int `json:"off"`
	Blank      int `json:"blank"`
	Transition int `json:"transition"`
}

type OnOffColor struct {
	On    int `json:"on"`
	Off   int `json:"off"`
	Blank int `json:"blank"`
}

func DefaultColorSchema() ColorSchema {
	return ColorSchema{
		Scenes: OnOffColor{
			On:    90,
			Off:   52,
			Blank: 82,
		},
		Sources: OnOffColor{
			On:    98,
			Off:   90,
			Blank: 118,
		},
		Toggles: OnOffColor{
			On:    38,
			Off:   81,
			Blank: 118,
		},
		Transitions: TransitionsColor{
			On:         126,
			Off:        4,
			Blank:      34,
			Transition: 5,
		},
	}
}

func GreenColorSchema() ColorSchema {
	return ColorSchema{
		Scenes: OnOffColor{
			On:    13,
			Off:   84,
			Blank: 108,
		},
		Sources: OnOffColor{
			On:    57,
			Off:   89,
			Blank: 100,
		},
		Toggles: OnOffColor{
			On:    113,
			Off:   124,
			Blank: 100,
		},
		Transitions: TransitionsColor{
			On:         87,
			Off:        110,
			Blank:      34,
			Transition: 5,
		},
	}
}
