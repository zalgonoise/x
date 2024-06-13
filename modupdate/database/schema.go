package database

const (
	createRepositories = `CREATE TABLE repositories (
    id          		TEXT PRIMARY KEY    NOT NULL,
    uri         		TEXT                NOT NULL,
    module      		TEXT,
    branch      		TEXT,
    username    		TEXT,
    token       		TEXT,
    cron_schedule 	TEXT,
    dry_run         INTEGER          		NOT NULL,
    fs_path         TEXT             		NOT NULL,
    commit_message  TEXT             		NOT NULL
) STRICT;`

	createOverrides = `CREATE TABLE overrides (
    id          TEXT PRIMARY KEY    NOT NULL,
    type        TEXT                NOT NULL,
    command     TEXT                NOT NULL
) STRICT;`

	createBin = `CREATE TABLE bin (
    git TEXT    NOT NULL,
    go  TEXT    NOT NULL
) STRICT ;`
)

func newSchema() []migration {
	return []migration{
		{table: "repositories", create: createRepositories},
		{table: "overrides", create: createOverrides},
		{table: "bin", create: createBin},
	}
}
