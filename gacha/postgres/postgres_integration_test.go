//go:build integration

package postgres_test

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/zalgonoise/cfg"

	"github.com/zalgonoise/x/gacha"
	"github.com/zalgonoise/x/gacha/gachatest"
	"github.com/zalgonoise/x/gacha/poller"
	"github.com/zalgonoise/x/gacha/postgres"
	"github.com/zalgonoise/x/gacha/writer"
)

const image = "postgres:18-alpine"

var (
	pool     *pgxpool.Pool
	tableSeq atomic.Int64
)

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx, image,
		tcpostgres.WithDatabase("gacha"),
		tcpostgres.WithUsername("gacha"),
		tcpostgres.WithPassword("gacha"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "start postgres:", err)

		return 1
	}

	defer func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			fmt.Fprintln(os.Stderr, "terminate postgres:", err)
		}
	}()

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintln(os.Stderr, "postgres connection string:", err)

		return 1
	}

	pool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Fprintln(os.Stderr, "connect to postgres:", err)

		return 1
	}

	defer pool.Close()

	return m.Run()
}

// newTable returns a unique table name, whose tables are dropped when the test ends.
func newTable(t testing.TB) string {
	t.Helper()

	table := "outbox_" + strconv.FormatInt(tableSeq.Add(1), 10)

	t.Cleanup(func() {
		_, err := pool.Exec(context.Background(), "DROP TABLE IF EXISTS "+table+", "+table+"_dlq")
		require.NoError(t, err)
	})

	return table
}

func newRepo(t testing.TB, opts ...cfg.Option[postgres.Config]) *postgres.Repository {
	t.Helper()

	repo, _ := newRepoTable(t, opts...)

	return repo
}

// newRepoTable returns a migrated repository over a new table, along with the table's name.
func newRepoTable(t testing.TB, opts ...cfg.Option[postgres.Config]) (*postgres.Repository, string) {
	t.Helper()

	table := newTable(t)

	repo, err := postgres.New(pool, table, opts...)
	require.NoError(t, err)
	require.NoError(t, repo.Migrate(context.Background()))

	return repo, table
}

func TestRepository(t *testing.T) {
	gachatest.Suite[[]byte]{
		New: func(t *testing.T) gacha.Repository[[]byte] {
			return newRepo(t)
		},
		NewDeleteOnComplete: func(t *testing.T) gacha.Repository[[]byte] {
			return newRepo(t, postgres.WithDeleteOnComplete())
		},
		Value: gachatest.Bytes,
		Tx: func(t *testing.T, repo gacha.Repository[[]byte]) (gacha.Sink[[]byte], func(bool) error) {
			tx, err := pool.Begin(t.Context())
			require.NoError(t, err)

			t.Cleanup(func() {
				_ = tx.Rollback(context.Background())
			})

			return repo.(*postgres.Repository).Tx(tx), func(commit bool) error {
				if commit {
					return tx.Commit(t.Context())
				}

				return tx.Rollback(t.Context())
			}
		},
	}.Run(t)
}

func TestMigrate_Concurrent(t *testing.T) {
	repo, err := postgres.New(pool, newTable(t))
	require.NoError(t, err)

	var wg sync.WaitGroup

	errs := make(chan error, 8)

	for range 8 {
		wg.Go(func() {
			errs <- repo.Migrate(t.Context())
		})
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		require.NoError(t, err)
	}

	require.NoError(t, repo.Write(t.Context(), []byte("ok")))
}

func TestSchemaQualified(t *testing.T) {
	_, err := pool.Exec(t.Context(), "CREATE SCHEMA IF NOT EXISTS gacha_it")
	require.NoError(t, err)

	t.Cleanup(func() {
		_, err := pool.Exec(context.Background(), "DROP SCHEMA gacha_it CASCADE")
		require.NoError(t, err)
	})

	repo, err := postgres.New(pool, "gacha_it.orders", postgres.WithDeadLetterTable("gacha_it.dead_orders"))
	require.NoError(t, err)
	require.NoError(t, repo.Migrate(t.Context()))
	require.NoError(t, repo.Write(t.Context(), []byte("a")))

	entries, err := repo.Claim(t.Context(), 1, time.Minute)
	require.NoError(t, err)
	require.Len(t, entries, 1)

	require.NoError(t, repo.Settle(t.Context(), gacha.Outcome{
		ID: entries[0].ID, Attempts: entries[0].Attempts, Status: gacha.StatusFailed, Err: "boom",
	}))

	var (
		data      []byte
		attempts  int
		lastError string
	)

	err = pool.QueryRow(t.Context(), "SELECT data, attempts, last_error FROM gacha_it.dead_orders WHERE id = $1",
		entries[0].ID).Scan(&data, &attempts, &lastError)
	require.NoError(t, err)
	require.Equal(t, []byte("a"), data)
	require.Equal(t, 1, attempts)
	require.Equal(t, "boom", lastError)
}

func TestLastError(t *testing.T) {
	repo, table := newRepoTable(t)
	require.NoError(t, repo.Write(t.Context(), []byte("a")))

	lastError := func(id uuid.UUID) *string {
		var err *string

		require.NoError(t, pool.QueryRow(t.Context(),
			"SELECT last_error FROM "+table+" WHERE id = $1", id).Scan(&err))

		return err
	}

	entries, err := repo.Claim(t.Context(), 1, time.Minute)
	require.NoError(t, err)
	require.Nil(t, lastError(entries[0].ID))

	require.NoError(t, repo.Settle(t.Context(), gacha.Outcome{
		ID: entries[0].ID, Attempts: 1, Status: gacha.StatusPending, Err: "first",
	}))
	require.Equal(t, "first", *lastError(entries[0].ID))

	entries, err = repo.Claim(t.Context(), 1, time.Minute)
	require.NoError(t, err)

	// An outcome without an error keeps the previous one.
	require.NoError(t, repo.Settle(t.Context(), gacha.Outcome{ID: entries[0].ID, Attempts: 2, Status: gacha.StatusCompleted}))
	require.Equal(t, "first", *lastError(entries[0].ID))
}

func TestNilData(t *testing.T) {
	repo := newRepo(t)
	require.NoError(t, repo.Write(t.Context(), nil, []byte{}, []byte("x")))

	entries, err := repo.Claim(t.Context(), 10, time.Minute)
	require.NoError(t, err)
	require.Len(t, entries, 3)
	require.Empty(t, entries[0].Data)
	require.Empty(t, entries[1].Data)
	require.Equal(t, []byte("x"), entries[2].Data)
}

// startListener runs a Listener on channel until the test ends, and waits until it listens.
func startListener(t *testing.T, channel string, opts ...cfg.Option[postgres.ListenerConfig]) *postgres.Listener {
	t.Helper()

	listener, err := postgres.NewListener(pool, channel, opts...)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)

	go func() {
		done <- listener.Run(ctx)
	}()

	t.Cleanup(func() {
		cancel()
		require.NoError(t, <-done)
	})

	// The Listener fires once it listens.
	requireFires(t, listener)

	return listener
}

func requireFires(t *testing.T, trigger poller.Trigger) {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	require.NoError(t, trigger.Wait(ctx, poller.Result{}))
}

func requireQuiet(t *testing.T, trigger poller.Trigger) {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), 300*time.Millisecond)
	defer cancel()

	require.ErrorIs(t, trigger.Wait(ctx, poller.Result{}), context.DeadlineExceeded)
}

func TestListener(t *testing.T) {
	const channel = "gacha_listener"

	repo := newRepo(t, postgres.WithNotify(channel))
	listener := startListener(t, channel)

	requireQuiet(t, listener)

	require.NoError(t, repo.Write(t.Context(), []byte("a")))
	requireFires(t, listener)

	// Notifications sent in a transaction are delivered on commit.
	tx, err := pool.Begin(t.Context())
	require.NoError(t, err)
	require.NoError(t, repo.Tx(tx).Write(t.Context(), []byte("b")))
	requireQuiet(t, listener)
	require.NoError(t, tx.Commit(t.Context()))
	requireFires(t, listener)

	// Redrives notify too.
	entries, err := repo.Claim(t.Context(), 1, time.Minute)
	require.NoError(t, err)
	require.NoError(t, repo.Settle(t.Context(), gacha.Outcome{
		ID: entries[0].ID, Attempts: 1, Status: gacha.StatusFailed,
	}))
	requireQuiet(t, listener)
	require.NoError(t, repo.Redrive(t.Context(), entries[0].ID))
	requireFires(t, listener)
}

func TestListener_Reconnects(t *testing.T) {
	const channel = "gacha_reconnect"

	repo := newRepo(t, postgres.WithNotify(channel))
	listener := startListener(t, channel, postgres.WithReconnect(poller.Constant(10*time.Millisecond)))

	var terminated int

	err := pool.QueryRow(t.Context(), `SELECT count(pg_terminate_backend(pid)) FROM pg_stat_activity
		WHERE query = $1`, "LISTEN "+pgx.Identifier{channel}.Sanitize()).Scan(&terminated)
	require.NoError(t, err)
	require.Equal(t, 1, terminated)

	// The Listener fires when it is back, to catch up on what it may have missed.
	requireFires(t, listener)

	require.NoError(t, repo.Write(t.Context(), []byte("a")))
	requireFires(t, listener)
}

func TestClaim_UsesIndex(t *testing.T) {
	repo, table := newRepoTable(t)

	data := make([][]byte, 1000)
	for i := range data {
		data[i] = gachatest.Bytes(i)
	}

	require.NoError(t, repo.Write(t.Context(), data...))

	_, err := pool.Exec(t.Context(), "ANALYZE "+table)
	require.NoError(t, err)

	tx, err := pool.Begin(t.Context())
	require.NoError(t, err)

	defer func() {
		_ = tx.Rollback(context.Background())
	}()

	_, err = tx.Exec(t.Context(), "SET LOCAL enable_seqscan = off")
	require.NoError(t, err)

	rows, err := tx.Query(t.Context(), "EXPLAIN "+postgres.ClaimQuery(repo), 100, time.Minute)
	require.NoError(t, err)

	lines, err := pgx.CollectRows(rows, pgx.RowTo[string])
	require.NoError(t, err)

	var plan string
	for _, line := range lines {
		plan += line + "\n"
	}

	t.Log(plan)
	require.Contains(t, plan, table+"_claim_idx", plan)
}

func TestPollers(t *testing.T) {
	const (
		items   = 500
		pollers = 3
	)

	type order struct {
		ID int `json:"id"`
	}

	repo := newRepo(t)
	w := writer.NewBatch(repo, gacha.JSON[order](), writer.WithLinger(5*time.Millisecond))

	var (
		mu      sync.Mutex
		handled = make(map[int]int, items)
		total   atomic.Int64
	)

	handler := func(_ context.Context, o order) error {
		mu.Lock()
		handled[o.ID]++
		mu.Unlock()

		total.Add(1)

		return nil
	}

	ctx, cancel := context.WithCancel(t.Context())

	var running sync.WaitGroup

	for range pollers {
		p := poller.New(repo, gacha.JSON[order](), handler,
			poller.WithBatchSize(25),
			poller.WithConcurrency(4),
			poller.WithTrigger(poller.Backoff(5*time.Millisecond, 50*time.Millisecond)),
		)

		running.Go(func() {
			if err := p.Run(ctx); err != nil {
				t.Error(err)
			}
		})
	}

	var writers sync.WaitGroup

	for id := range items {
		writers.Go(func() {
			if err := w.Write(t.Context(), order{ID: id}); err != nil {
				t.Error(err)
			}
		})
	}

	writers.Wait()
	require.NoError(t, w.Close(t.Context()))

	require.Eventually(t, func() bool {
		return total.Load() >= items
	}, 30*time.Second, 10*time.Millisecond)

	cancel()
	running.Wait()

	require.Len(t, handled, items)

	for id, n := range handled {
		require.Equal(t, 1, n, "order %d handled %d times", id, n)
	}

	stats, err := repo.Stats(t.Context())
	require.NoError(t, err)
	require.Equal(t, gacha.Stats{Completed: items}, stats)
}

func TestShortLease(t *testing.T) {
	repo, table := newRepoTable(t)
	require.NoError(t, repo.Write(t.Context(), []byte("a")))

	_, err := repo.Claim(t.Context(), 1, time.Nanosecond)
	require.NoError(t, err)

	var leased bool

	require.NoError(t, pool.QueryRow(t.Context(), "SELECT available_at > updated_at FROM "+table).Scan(&leased))
	require.True(t, leased, "a sub-microsecond lease is rounded up, not truncated to zero")
}
