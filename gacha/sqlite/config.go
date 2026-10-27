package sqlite

import "github.com/zalgonoise/cfg"

const deadLetterSuffix = "_dlq"

// Config holds the configuration of a SQLite Repository.
type Config struct {
	dlqTable         string
	deleteOnComplete bool
}

func defaultConfig() Config {
	return Config{}
}

// WithDeadLetterTable sets the name of the dead-letter table, which defaults to the outbox table's name with a
// "_dlq" suffix. Like the outbox table, it may be qualified with the name of an attached database. In WAL mode, keep
// both tables in the same database: SQLite only commits transactions spanning several database files atomically in
// rollback-journal mode, so a crash could leave an entry moved to the dead-letter table in one file but not deleted
// from the outbox in the other.
func WithDeadLetterTable(table string) cfg.Option[Config] {
	if table == "" {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register(func(c Config) Config {
		c.dlqTable = table

		return c
	})
}

// WithDeleteOnComplete makes Settle delete completed entries instead of keeping them for the janitor. Use it for
// streams that need no audit trail: it saves the janitor's purge per row. The Completed view is then always empty.
func WithDeleteOnComplete() cfg.Option[Config] {
	return cfg.Register(func(c Config) Config {
		c.deleteOnComplete = true

		return c
	})
}
