package sqlite

import "strings"

// Times are INTEGER microseconds since the Unix epoch, in UTC, which sort and compare as plain integers.
const schemaTemplate = `CREATE TABLE IF NOT EXISTS {table} (
    id           BLOB    NOT NULL PRIMARY KEY,
    data         BLOB    NOT NULL,
    status       TEXT    NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed')),
    attempts     INTEGER NOT NULL DEFAULT 0,
    last_error   TEXT,
    created_at   INTEGER NOT NULL,
    updated_at   INTEGER NOT NULL,
    available_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS {claim_idx}
    ON {table_name} (available_at, id) WHERE status IN ('pending', 'processing');

CREATE INDEX IF NOT EXISTS {completed_idx}
    ON {table_name} (updated_at, id) WHERE status = 'completed';

CREATE TABLE IF NOT EXISTS {dlq} (
    id           BLOB    NOT NULL PRIMARY KEY,
    data         BLOB    NOT NULL,
    attempts     INTEGER NOT NULL,
    last_error   TEXT,
    created_at   INTEGER NOT NULL,
    failed_at    INTEGER NOT NULL,
    available_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS {failed_idx}
    ON {dlq_name} (failed_at, id);
`

// maxTableParts is the number of parts of a schema-qualified table name.
const maxTableParts = 2

// table is a validated, possibly schema-qualified, table name.
type table struct {
	// ident is the quoted identifier, such as "aux"."orders_outbox".
	ident string
	// schema is the quoted schema prefix with its dot, such as "aux"., or empty.
	schema string
	// name is the unqualified table name.
	name string
}

func quote(ident string) string {
	return `"` + strings.ReplaceAll(ident, `"`, `""`) + `"`
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

	t := table{name: parts[len(parts)-1]}
	if len(parts) == maxTableParts {
		t.schema = quote(parts[0]) + "."
	}

	t.ident = t.schema + quote(t.name)

	return t, nil
}

// index returns the quoted name of the table's index with the given suffix. In SQLite, the schema qualifies the
// index name (indexes share their table's schema), and index names are unique per schema.
func (t table) index(suffix string) string {
	return t.schema + quote(t.name+"_"+suffix)
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

func schemaStatements(out, dlq table) []string {
	r := strings.NewReplacer(
		"{table}", out.ident,
		"{table_name}", quote(out.name),
		"{dlq}", dlq.ident,
		"{dlq_name}", quote(dlq.name),
		"{claim_idx}", out.index("claim_idx"),
		"{completed_idx}", out.index("completed_idx"),
		"{failed_idx}", dlq.index("failed_idx"),
	)

	statements := strings.Split(strings.TrimSpace(schemaTemplate), ";\n")

	for i := range statements {
		statements[i] = r.Replace(strings.TrimSuffix(strings.TrimSpace(statements[i]), ";"))
	}

	return statements
}

// Schema returns the DDL that creates the outbox table, its dead-letter table and their indexes, for use with a
// migration tool. Table names may be qualified with the name of an attached database ("aux.table"). An empty
// dlqTable defaults to the outbox's name with a "_dlq" suffix. The statements are idempotent and separated by
// semicolons.
func Schema(outbox, dlqTable string) (string, error) {
	out, dlq, err := tables(outbox, dlqTable)
	if err != nil {
		return "", err
	}

	return strings.Join(schemaStatements(out, dlq), ";\n\n") + ";\n", nil
}
