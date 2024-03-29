package vehicles

type VehicleList struct {
	Vehicles []Vehicle `json:"vehicles"`
}

type Vehicle struct {
	Name                    string          `json:"Name"`
	DisplayName             DisplayName     `json:"DisplayName"`
	Hash                    int64           `json:"Hash"`
	SignedHash              int             `json:"SignedHash"`
	HexHash                 string          `json:"HexHash"`
	DlcName                 string          `json:"DlcName"`
	HandlingId              string          `json:"HandlingId"`
	LayoutId                string          `json:"LayoutId"`
	Manufacturer            string          `json:"Manufacturer"`
	ManufacturerDisplayName DisplayName     `json:"ManufacturerDisplayName"`
	Class                   string          `json:"Class"`
	ClassId                 int             `json:"ClassId"`
	Type                    string          `json:"Type"`
	PlateType               string          `json:"PlateType"`
	DashboardType           *string         `json:"DashboardType"`
	WheelType               *string         `json:"WheelType"`
	Flags                   []string        `json:"Flags"`
	Seats                   int             `json:"Seats"`
	Price                   int             `json:"Price"`
	MonetaryValue           int             `json:"MonetaryValue"`
	HasConvertibleRoof      bool            `json:"HasConvertibleRoof"`
	HasSirens               bool            `json:"HasSirens"`
	Weapons                 []string        `json:"Weapons"`
	ModKits                 []string        `json:"ModKits"`
	DimensionsMin           *Dimensions     `json:"DimensionsMin"`
	DimensionsMax           *Dimensions     `json:"DimensionsMax"`
	BoundingCenter          BoundingCenter  `json:"BoundingCenter"`
	BoundingSphereRadius    float64         `json:"BoundingSphereRadius"`
	Rewards                 []string        `json:"Rewards"`
	MaxBraking              float64         `json:"MaxBraking"`
	MaxBrakingMods          float64         `json:"MaxBrakingMods"`
	MaxSpeed                float64         `json:"MaxSpeed"`
	MaxTraction             float64         `json:"MaxTraction"`
	Acceleration            float64         `json:"Acceleration"`
	Agility                 float64         `json:"Agility"`
	MaxKnots                float64         `json:"MaxKnots"`
	MoveResistance          float64         `json:"MoveResistance"`
	HasArmoredWindows       bool            `json:"HasArmoredWindows"`
	DefaultColors           []DefaultColors `json:"DefaultColors"`
	DefaultBodyHealth       float64         `json:"DefaultBodyHealth"`
	DirtLevelMin            float64         `json:"DirtLevelMin"`
	DirtLevelMax            float64         `json:"DirtLevelMax"`
	Trailers                []string        `json:"Trailers"`
	AdditionalTrailers      []string        `json:"AdditionalTrailers"`
	Extras                  []int           `json:"Extras"`
	RequiredExtras          []int           `json:"RequiredExtras"`
	SpawnFrequency          float64         `json:"SpawnFrequency"`
	WheelsCount             int             `json:"WheelsCount"`
	HasParachute            bool            `json:"HasParachute"`
	HasKers                 bool            `json:"HasKers"`
	DefaultHorn             int64           `json:"DefaultHorn"`
	DefaultHornVariation    int             `json:"DefaultHornVariation"`
	Bones                   []Bones         `json:"Bones"`
}

type DisplayName struct {
	Hash               int64   `json:"Hash"`
	English            *string `json:"English"`
	German             *string `json:"German"`
	French             *string `json:"French"`
	Italian            *string `json:"Italian"`
	Russian            *string `json:"Russian"`
	Polish             *string `json:"Polish"`
	Name               *string `json:"Name"`
	TraditionalChinese *string `json:"TraditionalChinese"`
	SimplifiedChinese  *string `json:"SimplifiedChinese"`
	Spanish            *string `json:"Spanish"`
	Japanese           *string `json:"Japanese"`
	Korean             *string `json:"Korean"`
	Portuguese         *string `json:"Portuguese"`
	Mexican            *string `json:"Mexican"`
}

type Dimensions struct {
	X float64 `json:"X"`
	Y float64 `json:"Y"`
	Z float64 `json:"Z"`
}

type BoundingCenter struct {
	X float64 `json:"X"`
	Y float64 `json:"Y"`
	Z float64 `json:"Z"`
}

type DefaultColors struct {
	DefaultPrimaryColor   int `json:"DefaultPrimaryColor"`
	DefaultSecondaryColor int `json:"DefaultSecondaryColor"`
	DefaultPearlColor     int `json:"DefaultPearlColor"`
	DefaultWheelsColor    int `json:"DefaultWheelsColor"`
	DefaultInteriorColor  int `json:"DefaultInteriorColor"`
	DefaultDashboardColor int `json:"DefaultDashboardColor"`
}

type Bones struct {
	BoneIndex int    `json:"BoneIndex"`
	BoneId    int    `json:"BoneId"`
	BoneName  string `json:"BoneName"`
}
