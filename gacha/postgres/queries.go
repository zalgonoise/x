package postgres

// The statements below use database time exclusively (now() plus interval parameters), so writers and pollers on
// different hosts never disagree about leases and retries. Each one is a single statement, so it is atomic without an
// explicit transaction.
const (
	writeQuery = `INSERT INTO {table} (id, data) SELECT * FROM unnest($1::uuid[], $2::bytea[])`

	// Data-modifying CTEs always run to completion, so LIMIT 1 only caps the number of notifications.
	writeNotifyQuery = `WITH inserted AS (
	INSERT INTO {table} (id, data) SELECT * FROM unnest($1::uuid[], $2::bytea[]) RETURNING 1
)
SELECT pg_notify($3, '') FROM inserted LIMIT 1`

	claimQuery = `WITH claimable AS (
	SELECT id, available_at FROM {table}
	WHERE status IN ('pending', 'processing') AND available_at <= now()
	ORDER BY available_at, id
	LIMIT $1
	FOR UPDATE SKIP LOCKED
), claimed AS (
	UPDATE {table} AS t
	SET status = 'processing', attempts = t.attempts + 1, updated_at = now(), available_at = now() + $2::interval
	FROM claimable AS c
	WHERE t.id = c.id
	RETURNING t.id, t.data, t.attempts, t.created_at, c.available_at AS claimed_from
)
SELECT id, data, attempts, created_at FROM claimed ORDER BY claimed_from, id`

	// Every outcome is fenced on status = 'processing' and the claim's attempt count. Completed rows reset
	// available_at, so the Completed view's lease starts at completion.
	settleQuery = `WITH completed AS (
	UPDATE {table} AS t
	SET status = 'completed', updated_at = now(), available_at = now(),
		last_error = COALESCE(NULLIF(i.err, ''), t.last_error)
	FROM unnest($1::uuid[], $2::int4[], $3::text[]) AS i(id, attempts, err)
	WHERE t.id = i.id AND t.attempts = i.attempts AND t.status = 'processing'
), ` + settleRetryFailCTEs

	settleDeleteQuery = `WITH completed AS (
	DELETE FROM {table} AS t
	USING unnest($1::uuid[], $2::int4[], $3::text[]) AS i(id, attempts, err)
	WHERE t.id = i.id AND t.attempts = i.attempts AND t.status = 'processing'
), ` + settleRetryFailCTEs

	// Released entries go back to pending without counting the claim as an attempt.
	settleRetryFailCTEs = `retried AS (
	UPDATE {table} AS t
	SET status = 'pending', updated_at = now(), available_at = now() + i.retry_in,
		last_error = COALESCE(NULLIF(i.err, ''), t.last_error)
	FROM unnest($4::uuid[], $5::int4[], $6::text[], $7::interval[]) AS i(id, attempts, err, retry_in)
	WHERE t.id = i.id AND t.attempts = i.attempts AND t.status = 'processing'
), released AS (
	UPDATE {table} AS t
	SET status = 'pending', attempts = t.attempts - 1, updated_at = now(), available_at = now(),
		last_error = COALESCE(NULLIF(i.err, ''), t.last_error)
	FROM unnest($11::uuid[], $12::int4[], $13::text[]) AS i(id, attempts, err)
	WHERE t.id = i.id AND t.attempts = i.attempts AND t.status = 'processing'
), failed AS (
	DELETE FROM {table} AS t
	USING unnest($8::uuid[], $9::int4[], $10::text[]) AS i(id, attempts, err)
	WHERE t.id = i.id AND t.attempts = i.attempts AND t.status = 'processing'
	RETURNING t.id, t.data, t.attempts, COALESCE(NULLIF(i.err, ''), t.last_error) AS last_error, t.created_at
)
INSERT INTO {dlq} (id, data, attempts, last_error, created_at)
SELECT id, data, attempts, last_error, created_at FROM failed`

	extendQuery = `UPDATE {table} SET available_at = now() + $1::interval, updated_at = now()
WHERE id = ANY($2::uuid[]) AND status = 'processing'`

	redriveQuery = `WITH moved AS (
	DELETE FROM {dlq} WHERE id = ANY($1::uuid[])
	RETURNING id, data, last_error, created_at
)
INSERT INTO {table} (id, data, status, attempts, last_error, created_at, updated_at, available_at)
SELECT id, data, 'pending', 0, last_error, created_at, now(), now() FROM moved`

	redriveNotifyQuery = `WITH moved AS (
	DELETE FROM {dlq} WHERE id = ANY($1::uuid[])
	RETURNING id, data, last_error, created_at
), inserted AS (
	INSERT INTO {table} (id, data, status, attempts, last_error, created_at, updated_at, available_at)
	SELECT id, data, 'pending', 0, last_error, created_at, now(), now() FROM moved
	RETURNING 1
)
SELECT pg_notify($2, '') FROM inserted LIMIT 1`

	statsQuery = `SELECT
	count(*) FILTER (WHERE status = 'pending'),
	count(*) FILTER (WHERE status = 'processing'),
	count(*) FILTER (WHERE status = 'completed'),
	(SELECT count(*) FROM {dlq})
FROM {table}`

	// Views lease rows through available_at, without touching their attempt counts or update times.
	completedClaimQuery = `WITH claimable AS (
	SELECT id, updated_at FROM {table}
	WHERE status = 'completed' AND updated_at <= now() - $3::interval AND available_at <= now()
	ORDER BY updated_at, id
	LIMIT $1
	FOR UPDATE SKIP LOCKED
), claimed AS (
	UPDATE {table} AS t
	SET available_at = now() + $2::interval
	FROM claimable AS c
	WHERE t.id = c.id
	RETURNING t.id, t.data, t.attempts, t.created_at, c.updated_at AS since
)
SELECT id, data, attempts, created_at FROM claimed ORDER BY since, id`

	completedSettleQuery = `WITH deleted AS (
	DELETE FROM {table} WHERE id = ANY($1::uuid[]) AND status = 'completed'
)
UPDATE {table} AS t
SET available_at = now() + i.retry_in
FROM unnest($2::uuid[], $3::interval[]) AS i(id, retry_in)
WHERE t.id = i.id AND t.status = 'completed'`

	deadLettersClaimQuery = `WITH claimable AS (
	SELECT id, failed_at FROM {dlq}
	WHERE failed_at <= now() - $3::interval AND available_at <= now()
	ORDER BY failed_at, id
	LIMIT $1
	FOR UPDATE SKIP LOCKED
), claimed AS (
	UPDATE {dlq} AS t
	SET available_at = now() + $2::interval
	FROM claimable AS c
	WHERE t.id = c.id
	RETURNING t.id, t.data, t.attempts, t.created_at, c.failed_at AS since
)
SELECT id, data, attempts, created_at FROM claimed ORDER BY since, id`

	deadLettersSettleQuery = `WITH deleted AS (
	DELETE FROM {dlq} WHERE id = ANY($1::uuid[])
)
UPDATE {dlq} AS t
SET available_at = now() + i.retry_in
FROM unnest($2::uuid[], $3::interval[]) AS i(id, retry_in)
WHERE t.id = i.id`
)

// queries holds the statements of a Repository, with its table names filled in.
type queries struct {
	write, claim, settle, extend, redrive, stats string
	completedClaim, completedSettle              string
	deadLettersClaim, deadLettersSettle          string
}

func newQueries(out, dlq table, config Config) queries {
	r := replacer(out, dlq)

	q := queries{
		write:             r.Replace(writeQuery),
		claim:             r.Replace(claimQuery),
		settle:            r.Replace(settleQuery),
		extend:            r.Replace(extendQuery),
		redrive:           r.Replace(redriveQuery),
		stats:             r.Replace(statsQuery),
		completedClaim:    r.Replace(completedClaimQuery),
		completedSettle:   r.Replace(completedSettleQuery),
		deadLettersClaim:  r.Replace(deadLettersClaimQuery),
		deadLettersSettle: r.Replace(deadLettersSettleQuery),
	}

	if config.channel != "" {
		q.write = r.Replace(writeNotifyQuery)
		q.redrive = r.Replace(redriveNotifyQuery)
	}

	if config.deleteOnComplete {
		q.settle = r.Replace(settleDeleteQuery)
	}

	return q
}
