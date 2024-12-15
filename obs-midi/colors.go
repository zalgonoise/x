package obs_midi

type ColorValue int

const (
	ColorOff ColorValue = iota
	ColorGrayLow
	ColorGrayMed
	ColorGrayHigh
	ColorLightPinkMed
	ColorRedHigh
	ColorRedMed
	ColorRedLow
	ColorWhiteHigh
	ColorGoldHigh
	ColorGoldMed
	ColorGoldLow
	ColorLightGreenHigh
	ColorYellowGreenHigh
	ColorYellowGreenMed
	ColorYellowGreenLow
	ColorLightTealHigh
	ColorYellowishGreenHigh
	ColorYellowishGreenMed
	ColorYellowishGreenLow
	ColorTealHigh
	ColorGreenHigh
	ColorGreenMed
	ColorGreenLow
	ColorBluishTealHigh
	ColorBluishGreenHigh
	ColorBluishGreenMed
	ColorBluishGreenLow
	ColorGreenishBlueHigh
	ColorGreenBlueHigh
	ColorGreenBlueMed
	ColorGreenBlueLow
	ColorLightBlueHigh
	ColorSkyBlueHigh
	ColorSkyBlueMed
	ColorSkyBlueLow
	ColorLightVioletHigh
	ColorFadedLightBlueHigh
	ColorFadedBlueMed
	ColorFadedBlueLow
	ColorBluishVioletHigh
	ColorBlueHigh
	ColorBlueMed
	ColorBlueLow
	ColorLightPurpleHigh
	ColorNavyBlueHigh
	ColorNavyBlueMed
	ColorNavyBlueLow
	ColorVioletHigh
	ColorPurpleHigh
	ColorPurpleMed
	ColorPurpleLow
	ColorPurplishPinkHigh
	ColorPinkHigh
	ColorPinkMed
	ColorPinkLow
	ColorBrightPinkHigh
	ColorShockPinkHigh
	ColorShockPinkMed
	ColorShockPinkLow
	ColorBrightOrangeHigh
	ColorFadedOrangeHigh
	ColorFadedGoldHigh
	ColorFadedYellowGreenHigh
	ColorFadedGreenLow
	ColorFadedBlueGreenHigh
	ColorFadedBluishGreenHigh
	ColorFadedBlueHigh
	ColorFadedTealHigh
	ColorFadedPurpleHigh
	ColorFadedWhiteMed
	ColorFadedGrayLow
	ColorBrightRedHigh
	ColorLightOliveHigh
	ColorLightGreenishYellowHigh
	ColorLightYellowishGreenHigh
	ColorDarkGreenHigh
	ColorBrightTealHigh
	ColorBrightPurplishBlueHigh
	ColorBrightBlueHigh
	ColorBrightPurpleHigh
	ColorBrightLightPurpleHigh
	ColorBrightHigh
	ColorFadedGoldMed
	ColorBrightGoldHigh
	ColorBrightYellowGreenHigh
	ColorBrightLightYellowishGreenHigh
	ColorBrightGreenHigh
	ColorBrightLightGreenHigh
	ColorBrightLightTealHigh
	ColorBrightLightBlueHigh
	ColorBrightBluishVioletHigh
	ColorBrightVioletHigh
	ColorBrightLightVioletHigh
	ColorBrightPurplishPinkHigh
	ColorBrightShockPinkHigh
	ColorBrightLightGoldHigh
	ColorBrightGreenYellowHigh
	ColorBrightYellowishGreenHigh
	ColorFadedLightGoldHigh
	ColorFadedLightGoldMed
	ColorFadedBlueishGreenHigh
	ColorFadedBlueGreenMed
	ColorFadedVioletLow
	ColorFadedBluishVioletHigh
	ColorLightBrownHigh
	ColorMagentaHigh
	ColorLightPinkHigh
	ColorLightGoldHigh
	ColorLightYellowHigh
	ColorFadedLightYellowishGreenHigh
	ColorLightYellowGreenHigh
	ColorDarkVioletHigh
	ColorFadedLightYellowGreenHigh
	ColorFadedGreenishBlueHigh
	ColorFadedVioletHigh
	ColorFadedVioletMed
	ColorFadedGrayHigh
	ColorFadedGrayMed
	ColorFadedWhiteHigh
	ColorFadedRedHigh
	ColorFadedRedMed
	ColorFadedGreenHigh
	ColorFadedGreenMed
	ColorFadedYellowHigh
	ColorFadedYellowMed
	ColorFadedLightOrangeHigh
	ColorFadedOrangeMed
)

type ColorSchema struct {
	Scenes      OnOffColor       `json:"scenes"`
	Sources     OnOffColor       `json:"sources"`
	Toggles     OnOffColor       `json:"toggles"`
	Transitions TransitionsColor `json:"transitions"`
}

type TransitionsColor struct {
	On         ColorValue `json:"on"`
	Off        ColorValue `json:"off"`
	Blank      ColorValue `json:"blank"`
	Transition ColorValue `json:"transition"`
}

type OnOffColor struct {
	On    ColorValue `json:"on"`
	Off   ColorValue `json:"off"`
	Blank ColorValue `json:"blank"`
}

func DefaultColorSchema() ColorSchema {
	return ColorSchema{
		Scenes: OnOffColor{
			On:    ColorBrightLightBlueHigh,
			Off:   ColorPurplishPinkHigh,
			Blank: ColorBrightHigh,
		},
		Sources: OnOffColor{
			On:    ColorBrightYellowishGreenHigh,
			Off:   ColorBrightLightBlueHigh,
			Blank: ColorFadedGrayMed,
		},
		Toggles: OnOffColor{
			On:    ColorFadedBlueMed,
			Off:   ColorBrightLightPurpleHigh,
			Blank: ColorFadedGrayMed,
		},
		Transitions: TransitionsColor{
			On:         ColorFadedLightOrangeHigh,
			Off:        ColorLightPinkMed,
			Blank:      ColorSkyBlueMed,
			Transition: ColorRedHigh,
		},
	}
}

func GreenColorSchema() ColorSchema {
	return ColorSchema{
		Scenes: OnOffColor{
			On:    ColorYellowGreenHigh,
			Off:   ColorBrightGoldHigh,
			Blank: ColorLightGoldHigh,
		},
		Sources: OnOffColor{
			On:    ColorShockPinkHigh,
			Off:   ColorBrightLightTealHigh,
			Blank: ColorFadedLightGoldMed,
		},
		Toggles: OnOffColor{
			On:    ColorFadedLightYellowGreenHigh,
			Off:   ColorFadedYellowHigh,
			Blank: ColorFadedLightGoldMed,
		},
		Transitions: TransitionsColor{
			On:         ColorBrightGreenHigh,
			Off:        ColorFadedLightYellowishGreenHigh,
			Blank:      ColorSkyBlueMed,
			Transition: ColorRedHigh,
		},
	}
}
