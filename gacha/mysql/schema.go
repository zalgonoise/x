package mysql

import "strings"

// The outbox emulates partial indexes with stored generated columns: claim_at is the claim time of pending and
// processing rows (NULL for completed ones) and completed_at the completion time of completed rows (NULL otherwise).
// A range on either column skips the NULLs, so claims and Completed view claims scan their index in order and stop
// after LIMIT rows. A (status, available_at, id) index cannot do that, since status IN (...) splits it in two ranges
// that MySQL has to sort.
const (
	outboxTemplate = `CREATE TABLE IF NOT EXISTS {table} (
    id           BINARY(16)  NOT NULL,
    data         LONGBLOB    NOT NULL,
    status       ENUM('pending', 'processing', 'completed') NOT NULL DEFAULT 'pending',
    attempts     INT         NOT NULL DEFAULT 0,
    last_error   MEDIUMTEXT  NULL,
    created_at   DATETIME(6) NOT NULL,
    updated_at   DATETIME(6) NOT NULL,
    available_at DATETIME(6) NOT NULL,
    claim_at     DATETIME(6) AS (IF(status = 'completed', NULL, available_at)) STORED,
    completed_at DATETIME(6) AS (IF(status = 'completed', updated_at, NULL)) STORED,
    PRIMARY KEY (id),
    INDEX claim_idx (claim_at, id),
    INDEX completed_idx (completed_at, id)
)`

	deadLettersTemplate = `CREATE TABLE IF NOT EXISTS {dlq} (
    id           BINARY(16)  NOT NULL,
    data         LONGBLOB    NOT NULL,
    attempts     INT         NOT NULL,
    last_error   MEDIUMTEXT  NULL,
    created_at   DATETIME(6) NOT NULL,
    failed_at    DATETIME(6) NOT NULL,
    available_at DATETIME(6) NOT NULL,
    PRIMARY KEY (id),
    INDEX failed_idx (failed_at, id)
)`
)

// maxTableParts is the number of parts of a database-qualified table name.
const maxTableParts = 2

// parseTable returns the quoted identifier of a table name, which may be qualified with a database name.
func parseTable(name string) (string, error) {
	parts := strings.Split(name, ".")
	if len(parts) > maxTableParts {
		return "", ErrInvalidTable
	}

	for i, part := range parts {
		if part == "" {
			return "", ErrInvalidTable
		}

		parts[i] = "`" + strings.ReplaceAll(part, "`", "``") + "`"
	}

	return strings.Join(parts, "."), nil
}

// tables returns the quoted outbox and dead-letter tables. An empty dlqTable defaults to the outbox's name with a
// "_dlq" suffix.
func tables(outbox, dlqTable string) (out, dlq string, err error) {
	if dlqTable == "" {
		dlqTable = outbox + deadLetterSuffix
	}

	if out, err = parseTable(outbox); err != nil {
		return "", "", err
	}

	if dlq, err = parseTable(dlqTable); err != nil {
		return "", "", err
	}

	return out, dlq, nil
}

func schemaStatements(out, dlq string) []string {
	r := strings.NewReplacer("{table}", out, "{dlq}", dlq)

	return []string{r.Replace(outboxTemplate), r.Replace(deadLettersTemplate)}
}

// Schema returns the DDL that creates the outbox table and its dead-letter table, with their indexes, for use with a
// migration tool. Table names may be qualified with a database name ("db.table"). An empty dlqTable defaults to the
// outbox's name with a "_dlq" suffix. The statements are idempotent and separated by semicolons.
func Schema(outbox, dlqTable string) (string, error) {
	out, dlq, err := tables(outbox, dlqTable)
	if err != nil {
		return "", err
	}

	return strings.Join(schemaStatements(out, dlq), ";\n\n") + ";\n", nil
}
