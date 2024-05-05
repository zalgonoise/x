package pokemon

import (
	"time"
)

const (
	defaultTimeout = time.Minute

	baseURLFormat = "https://pokeapi.co/api/v2/pokemon/%d"

	spriteURLFormat = "https://img.pokemondb.net/artwork/large/%s.jpg"
)

type Summary struct {
	ID     int
	Name   string
	Sprite string
}

type Pokemon struct {
	Abilities              []Ability     `json:"abilities"`
	BaseExperience         int           `json:"base_experience"`
	Cries                  Cries         `json:"cries"`
	Forms                  []Form        `json:"forms"`
	GameIndices            []GameIndex   `json:"game_indices"`
	Height                 int           `json:"height"`
	HeldItems              []interface{} `json:"held_items"`
	Id                     int           `json:"id"`
	IsDefault              bool          `json:"is_default"`
	LocationAreaEncounters string        `json:"location_area_encounters"`
	Moves                  []MoveDetails `json:"moves"`
	Name                   string        `json:"name"`
	Order                  int           `json:"order"`
	PastAbilities          []interface{} `json:"past_abilities"`
	PastTypes              []interface{} `json:"past_types"`
	Species                Species       `json:"species"`
	Sprites                Sprites       `json:"sprites"`
	Stats                  []StatSet     `json:"stats"`
	Types                  []TypeSet     `json:"types"`
	Weight                 int           `json:"weight"`
}

type AbilityType struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type Ability struct {
	Ability  AbilityType `json:"ability"`
	IsHidden bool        `json:"is_hidden"`
	Slot     int         `json:"slot"`
}

type Cries struct {
	Latest string `json:"latest"`
	Legacy string `json:"legacy"`
}

type Form struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type Version struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type Move struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type MoveLearnMethod struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type VersionGroup struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type VersionGroupDetails struct {
	LevelLearnedAt  int             `json:"level_learned_at"`
	MoveLearnMethod MoveLearnMethod `json:"move_learn_method"`
	VersionGroup    VersionGroup    `json:"version_group"`
}

type MoveDetails struct {
	Move                Move                  `json:"move"`
	VersionGroupDetails []VersionGroupDetails `json:"version_group_details"`
}

type Species struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type GameIndex struct {
	GameIndex int     `json:"game_index"`
	Version   Version `json:"version"`
}

type DreamWorld struct {
	FrontDefault string      `json:"front_default"`
	FrontFemale  interface{} `json:"front_female"`
}

type Home struct {
	FrontDefault     string      `json:"front_default"`
	FrontFemale      interface{} `json:"front_female"`
	FrontShiny       string      `json:"front_shiny"`
	FrontShinyFemale interface{} `json:"front_shiny_female"`
}

type OfficialArtwork struct {
	FrontDefault string `json:"front_default"`
	FrontShiny   string `json:"front_shiny"`
}

type Showdown struct {
	BackDefault      string      `json:"back_default"`
	BackFemale       interface{} `json:"back_female"`
	BackShiny        string      `json:"back_shiny"`
	BackShinyFemale  interface{} `json:"back_shiny_female"`
	FrontDefault     string      `json:"front_default"`
	FrontFemale      interface{} `json:"front_female"`
	FrontShiny       string      `json:"front_shiny"`
	FrontShinyFemale interface{} `json:"front_shiny_female"`
}

type Other struct {
	DreamWorld      DreamWorld      `json:"dream_world"`
	Home            Home            `json:"home"`
	OfficialArtwork OfficialArtwork `json:"official-artwork"`
	Showdown        Showdown        `json:"showdown"`
}

type RedBlue struct {
	BackDefault      string `json:"back_default"`
	BackGray         string `json:"back_gray"`
	BackTransparent  string `json:"back_transparent"`
	FrontDefault     string `json:"front_default"`
	FrontGray        string `json:"front_gray"`
	FrontTransparent string `json:"front_transparent"`
}

type Yellow struct {
	BackDefault      string `json:"back_default"`
	BackGray         string `json:"back_gray"`
	BackTransparent  string `json:"back_transparent"`
	FrontDefault     string `json:"front_default"`
	FrontGray        string `json:"front_gray"`
	FrontTransparent string `json:"front_transparent"`
}

type GenerationI struct {
	RedBlue RedBlue `json:"red-blue"`
	Yellow  Yellow  `json:"yellow"`
}

type Crystal struct {
	BackDefault           string `json:"back_default"`
	BackShiny             string `json:"back_shiny"`
	BackShinyTransparent  string `json:"back_shiny_transparent"`
	BackTransparent       string `json:"back_transparent"`
	FrontDefault          string `json:"front_default"`
	FrontShiny            string `json:"front_shiny"`
	FrontShinyTransparent string `json:"front_shiny_transparent"`
	FrontTransparent      string `json:"front_transparent"`
}

type Gold struct {
	BackDefault      string `json:"back_default"`
	BackShiny        string `json:"back_shiny"`
	FrontDefault     string `json:"front_default"`
	FrontShiny       string `json:"front_shiny"`
	FrontTransparent string `json:"front_transparent"`
}

type Silver struct {
	BackDefault      string `json:"back_default"`
	BackShiny        string `json:"back_shiny"`
	FrontDefault     string `json:"front_default"`
	FrontShiny       string `json:"front_shiny"`
	FrontTransparent string `json:"front_transparent"`
}

type GenerationIi struct {
	Crystal Crystal `json:"crystal"`
	Gold    Gold    `json:"gold"`
	Silver  Silver  `json:"silver"`
}

type Emerald struct {
	FrontDefault string `json:"front_default"`
	FrontShiny   string `json:"front_shiny"`
}

type FireredLeafgreen struct {
	BackDefault  string `json:"back_default"`
	BackShiny    string `json:"back_shiny"`
	FrontDefault string `json:"front_default"`
	FrontShiny   string `json:"front_shiny"`
}

type RubySapphire struct {
	BackDefault  string `json:"back_default"`
	BackShiny    string `json:"back_shiny"`
	FrontDefault string `json:"front_default"`
	FrontShiny   string `json:"front_shiny"`
}

type GenerationIii struct {
	Emerald          Emerald          `json:"emerald"`
	FireredLeafgreen FireredLeafgreen `json:"firered-leafgreen"`
	RubySapphire     RubySapphire     `json:"ruby-sapphire"`
}

type DiamondPearl struct {
	BackDefault      string      `json:"back_default"`
	BackFemale       interface{} `json:"back_female"`
	BackShiny        string      `json:"back_shiny"`
	BackShinyFemale  interface{} `json:"back_shiny_female"`
	FrontDefault     string      `json:"front_default"`
	FrontFemale      interface{} `json:"front_female"`
	FrontShiny       string      `json:"front_shiny"`
	FrontShinyFemale interface{} `json:"front_shiny_female"`
}

type HeartgoldSoulsilver struct {
	BackDefault      string      `json:"back_default"`
	BackFemale       interface{} `json:"back_female"`
	BackShiny        string      `json:"back_shiny"`
	BackShinyFemale  interface{} `json:"back_shiny_female"`
	FrontDefault     string      `json:"front_default"`
	FrontFemale      interface{} `json:"front_female"`
	FrontShiny       string      `json:"front_shiny"`
	FrontShinyFemale interface{} `json:"front_shiny_female"`
}

type Platinum struct {
	BackDefault      string      `json:"back_default"`
	BackFemale       interface{} `json:"back_female"`
	BackShiny        string      `json:"back_shiny"`
	BackShinyFemale  interface{} `json:"back_shiny_female"`
	FrontDefault     string      `json:"front_default"`
	FrontFemale      interface{} `json:"front_female"`
	FrontShiny       string      `json:"front_shiny"`
	FrontShinyFemale interface{} `json:"front_shiny_female"`
}

type GenerationIv struct {
	DiamondPearl        DiamondPearl        `json:"diamond-pearl"`
	HeartgoldSoulsilver HeartgoldSoulsilver `json:"heartgold-soulsilver"`
	Platinum            Platinum            `json:"platinum"`
}

type Animated struct {
	BackDefault      string      `json:"back_default"`
	BackFemale       interface{} `json:"back_female"`
	BackShiny        string      `json:"back_shiny"`
	BackShinyFemale  interface{} `json:"back_shiny_female"`
	FrontDefault     string      `json:"front_default"`
	FrontFemale      interface{} `json:"front_female"`
	FrontShiny       string      `json:"front_shiny"`
	FrontShinyFemale interface{} `json:"front_shiny_female"`
}

type BlackWhite struct {
	Animated         Animated    `json:"animated"`
	BackDefault      string      `json:"back_default"`
	BackFemale       interface{} `json:"back_female"`
	BackShiny        string      `json:"back_shiny"`
	BackShinyFemale  interface{} `json:"back_shiny_female"`
	FrontDefault     string      `json:"front_default"`
	FrontFemale      interface{} `json:"front_female"`
	FrontShiny       string      `json:"front_shiny"`
	FrontShinyFemale interface{} `json:"front_shiny_female"`
}

type GenerationV struct {
	BlackWhite BlackWhite `json:"black-white"`
}

type OmegarubyAlphasapphire struct {
	FrontDefault     string      `json:"front_default"`
	FrontFemale      interface{} `json:"front_female"`
	FrontShiny       string      `json:"front_shiny"`
	FrontShinyFemale interface{} `json:"front_shiny_female"`
}

type XY struct {
	FrontDefault     string      `json:"front_default"`
	FrontFemale      interface{} `json:"front_female"`
	FrontShiny       string      `json:"front_shiny"`
	FrontShinyFemale interface{} `json:"front_shiny_female"`
}

type GenerationVi struct {
	OmegarubyAlphasapphire OmegarubyAlphasapphire `json:"omegaruby-alphasapphire"`
	XY                     XY                     `json:"x-y"`
}

type Icons struct {
	FrontDefault string      `json:"front_default"`
	FrontFemale  interface{} `json:"front_female"`
}

type UltraSunUltraMoon struct {
	FrontDefault     string      `json:"front_default"`
	FrontFemale      interface{} `json:"front_female"`
	FrontShiny       string      `json:"front_shiny"`
	FrontShinyFemale interface{} `json:"front_shiny_female"`
}

type GenerationVii struct {
	Icons             Icons             `json:"icons"`
	UltraSunUltraMoon UltraSunUltraMoon `json:"ultra-sun-ultra-moon"`
}

type GenerationViii struct {
	Icons Icons `json:"icons"`
}

type Versions struct {
	GenerationI    GenerationI    `json:"generation-i"`
	GenerationIi   GenerationIi   `json:"generation-ii"`
	GenerationIii  GenerationIii  `json:"generation-iii"`
	GenerationIv   GenerationIv   `json:"generation-iv"`
	GenerationV    GenerationV    `json:"generation-v"`
	GenerationVi   GenerationVi   `json:"generation-vi"`
	GenerationVii  GenerationVii  `json:"generation-vii"`
	GenerationViii GenerationViii `json:"generation-viii"`
}

type Sprites struct {
	BackDefault      string      `json:"back_default"`
	BackFemale       interface{} `json:"back_female"`
	BackShiny        string      `json:"back_shiny"`
	BackShinyFemale  interface{} `json:"back_shiny_female"`
	FrontDefault     string      `json:"front_default"`
	FrontFemale      interface{} `json:"front_female"`
	FrontShiny       string      `json:"front_shiny"`
	FrontShinyFemale interface{} `json:"front_shiny_female"`
	Other            Other       `json:"other"`
	Versions         Versions    `json:"versions"`
}

type Stat struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type StatSet struct {
	BaseStat int  `json:"base_stat"`
	Effort   int  `json:"effort"`
	Stat     Stat `json:"stat"`
}

type Type struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type TypeSet struct {
	Slot int  `json:"slot"`
	Type Type `json:"type"`
}
