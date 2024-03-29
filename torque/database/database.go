package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/XSAM/otelsql"
	"github.com/zalgonoise/cfg"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	_ "modernc.org/sqlite" // Database driver
)

const (
	uriFormat = "file:%s?cache=shared"
	inMemory  = ":memory:"

	defaultMaxOpenConns = 16
	defaultMaxIdleConns = 16

	checkTableExists = `
	SELECT EXISTS(SELECT 1 FROM sqlite_master
	WHERE type='table'
	AND name='%s');
`

	createVehiclesTable = `
CREATE TABLE vehicles 
(
    id         								INTEGER PRIMARY KEY NOT NULL,
    name 											TEXT 								NOT NULL,
    hash 											INTEGER 						NOT NULL,
    signed_hash 							INTEGER 						NOT NULL,
    hex_hash 									TEXT 								NOT NULL,
    dlc_name 									TEXT 								NOT NULL,
    handling_id 							TEXT 								NOT NULL,
    layout_id 								TEXT 								NOT NULL,
    manufacturer 							TEXT 								NOT NULL,
    class 										TEXT 								NOT NULL,
    class_id 									INTEGER 						NOT NULL,
    type 											TEXT 								NOT NULL,
    plate_type 								TEXT 								NOT NULL,
    dashboard_type 						TEXT,
    wheel_type 								TEXT,
    seats 										INTEGER 						NOT NULL,
    price 										INTEGER 						NOT NULL,
    monetary_value 						INTEGER 						NOT NULL,
    has_convertible_roof 			BOOLEAN 						NOT NULL,
    has_sirens 								BOOLEAN 						NOT NULL,
    bounding_sphere_radius 		FLOAT 							NOT NULL,
    max_braking 							FLOAT 							NOT NULL,
    max_braking_mods 					FLOAT 							NOT NULL,
    max_speed 								FLOAT 							NOT NULL,
    max_traction 							FLOAT 							NOT NULL,
    acceleration 							FLOAT 							NOT NULL,
    agility 									FLOAT 							NOT NULL,
    max_knots 								FLOAT 							NOT NULL,
    move_resistance 					FLOAT 							NOT NULL,
    has_armored_windows 			BOOLEAN 						NOT NULL,
    default_body_health 			FLOAT 							NOT NULL,
    dirt_level_min 						FLOAT 							NOT NULL,
    dirt_level_max 						FLOAT 							NOT NULL,
    spawn_frequency 					FLOAT 							NOT NULL,
    wheels_count 							INTEGER 						NOT NULL,
	  has_parachute 						BOOLEAN 						NOT NULL,
	  has_kers 									BOOLEAN 						NOT NULL,
	  default_horn 							INTEGER 						NOT NULL,
	  default_horn_variation 		INTEGER 						NOT NULL
);
`
	createDisplayNamesTable = `
CREATE TABLE vehicles_display_names
(
    vehicle_id    				INTEGER REFERENCES vehicles (id) NOT NULL,
    hash 									INTEGER													 NOT NULL,
    english 							TEXT,
    german 								TEXT,
    french 								TEXT,
    italian 							TEXT,
    russian 							TEXT,
    polish 								TEXT,
    name 									TEXT,
    traditional_chinese 	TEXT,
    simplified_chinese 		TEXT,
    spanish 							TEXT,
    japanese 							TEXT,
    korean 								TEXT,
    portuguese 						TEXT,
    mexican 							TEXT
);

CREATE INDEX idx_vehicles_display_names_vehicle_id ON vehicles_display_names (vehicle_id);
`

	createManufacturerDisplayNameTable = `
CREATE TABLE vehicles_manufacturer_display_names
(
    vehicle_id    				INTEGER REFERENCES vehicles (id) NOT NULL,
    hash 									INTEGER													 NOT NULL,
    english 							TEXT,
    german 								TEXT,
    french 								TEXT,
    italian 							TEXT,
    russian 							TEXT,
    polish 								TEXT,
    name 									TEXT,
    traditional_chinese 	TEXT,
    simplified_chinese 		TEXT,
    spanish 							TEXT,
    japanese 							TEXT,
    korean 								TEXT,
    portuguese 						TEXT,
    mexican 							TEXT
);

CREATE INDEX idx_vehicles_manufacturer_display_names_vehicle_id ON vehicles_manufacturer_display_names (vehicle_id);
`

	createFlagsTable = `
CREATE TABLE vehicles_flags
(
    vehicle_id    				INTEGER REFERENCES vehicles (id) NOT NULL,
    flag 									TEXT													 	 NOT NULL
);

CREATE INDEX idx_vehicles_flags_vehicle_id ON vehicles_flags (vehicle_id);
`

	createWeaponsTable = `
CREATE TABLE vehicles_weapons
(
    vehicle_id    				INTEGER REFERENCES vehicles (id) NOT NULL,
    weapon 								TEXT														 NOT NULL
);

CREATE INDEX idx_vehicles_weapons_vehicle_id ON vehicles_weapons (vehicle_id);
`

	createModKitsTable = `
CREATE TABLE vehicles_mod_kits
(
    vehicle_id    				INTEGER REFERENCES vehicles (id) NOT NULL,
    mod_kit 							TEXT														 NOT NULL
);

CREATE INDEX idx_vehicles_mod_kits_vehicle_id ON vehicles_mod_kits (vehicle_id);
`

	createDimensionsTable = `
CREATE TABLE vehicles_dimensions
(
    vehicle_id    				INTEGER REFERENCES vehicles (id) NOT NULL,
    min_x 								FLOAT														 NOT NULL,
    min_y 								FLOAT														 NOT NULL,
    min_z 								FLOAT														 NOT NULL,
    max_x 								FLOAT														 NOT NULL,
    max_y 								FLOAT														 NOT NULL,
    max_z 								FLOAT														 NOT NULL
);

CREATE INDEX idx_vehicles_dimensions_vehicle_id ON vehicles_dimensions (vehicle_id);
`

	createRewardsTable = `
CREATE TABLE vehicles_rewards
(
    vehicle_id    				INTEGER REFERENCES vehicles (id) NOT NULL,
    rewards 							TEXT														 NOT NULL
);

CREATE INDEX idx_vehicles_rewards_vehicle_id ON vehicles_rewards (vehicle_id);
`

	createDefaultColorsTable = `
CREATE TABLE vehicles_default_colors
(
    vehicle_id    					INTEGER REFERENCES vehicles (id) NOT NULL,
    default_primary_color 	INTEGER													 NOT NULL,
    default_secondary_color INTEGER													 NOT NULL,
    default_pearl_color 		INTEGER													 NOT NULL,
		default_wheels_color 		INTEGER													 NOT NULL,
		default_interior_color 	INTEGER													 NOT NULL,
		default_dashboard_color INTEGER													 NOT NULL
);

CREATE INDEX idx_vehicles_default_colors_vehicle_id ON vehicles_default_colors (vehicle_id);
`

	createTrailersTable = `
CREATE TABLE vehicles_trailers
(
    vehicle_id    				INTEGER REFERENCES vehicles (id) NOT NULL,
    trailers 							TEXT														 NOT NULL
);

CREATE INDEX idx_vehicles_trailers_vehicle_id ON vehicles_trailers (vehicle_id);
`

	createAdditionalTrailersTable = `
CREATE TABLE vehicles_additional_trailers
(
    vehicle_id    				INTEGER REFERENCES vehicles (id) NOT NULL,
    additional_trailers 	TEXT														 NOT NULL
);

CREATE INDEX idx_vehicles_additional_trailers_vehicle_id ON vehicles_additional_trailers (vehicle_id);
`

	createExtrasTable = `
CREATE TABLE vehicles_extras
(
    vehicle_id    	INTEGER REFERENCES vehicles (id) NOT NULL,
    extras				 	INTEGER													 NOT NULL
);

CREATE INDEX idx_vehicles_extras_vehicle_id ON vehicles_extras (vehicle_id);
`

	createRequiredExtrasTable = `
CREATE TABLE vehicles_required_extras
(
    vehicle_id    		INTEGER REFERENCES vehicles (id) NOT NULL,
    required_extras		INTEGER													 NOT NULL
);

CREATE INDEX idx_vehicles_required_extras_vehicle_id ON vehicles_required_extras (vehicle_id);
`

	createBonesTable = `
CREATE TABLE vehicles_bones
(
    vehicle_id    INTEGER REFERENCES vehicles (id) NOT NULL,
    bone_index		INTEGER													 NOT NULL,
    bone_id				INTEGER													 NOT NULL,
    bone_name			TEXT													 	NOT NULL
);

CREATE INDEX idx_vehicles_bones_vehicle_id ON vehicles_bones (vehicle_id);
`

	createHandlingTable = `
CREATE TABLE vehicles_handling
(
    id    															TEXT		REFERENCES vehicles (handling_id) NOT NULL,
    mass																FLOAT													 						NOT NULL,
    initial_drag_coeff									FLOAT													 						NOT NULL,
    percent_submerged										FLOAT													 						NOT NULL,
    centre_of_mass_offset_x							FLOAT													 						NOT NULL,
    centre_of_mass_offset_y							FLOAT													 						NOT NULL,
    centre_of_mass_offset_z							FLOAT													 						NOT NULL,
    inertia_multiplier_x								FLOAT													 						NOT NULL,
    inertia_multiplier_y								FLOAT													 						NOT NULL,
    inertia_multiplier_z								FLOAT													 						NOT NULL,
    drive_bias_front										FLOAT													 						NOT NULL,
    initial_drive_gears									INTEGER													 					NOT NULL,
    initial_drive_force									FLOAT													 						NOT NULL,
    drive_inertia												FLOAT													 						NOT NULL,
    clutch_change_rate_scale_up_shift		FLOAT													 						NOT NULL,
    clutch_change_rate_scale_down_shift	FLOAT													 						NOT NULL,
    initial_drive_max_flat_vel					FLOAT													 						NOT NULL,
    brake_force													FLOAT													 						NOT NULL,
    brake_bias_front										FLOAT													 						NOT NULL,
    handbrake_force											FLOAT													 						NOT NULL,
    steering_lock												FLOAT													 						NOT NULL,
    traction_curve_max									FLOAT													 						NOT NULL,
    traction_curve_min									FLOAT													 						NOT NULL,
    traction_curve_lateral							FLOAT													 						NOT NULL,
    traction_spring_delta_max						FLOAT													 						NOT NULL,
    low_speed_traction_loss_mult				FLOAT													 						NOT NULL,
    camber_stiffness										FLOAT													 						NOT NULL,
    traction_bias_front									FLOAT													 						NOT NULL,
    traction_loss_mult									FLOAT													 						NOT NULL,
    suspension_force										FLOAT													 						NOT NULL,
    suspension_comp_damp								FLOAT													 						NOT NULL,
    suspension_rebound_damp							FLOAT													 						NOT NULL,
    suspension_upper_limit							FLOAT													 						NOT NULL,
    suspension_lower_limit							FLOAT													 						NOT NULL,
    suspension_raise										FLOAT													 						NOT NULL,
    suspension_bias_front								FLOAT													 						NOT NULL,
    anti_roll_bar_force									FLOAT													 						NOT NULL,
    anti_roll_bar_bias_front						FLOAT													 						NOT NULL,
    roll_centre_height_front						FLOAT													 						NOT NULL,
    roll_centre_height_rear							FLOAT													 						NOT NULL,
    collision_damage_mult								FLOAT													 						NOT NULL,
    weapon_damage_mult									FLOAT													 						NOT NULL,
    deformation_damage_mult							FLOAT													 						NOT NULL,
    engine_damage_mult									FLOAT													 						NOT NULL,
    petrol_tank_volume									FLOAT													 						NOT NULL,
    oil_volume													FLOAT													 						NOT NULL,
    seat_offset_dist_x									FLOAT													 						NOT NULL,
    seat_offset_dist_y									FLOAT													 						NOT NULL,
    seat_offset_dist_z									FLOAT													 						NOT NULL,
    monetary_value											INTEGER													 					NOT NULL,
    ai_handling													TEXT													 						NOT NULL
);

CREATE INDEX idx_vehicles_handling_handling_id ON vehicles_handling (id);
`

	createHandlingModelsTable = `
CREATE TABLE vehicles_handling_models
(
    vehicle_id    				TEXT		REFERENCES vehicles_handling (id) NOT NULL,
    model 								TEXT														 					NOT NULL
);

CREATE INDEX idx_vehicles_handling_models_vehicle_id ON vehicles_handling_models (vehicle_id);
`
)

func Open(uri string, opts ...cfg.Option[Config]) (*sql.DB, error) {
	switch uri {
	case inMemory:
	case "":
		uri = inMemory
	default:
		if err := validateURI(uri); err != nil {
			return nil, err
		}
	}

	config := cfg.New(opts...)

	db, err := otelsql.Open("sqlite",
		fmt.Sprintf(uriFormat, uri),
		otelsql.WithAttributes(semconv.DBSystemSqlite),
		otelsql.WithSpanOptions(otelsql.SpanOptions{
			OmitConnResetSession: true,
		}),
	)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(config.maxOpenConns)
	db.SetMaxIdleConns(config.maxIdleConns)

	return db, nil
}

func validateURI(uri string) error {
	stat, err := os.Stat(uri)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			f, err := os.Create(uri)
			if err != nil {
				return err
			}

			return f.Close()
		}

		return err
	}

	if stat.IsDir() {
		return fmt.Errorf("%s is a directory", uri)
	}

	return nil
}

func Migrate(ctx context.Context, db *sql.DB, log *slog.Logger) error {
	return runMigrations(ctx, db,
		migration{"vehicles", createVehiclesTable},
		migration{"vehicles_display_names", createDisplayNamesTable},
		migration{"vehicles_manufacturer_display_names", createManufacturerDisplayNameTable},
		migration{"vehicles_flags", createFlagsTable},
		migration{"vehicles_weapons", createWeaponsTable},
		migration{"vehicles_mod_kits", createModKitsTable},
		migration{"vehicles_dimensions", createDimensionsTable},
		migration{"vehicles_rewards", createRewardsTable},
		migration{"vehicles_default_colors", createDefaultColorsTable},
		migration{"vehicles_trailers", createTrailersTable},
		migration{"vehicles_additional_trailers", createAdditionalTrailersTable},
		migration{"vehicles_extras", createExtrasTable},
		migration{"vehicles_required_extras", createRequiredExtrasTable},
		migration{"vehicles_bones", createBonesTable},
		migration{"vehicles_handling", createHandlingTable},
		migration{"vehicles_handling_models", createHandlingModelsTable},
	)
}

type migration struct {
	table  string
	create string
}

func runMigrations(ctx context.Context, db *sql.DB, migrations ...migration) error {
	for i := range migrations {
		r, err := db.QueryContext(ctx, fmt.Sprintf(checkTableExists, migrations[i].table))
		if err != nil {
			return err
		}

		var count int

		if !r.Next() {
			return r.Err()
		}

		if err = r.Scan(&count); err != nil {
			_ = r.Close()

			return err
		}

		_ = r.Close()

		if count == 1 {
			continue
		}

		_, err = db.ExecContext(ctx, migrations[i].create)
		if err != nil {
			return err
		}
	}

	return nil
}
