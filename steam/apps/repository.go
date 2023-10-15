package apps

import (
	"context"
	"database/sql"

	"github.com/zalgonoise/x/cfg"
	"github.com/zalgonoise/x/steam"
	_ "modernc.org/sqlite"
)

const (
	minAlloc = 64

	checkTableExists = `
SELECT EXISTS(SELECT 1 FROM sqlite_master 
	WHERE type='table' 
	AND name='steam_apps');
`

	createTableQuery = `
CREATE VIRTUAL TABLE steam_apps 
	USING FTS5(
		app_id, 
		app_name
	);
`

	insertValueQuery = `
INSERT INTO steam_apps (app_id, app_name) 
	VALUES (?, ?);
`

	searchQuery = `
SELECT app_id, app_name 
	FROM steam_apps 
	WHERE app_name MATCH ?;
`
)

type Repository struct {
	config Config

	db *sql.DB
}

func NewRepository(opts ...cfg.Option[Config]) (*Repository, error) {
	config := cfg.New(opts...)

	if config.uri == "" {
		config.uri = inMemory
	}

	db, err := open(config)
	if err != nil {
		return nil, err
	}

	if err = initDatabase(db); err != nil {
		return nil, err
	}

	return &Repository{
		config: config,
		db:     db}, nil
}

func (r *Repository) Search(ctx context.Context, name string) ([]steam.App, error) {
	rows, err := r.db.QueryContext(ctx, searchQuery, name)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	res := make([]steam.App, 0, minAlloc)

	for rows.Next() {
		var (
			appID   int64
			appName string
		)

		if err = rows.Scan(&appID, &appName); err != nil {
			return nil, err
		}

		res = append(res, steam.App{
			AppID: appID,
			Name:  appName,
		})
	}

	return res, nil
}
