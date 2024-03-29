package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/zalgonoise/x/torque/vehicles"
)

const (
	insertVehicleQuery = `
	INSERT INTO vehicles (
		name, hash, signed_hash, hex_hash, dlc_name, handling_id, layout_id, manufacturer, class,
		class_id, type, plate_type, dashboard_type, wheel_type, seats, price, monetary_value,
		has_convertible_roof, has_sirens, bounding_sphere_radius, max_braking, max_braking_mods, max_speed,
		max_traction, acceleration, agility, max_knots, move_resistance, has_armored_windows,
		default_body_health, dirt_level_min, dirt_level_max, spawn_frequency, wheels_count, has_parachute,
		has_kers, default_horn, default_horn_variation
	) VALUES (
		?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
	) RETURNING id
`

	insertTableQuery = `
	INSERT INTO %s (%s) VALUES (%s);
`
)

var (
	vehiclesDisplayNamesCols = []string{
		"vehicle_id", "hash", "english", "german", "french", "italian", "russian", "polish", "name", "traditional_chinese",
		"simplified_chinese", "spanish", "japanese", "korean", "portuguese", "mexican",
	}

	vehiclesFlagsCols = []string{"vehicle_id", "flag"}

	vehiclesWeaponsCols = []string{"vehicle_id", "weapon"}

	vehiclesModKitsCols = []string{"vehicle_id", "mod_kit"}

	vehiclesDimensionsCols = []string{"vehicle_id", "min_x", "min_y", "min_z", "max_x", "max_y", "max_z"}

	vehiclesRewardsCols = []string{"vehicle_id", "rewards"}

	vehiclesDefaultColorsCols = []string{
		"vehicle_id", "default_primary_color", "default_secondary_color", "default_pearl_color", "default_wheels_color",
		"default_interior_color", "default_dashboard_color",
	}

	vehiclesTrailersCols = []string{"vehicle_id", "trailers"}

	vehiclesAdditionalTrailersCols = []string{"vehicle_id", "additional_trailers"}

	vehiclesExtrasCols = []string{"vehicle_id", "extras"}

	vehiclesRequiredExtrasCols = []string{"vehicle_id", "required_extras"}

	vehiclesBonesCols = []string{"vehicle_id", "bone_index", "bone_id", "bone_name"}

	vehiclesHandlingCols = []string{
		"id", "mass", "initial_drag_coeff", "percent_submerged", "centre_of_mass_offset_x", "centre_of_mass_offset_y",
		"centre_of_mass_offset_z", "inertia_multiplier_x", "inertia_multiplier_y", "inertia_multiplier_z",
		"drive_bias_front", "initial_drive_gears", "initial_drive_force", "drive_inertia",
		"clutch_change_rate_scale_up_shift", "clutch_change_rate_scale_down_shift", "initial_drive_max_flat_vel",
		"brake_force", "brake_bias_front", "handbrake_force", "steering_lock", "traction_curve_max", "traction_curve_min",
		"traction_curve_lateral", "traction_spring_delta_max", "low_speed_traction_loss_mult", "camber_stiffness",
		"traction_bias_front", "traction_loss_mult", "suspension_force", "suspension_comp_damp", "suspension_rebound_damp",
		"suspension_upper_limit", "suspension_lower_limit", "suspension_raise", "suspension_bias_front",
		"anti_roll_bar_force", "anti_roll_bar_bias_front", "roll_centre_height_front", "roll_centre_height_rear",
		"collision_damage_mult", "weapon_damage_mult", "deformation_damage_mult", "engine_damage_mult",
		"petrol_tank_volume", "oil_volume", "seat_offset_dist_x", "seat_offset_dist_y", "seat_offset_dist_z",
		"monetary_value", "ai_handling",
	}

	vehiclesHandlingModelsCols = []string{"vehicle_id", "model"}
)

func (r *SQLite) InsertVehicle(ctx context.Context, v vehicles.Vehicle) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err := insertVehicle(ctx, tx, v); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *SQLite) BulkInsertVehicles(ctx context.Context, v []vehicles.Vehicle) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	for i := range v {
		if err := insertVehicle(ctx, tx, v[i]); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *SQLite) InsertHandling(ctx context.Context, v vehicles.Handling) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err := insertHandling(ctx, tx, v); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *SQLite) BulkInsertHandling(ctx context.Context, v []vehicles.Handling) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	for i := range v {
		if err := insertHandling(ctx, tx, v[i]); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func insertVehicle(ctx context.Context, tx *sql.Tx, v vehicles.Vehicle) error {
	var id int

	if err := tx.QueryRowContext(ctx, insertVehicleQuery,
		v.Name, v.Hash, v.SignedHash, v.HexHash, v.DlcName, v.HandlingId, v.LayoutId, v.Manufacturer, v.Class,
		v.ClassId, v.Type, v.PlateType, v.DashboardType, v.WheelType, v.Seats, v.Price, v.MonetaryValue,
		v.HasConvertibleRoof, v.HasSirens, v.BoundingSphereRadius, v.MaxBraking, v.MaxBrakingMods, v.MaxSpeed,
		v.MaxTraction, v.Acceleration, v.Agility, v.MaxKnots, v.MoveResistance, v.HasArmoredWindows,
		v.DefaultBodyHealth, v.DirtLevelMin, v.DirtLevelMax, v.SpawnFrequency, v.WheelsCount, v.HasParachute,
		v.HasKers, v.DefaultHorn, v.DefaultHornVariation,
	).Scan(&id); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx,
		fmt.Sprintf(insertTableQuery, "vehicles_display_names",
			strings.Join(vehiclesDisplayNamesCols, ","),
			createPlaceholders(len(vehiclesDisplayNamesCols)),
		),
		id, v.DisplayName.Hash, v.DisplayName.English, v.DisplayName.German, v.DisplayName.French, v.DisplayName.Italian,
		v.DisplayName.Russian, v.DisplayName.Polish, v.DisplayName.Name, v.DisplayName.TraditionalChinese,
		v.DisplayName.SimplifiedChinese, v.DisplayName.Spanish, v.DisplayName.Japanese, v.DisplayName.Korean,
		v.DisplayName.Portuguese, v.DisplayName.Mexican,
	); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx,
		fmt.Sprintf(insertTableQuery, "vehicles_manufacturer_display_names",
			strings.Join(vehiclesDisplayNamesCols, ","),
			createPlaceholders(len(vehiclesDisplayNamesCols)),
		),
		id, v.ManufacturerDisplayName.Hash, v.ManufacturerDisplayName.English, v.ManufacturerDisplayName.German,
		v.ManufacturerDisplayName.French, v.ManufacturerDisplayName.Italian, v.ManufacturerDisplayName.Russian,
		v.ManufacturerDisplayName.Polish, v.ManufacturerDisplayName.Name, v.ManufacturerDisplayName.TraditionalChinese,
		v.ManufacturerDisplayName.SimplifiedChinese, v.ManufacturerDisplayName.Spanish, v.ManufacturerDisplayName.Japanese,
		v.ManufacturerDisplayName.Korean, v.ManufacturerDisplayName.Portuguese, v.ManufacturerDisplayName.Mexican,
	); err != nil {
		return err
	}

	for i := range v.Flags {
		if _, err := tx.ExecContext(ctx,
			fmt.Sprintf(insertTableQuery, "vehicles_flags",
				strings.Join(vehiclesFlagsCols, ","),
				createPlaceholders(len(vehiclesFlagsCols)),
			),
			id, v.Flags[i],
		); err != nil {
			return err
		}
	}

	for i := range v.Weapons {
		if _, err := tx.ExecContext(ctx,
			fmt.Sprintf(insertTableQuery, "vehicles_weapons",
				strings.Join(vehiclesWeaponsCols, ","),
				createPlaceholders(len(vehiclesWeaponsCols)),
			),
			id, v.Weapons[i],
		); err != nil {
			return err
		}
	}

	for i := range v.ModKits {
		if _, err := tx.ExecContext(ctx,
			fmt.Sprintf(insertTableQuery, "vehicles_mod_kits",
				strings.Join(vehiclesModKitsCols, ","),
				createPlaceholders(len(vehiclesModKitsCols)),
			),
			id, v.ModKits[i],
		); err != nil {
			return err
		}
	}

	switch {
	case v.DimensionsMin != nil && v.DimensionsMax != nil:
		if _, err := tx.ExecContext(ctx,
			fmt.Sprintf(insertTableQuery, "vehicles_dimensions",
				strings.Join(vehiclesDimensionsCols, ","),
				createPlaceholders(len(vehiclesDimensionsCols)),
			),
			id,
			v.DimensionsMin.X, v.DimensionsMin.Y, v.DimensionsMin.Z, v.DimensionsMax.X, v.DimensionsMax.Y, v.DimensionsMax.Z,
		); err != nil {
			return err
		}
	case v.DimensionsMin == nil && v.DimensionsMax != nil:
		if _, err := tx.ExecContext(ctx,
			fmt.Sprintf(insertTableQuery, "vehicles_dimensions",
				strings.Join(vehiclesDimensionsCols, ","),
				createPlaceholders(len(vehiclesDimensionsCols)),
			),
			id, nil, nil, nil, v.DimensionsMax.X, v.DimensionsMax.Y, v.DimensionsMax.Z,
		); err != nil {
			return err
		}
	case v.DimensionsMin != nil && v.DimensionsMax == nil:
		if _, err := tx.ExecContext(ctx,
			fmt.Sprintf(insertTableQuery, "vehicles_dimensions",
				strings.Join(vehiclesDimensionsCols, ","),
				createPlaceholders(len(vehiclesDimensionsCols)),
			),
			id, v.DimensionsMin.X, v.DimensionsMin.Y, v.DimensionsMin.Z, nil, nil, nil,
		); err != nil {
			return err
		}
	case v.DimensionsMin == nil && v.DimensionsMax == nil:
	}

	for i := range v.Rewards {
		if _, err := tx.ExecContext(ctx,
			fmt.Sprintf(insertTableQuery, "vehicles_rewards",
				strings.Join(vehiclesRewardsCols, ","),
				createPlaceholders(len(vehiclesRewardsCols)),
			),
			id, v.Rewards[i],
		); err != nil {
			return err
		}
	}

	for i := range v.DefaultColors {
		if _, err := tx.ExecContext(ctx,
			fmt.Sprintf(insertTableQuery, "vehicles_default_colors",
				strings.Join(vehiclesDefaultColorsCols, ","),
				createPlaceholders(len(vehiclesDefaultColorsCols)),
			),
			id, v.DefaultColors[i].DefaultPrimaryColor, v.DefaultColors[i].DefaultSecondaryColor,
			v.DefaultColors[i].DefaultPearlColor, v.DefaultColors[i].DefaultWheelsColor,
			v.DefaultColors[i].DefaultInteriorColor, v.DefaultColors[i].DefaultDashboardColor,
		); err != nil {
			return err
		}
	}

	for i := range v.Trailers {
		if _, err := tx.ExecContext(ctx,
			fmt.Sprintf(insertTableQuery, "vehicles_trailers",
				strings.Join(vehiclesTrailersCols, ","),
				createPlaceholders(len(vehiclesTrailersCols)),
			),
			id, v.Trailers[i],
		); err != nil {
			return err
		}
	}

	for i := range v.AdditionalTrailers {
		if _, err := tx.ExecContext(ctx,
			fmt.Sprintf(insertTableQuery, "vehicles_additional_trailers",
				strings.Join(vehiclesAdditionalTrailersCols, ","),
				createPlaceholders(len(vehiclesAdditionalTrailersCols)),
			),
			id, v.AdditionalTrailers[i],
		); err != nil {
			return err
		}
	}

	for i := range v.Extras {
		if _, err := tx.ExecContext(ctx,
			fmt.Sprintf(insertTableQuery, "vehicles_extras",
				strings.Join(vehiclesExtrasCols, ","),
				createPlaceholders(len(vehiclesExtrasCols)),
			),
			id, v.Extras[i],
		); err != nil {
			return err
		}
	}

	for i := range v.RequiredExtras {
		if _, err := tx.ExecContext(ctx,
			fmt.Sprintf(insertTableQuery, "vehicles_required_extras",
				strings.Join(vehiclesRequiredExtrasCols, ","),
				createPlaceholders(len(vehiclesRequiredExtrasCols)),
			),
			id, v.RequiredExtras[i],
		); err != nil {
			return err
		}
	}

	for i := range v.Bones {
		if _, err := tx.ExecContext(ctx,
			fmt.Sprintf(insertTableQuery, "vehicles_bones",
				strings.Join(vehiclesBonesCols, ","),
				createPlaceholders(len(vehiclesBonesCols)),
			),
			id, v.Bones[i].BoneIndex, v.Bones[i].BoneId, v.Bones[i].BoneName,
		); err != nil {
			return err
		}
	}

	return nil
}

func insertHandling(ctx context.Context, tx *sql.Tx, v vehicles.Handling) error {
	if _, err := tx.ExecContext(ctx,
		fmt.Sprintf(insertTableQuery, "vehicles_handling",
			strings.Join(vehiclesHandlingCols, ","),
			createPlaceholders(len(vehiclesHandlingCols)),
		),
		v.Id, v.Mass, v.InitialDragCoeff, v.PercentSubmerged, v.CentreOfMassOffset.X, v.CentreOfMassOffset.Y,
		v.CentreOfMassOffset.Z, v.InertiaMultiplier.X, v.InertiaMultiplier.Y, v.InertiaMultiplier.Z,
		v.DriveBiasFront, v.InitialDriveGears, v.InitialDriveForce, v.DriveInertia, v.ClutchChangeRateScaleUpShift,
		v.ClutchChangeRateScaleDownShift, v.InitialDriveMaxFlatVel, v.BrakeForce, v.BrakeBiasFront, v.HandBrakeForce,
		v.SteeringLock, v.TractionCurveMax, v.TractionCurveMin, v.TractionCurveLateral, v.TractionSpringDeltaMax,
		v.LowSpeedTractionLossMult, v.CamberStiffnesss, v.TractionBiasFront, v.TractionLossMult, v.SuspensionForce,
		v.SuspensionCompDamp, v.SuspensionReboundDamp, v.SuspensionUpperLimit, v.SuspensionLowerLimit, v.SuspensionRaise,
		v.SuspensionBiasFront, v.AntiRollBarForce, v.AntiRollBarBiasFront, v.RollCentreHeightFront, v.RollCentreHeightRear,
		v.CollisionDamageMult, v.WeaponDamageMult, v.DeformationDamageMult, v.EngineDamageMult, v.PetrolTankVolume,
		v.OilVolume, v.SeatOffsetDistX, v.SeatOffsetDistY, v.SeatOffsetDistZ, v.MonetaryValue, v.AiHandling,
	); err != nil {
		return err
	}

	for i := range v.VehicleModels {
		if _, err := tx.ExecContext(ctx,
			fmt.Sprintf(insertTableQuery, "vehicles_handling_models",
				strings.Join(vehiclesHandlingModelsCols, ","),
				createPlaceholders(len(vehiclesHandlingModelsCols)),
			),
			v.Id, v.VehicleModels[i],
		); err != nil {
			return err
		}
	}

	return nil
}

func createPlaceholders(length int) string {
	sb := &strings.Builder{}

	for i := 0; i < length; i++ {
		sb.WriteByte('?')

		if i+1 < length {
			sb.WriteString(", ")
		}
	}

	return sb.String()
}
