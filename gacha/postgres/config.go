package postgres

import "github.com/zalgonoise/cfg"

const deadLetterSuffix = "_dlq"

// Config holds the configuration of a PostgreSQL Repository.
type Config struct {
	dlqTable         string
	channel          string
	deleteOnComplete bool
}

func defaultConfig() Config {
	return Config{}
}

// WithDeadLetterTable sets the name of the dead-letter table, which defaults to the outbox table's name with a
// "_dlq" suffix. Like the outbox table, it may be schema-qualified.
func WithDeadLetterTable(table string) cfg.Option[Config] {
	if table == "" {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register(func(c Config) Config {
		c.dlqTable = table

		return c
	})
}

// WithNotify makes every write (and redrive) send a notification on channel with pg_notify, in the same statement.
// Inside a transaction, PostgreSQL delivers it on commit. A Listener on the same channel turns the notifications into
// a poller.Trigger.
func WithNotify(channel string) cfg.Option[Config] {
	if channel == "" {
		return cfg.NoOp[Config]{}
	}

	return cfg.Register(func(c Config) Config {
		c.channel = channel

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
