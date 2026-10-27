package mysql

import "github.com/zalgonoise/cfg"

const deadLetterSuffix = "_dlq"

// Config holds the configuration of a MySQL Repository.
type Config struct {
	dlqTable         string
	deleteOnComplete bool
}

func defaultConfig() Config {
	return Config{}
}

// WithDeadLetterTable sets the name of the dead-letter table, which defaults to the outbox table's name with a
// "_dlq" suffix. Like the outbox table, it may be qualified with a database name.
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
