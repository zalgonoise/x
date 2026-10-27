package postgres

import (
	"strings"

	"github.com/jackc/pgx/v5"
)

// migrateLock serializes concurrent migrations: PostgreSQL runs a multi-statement simple query as one transaction,
// so the advisory lock is held until the schema is in place.
const migrateLock = "SELECT pg_advisory_xact_lock(hashtext('github.com/zalgonoise/x/gacha'));\n"

const schemaTemplate = `DO $$ BEGIN
    CREATE TYPE gacha_status AS ENUM ('pending', 'processing', 'completed');
EXCEPTION WHEN duplicate_object OR unique_violation THEN NULL; END $$;

CREATE TABLE IF NOT EXISTS {table} (
    id           UUID         PRIMARY KEY,
    data         BYTEA        NOT NULL,
    status       gacha_status NOT NULL DEFAULT 'pending',
    attempts     INTEGER      NOT NULL DEFAULT 0,
    last_error   TEXT,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    available_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS {claim_idx}
    ON {table} (available_at, id) WHERE status IN ('pending', 'processing');
CREATE INDEX IF NOT EXISTS {completed_idx}
    ON {table} (updated_at, id) WHERE status = 'completed';

CREATE TABLE IF NOT EXISTS {dlq} (
    id           UUID         PRIMARY KEY,
    data         BYTEA        NOT NULL,
    attempts     INTEGER      NOT NULL,
    last_error   TEXT,
    created_at   TIMESTAMPTZ  NOT NULL,
    failed_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    available_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS {failed_idx}
    ON {dlq} (failed_at, id);
`

// maxTableParts is the number of parts of a schema-qualified table name.
const maxTableParts = 2

// table is a validated, possibly schema-qualified, table name.
type table struct {
	// ident is the quoted identifier, such as "public"."orders_outbox".
	ident string
	// name is the unqualified table name, used to name its indexes.
	name string
}

func parseTable(name string) (table, error) {
	parts := strings.Split(name, ".")
	if len(parts) > maxTableParts {
		return table{}, ErrInvalidTable
	}

	for _, part := range parts {
		if part == "" {
			return table{}, ErrInvalidTable
		}
	}

	return table{
		ident: pgx.Identifier(parts).Sanitize(),
		name:  parts[len(parts)-1],
	}, nil
}

// index returns the quoted name of the table's index with the given suffix. Indexes live in their table's schema.
func (t table) index(suffix string) string {
	return pgx.Identifier{t.name + "_" + suffix}.Sanitize()
}

// tables returns the parsed outbox and dead-letter tables. An empty dlqTable defaults to the outbox's name with a
// "_dlq" suffix.
func tables(outbox, dlqTable string) (table, table, error) {
	if dlqTable == "" {
		dlqTable = outbox + deadLetterSuffix
	}

	out, err := parseTable(outbox)
	if err != nil {
		return table{}, table{}, err
	}

	dlq, err := parseTable(dlqTable)
	if err != nil {
		return table{}, table{}, err
	}

	return out, dlq, nil
}

// replacer fills in the table and index names of the SQL templates.
func replacer(out, dlq table) *strings.Replacer {
	return strings.NewReplacer(
		"{table}", out.ident,
		"{dlq}", dlq.ident,
		"{claim_idx}", out.index("claim_idx"),
		"{completed_idx}", out.index("completed_idx"),
		"{failed_idx}", dlq.index("failed_idx"),
	)
}

// Schema returns the DDL that creates the outbox table, its dead-letter table and their indexes, for use with a
// migration tool. Table names may be schema-qualified ("schema.table"). An empty dlqTable defaults to the outbox's
// name with a "_dlq" suffix. The statements are idempotent. They also create the gacha_status enum type, shared by
// every outbox table, in the first schema of the search path.
func Schema(outbox, dlqTable string) (string, error) {
	out, dlq, err := tables(outbox, dlqTable)
	if err != nil {
		return "", err
	}

	return replacer(out, dlq).Replace(schemaTemplate), nil
}
