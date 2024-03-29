package vehicles

type Handling struct {
	Id                             string   `json:"Id"`
	VehicleModels                  []string `json:"VehicleModels"`
	Mass                           float64  `json:"Mass"`
	InitialDragCoeff               float64  `json:"InitialDragCoeff"`
	PercentSubmerged               float64  `json:"PercentSubmerged"`
	CentreOfMassOffset             Coord    `json:"CentreOfMassOffset"`
	InertiaMultiplier              Coord    `json:"InertiaMultiplier"`
	DriveBiasFront                 float64  `json:"DriveBiasFront"`
	InitialDriveGears              int      `json:"InitialDriveGears"`
	InitialDriveForce              float64  `json:"InitialDriveForce"`
	DriveInertia                   float64  `json:"DriveInertia"`
	ClutchChangeRateScaleUpShift   float64  `json:"ClutchChangeRateScaleUpShift"`
	ClutchChangeRateScaleDownShift float64  `json:"ClutchChangeRateScaleDownShift"`
	InitialDriveMaxFlatVel         float64  `json:"InitialDriveMaxFlatVel"`
	BrakeForce                     float64  `json:"BrakeForce"`
	BrakeBiasFront                 float64  `json:"BrakeBiasFront"`
	HandBrakeForce                 float64  `json:"HandBrakeForce"`
	SteeringLock                   float64  `json:"SteeringLock"`
	TractionCurveMax               float64  `json:"TractionCurveMax"`
	TractionCurveMin               float64  `json:"TractionCurveMin"`
	TractionCurveLateral           float64  `json:"TractionCurveLateral"`
	TractionSpringDeltaMax         float64  `json:"TractionSpringDeltaMax"`
	LowSpeedTractionLossMult       float64  `json:"LowSpeedTractionLossMult"`
	CamberStiffnesss               float64  `json:"CamberStiffnesss"`
	TractionBiasFront              float64  `json:"TractionBiasFront"`
	TractionLossMult               float64  `json:"TractionLossMult"`
	SuspensionForce                float64  `json:"SuspensionForce"`
	SuspensionCompDamp             float64  `json:"SuspensionCompDamp"`
	SuspensionReboundDamp          float64  `json:"SuspensionReboundDamp"`
	SuspensionUpperLimit           float64  `json:"SuspensionUpperLimit"`
	SuspensionLowerLimit           float64  `json:"SuspensionLowerLimit"`
	SuspensionRaise                float64  `json:"SuspensionRaise"`
	SuspensionBiasFront            float64  `json:"SuspensionBiasFront"`
	AntiRollBarForce               float64  `json:"AntiRollBarForce"`
	AntiRollBarBiasFront           float64  `json:"AntiRollBarBiasFront"`
	RollCentreHeightFront          float64  `json:"RollCentreHeightFront"`
	RollCentreHeightRear           float64  `json:"RollCentreHeightRear"`
	CollisionDamageMult            float64  `json:"CollisionDamageMult"`
	WeaponDamageMult               float64  `json:"WeaponDamageMult"`
	DeformationDamageMult          float64  `json:"DeformationDamageMult"`
	EngineDamageMult               float64  `json:"EngineDamageMult"`
	PetrolTankVolume               float64  `json:"PetrolTankVolume"`
	OilVolume                      float64  `json:"OilVolume"`
	SeatOffsetDistX                float64  `json:"SeatOffsetDistX"`
	SeatOffsetDistY                float64  `json:"SeatOffsetDistY"`
	SeatOffsetDistZ                float64  `json:"SeatOffsetDistZ"`
	MonetaryValue                  int      `json:"MonetaryValue"`
	AiHandling                     string   `json:"AiHandling"`
}

type Coord struct {
	X float64 `json:"X"`
	Y float64 `json:"Y"`
	Z float64 `json:"Z"`
}
