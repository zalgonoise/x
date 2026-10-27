package sqlite_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/zalgonoise/cfg"
	_ "modernc.org/sqlite"

	"github.com/zalgonoise/x/gacha"
	"github.com/zalgonoise/x/gacha/gachatest"
	"github.com/zalgonoise/x/gacha/poller"
	"github.com/zalgonoise/x/gacha/sqlite"
	"github.com/zalgonoise/x/gacha/writer"
)

// Compile-time checks that Repository implements every gacha capability.
var (
	_ gacha.Repository[[]byte] = (*sqlite.Repository)(nil)
	_ gacha.Extender           = (*sqlite.Repository)(nil)
	_ gacha.Redriver           = (*sqlite.Repository)(nil)
	_ gacha.Stater             = (*sqlite.Repository)(nil)
	_ gacha.Viewer[[]byte]     = (*sqlite.Repository)(nil)
)

const table = "orders_outbox"

// openFile opens a new database file in WAL mode, with a busy timeout so that concurrent writers wait for the lock.
func openFile(t testing.TB) *sql.DB {
	t.Helper()

	path := filepath.Join(t.TempDir(), "gacha.db")

	db, err := sql.Open("sqlite", "file:"+path+"?_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)")
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	return db
}

// openMemory opens a new in-memory database on a single connection, so that every operation sees it.
func openMemory(t testing.TB) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)

	db.SetMaxOpenConns(1)

	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	return db
}

func newRepo(t testing.TB, db *sql.DB, opts ...cfg.Option[sqlite.Config]) *sqlite.Repository {
	t.Helper()

	repo, err := sqlite.New(db, table, opts...)
	require.NoError(t, err)
	require.NoError(t, repo.Migrate(context.Background()))

	return repo
}

func TestRepository(t *testing.T) {
	for _, testcase := range []struct {
		name string
		open func(t testing.TB) *sql.DB
	}{
		{name: "File", open: openFile},
		{name: "Memory", open: openMemory},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			// The Tx tests need the database behind each repository.
			var dbs sync.Map

			newRepoDB := func(t *testing.T, opts ...cfg.Option[sqlite.Config]) gacha.Repository[[]byte] {
				db := testcase.open(t)
				repo := newRepo(t, db, opts...)
				dbs.Store(repo, db)

				return repo
			}

			gachatest.Suite[[]byte]{
				New: func(t *testing.T) gacha.Repository[[]byte] {
					return newRepoDB(t)
				},
				NewDeleteOnComplete: func(t *testing.T) gacha.Repository[[]byte] {
					return newRepoDB(t, sqlite.WithDeleteOnComplete())
				},
				Value: gachatest.Bytes,
				Tx: func(t *testing.T, repo gacha.Repository[[]byte]) (gacha.Sink[[]byte], func(bool) error) {
					db, ok := dbs.Load(repo)
					require.True(t, ok)

					tx, err := db.(*sql.DB).BeginTx(t.Context(), nil)
					require.NoError(t, err)

					t.Cleanup(func() {
						_ = tx.Rollback()
					})

					return repo.(*sqlite.Repository).Tx(tx), func(commit bool) error {
						if commit {
							return tx.Commit()
						}

						return tx.Rollback()
					}
				},
			}.Run(t)
		})
	}
}

func TestMigrate_Concurrent(t *testing.T) {
	db := openFile(t)

	repo, err := sqlite.New(db, table)
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

func TestAttachedDatabase(t *testing.T) {
	// ATTACH applies to one connection: keep the pool to a single one.
	db := openMemory(t)

	_, err := db.ExecContext(t.Context(), "ATTACH DATABASE ? AS aux", filepath.Join(t.TempDir(), "aux.db"))
	require.NoError(t, err)

	repo, err := sqlite.New(db, "aux.orders", sqlite.WithDeadLetterTable("aux.dead_orders"))
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

	err = db.QueryRowContext(t.Context(), "SELECT data, attempts, last_error FROM aux.dead_orders WHERE id = ?",
		entries[0].ID[:]).Scan(&data, &attempts, &lastError)
	require.NoError(t, err)
	require.Equal(t, []byte("a"), data)
	require.Equal(t, 1, attempts)
	require.Equal(t, "boom", lastError)

	var indexes int

	err = db.QueryRowContext(t.Context(), "SELECT count(*) FROM aux.sqlite_master WHERE type = 'index' AND name LIKE ?",
		"%_idx").Scan(&indexes)
	require.NoError(t, err)
	require.Equal(t, 3, indexes, "indexes live in the tables' database")
}

func TestNilData(t *testing.T) {
	repo := newRepo(t, openFile(t))
	require.NoError(t, repo.Write(t.Context(), nil, []byte{}, []byte("x")))

	entries, err := repo.Claim(t.Context(), 10, time.Minute)
	require.NoError(t, err)
	require.Len(t, entries, 3)
	require.Equal(t, []byte{}, entries[0].Data)
	require.Equal(t, []byte{}, entries[1].Data)
	require.Equal(t, []byte("x"), entries[2].Data)
}

func TestLastError(t *testing.T) {
	db := openFile(t)
	repo := newRepo(t, db)
	require.NoError(t, repo.Write(t.Context(), []byte("a")))

	lastError := func(id []byte) sql.NullString {
		var err sql.NullString

		require.NoError(t, db.QueryRowContext(t.Context(),
			"SELECT last_error FROM "+table+" WHERE id = ?", id).Scan(&err))

		return err
	}

	entries, err := repo.Claim(t.Context(), 1, time.Minute)
	require.NoError(t, err)
	require.False(t, lastError(entries[0].ID[:]).Valid)

	long := strings.Repeat("ガチャ", 30000)

	require.NoError(t, repo.Settle(t.Context(), gacha.Outcome{
		ID: entries[0].ID, Attempts: 1, Status: gacha.StatusPending, Err: long,
	}))
	require.Equal(t, long, lastError(entries[0].ID[:]).String)

	entries, err = repo.Claim(t.Context(), 1, time.Minute)
	require.NoError(t, err)

	// An outcome without an error keeps the previous one.
	require.NoError(t, repo.Settle(t.Context(), gacha.Outcome{
		ID: entries[0].ID, Attempts: 2, Status: gacha.StatusCompleted,
	}))
	require.Equal(t, long, lastError(entries[0].ID[:]).String)
}

func TestLargeWrite(t *testing.T) {
	repo := newRepo(t, openFile(t))

	data := make([][]byte, 2500)
	for i := range data {
		data[i] = gachatest.Bytes(i)
	}

	require.NoError(t, repo.Write(t.Context(), data...))

	var claimed int

	for {
		entries, err := repo.Claim(t.Context(), 1500, time.Minute)
		require.NoError(t, err)

		if len(entries) == 0 {
			break
		}

		for i, entry := range entries {
			require.Equal(t, data[claimed+i], entry.Data, "entry %d", claimed+i)
		}

		claimed += len(entries)
	}

	require.Equal(t, len(data), claimed)
}

// plan returns the details of EXPLAIN QUERY PLAN query.
func plan(t *testing.T, db *sql.DB, query string, args ...any) string {
	t.Helper()

	rows, err := db.QueryContext(t.Context(), "EXPLAIN QUERY PLAN "+query, args...)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, rows.Close())
	}()

	var details []string

	for rows.Next() {
		var (
			id, parent, unused int
			detail             string
		)

		require.NoError(t, rows.Scan(&id, &parent, &unused, &detail))

		details = append(details, detail)
	}

	require.NoError(t, rows.Err())

	return strings.Join(details, "\n")
}

func TestQueryPlans(t *testing.T) {
	db := openFile(t)
	repo := newRepo(t, db)

	data := make([][]byte, 2000)
	for i := range data {
		data[i] = gachatest.Bytes(i)
	}

	require.NoError(t, repo.Write(t.Context(), data...))

	entries, err := repo.Claim(t.Context(), 1000, time.Minute)
	require.NoError(t, err)

	outcomes := make([]gacha.Outcome, len(entries))
	for i := range entries {
		outcomes[i] = gacha.Outcome{ID: entries[i].ID, Attempts: 1, Status: gacha.StatusCompleted}
	}

	require.NoError(t, repo.Settle(t.Context(), outcomes...))

	_, err = db.ExecContext(t.Context(), "ANALYZE")
	require.NoError(t, err)

	for _, testcase := range []struct {
		name       string
		query      string
		args       []any
		wantsIndex string
	}{
		{name: "Claim", query: sqlite.ClaimQuery(repo), args: []any{100}, wantsIndex: table + "_claim_idx"},
		{
			name:       "CompletedView",
			query:      sqlite.ViewClaimQuery(repo.Completed(0)),
			args:       []any{0, 100},
			wantsIndex: table + "_completed_idx",
		},
		{
			name:       "DeadLettersView",
			query:      sqlite.ViewClaimQuery(repo.DeadLetters(0)),
			args:       []any{0, 100},
			wantsIndex: table + "_dlq_failed_idx",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			details := plan(t, db, testcase.query, testcase.args...)
			t.Log(details)

			require.Contains(t, details, "USING INDEX "+testcase.wantsIndex)
			require.NotContains(t, details, "TEMP B-TREE", "the index provides the order")
		})
	}

	t.Run("Settle", func(t *testing.T) {
		query, args := sqlite.SettleCompletedQuery(repo, outcomes[:10])
		details := plan(t, db, query, args...)
		t.Log(details)

		require.Contains(t, details, "SEARCH t USING INDEX sqlite_autoindex_"+table+"_1 (id=?)")
		require.NotContains(t, details, "SCAN t")
	})
}

func TestPollers(t *testing.T) {
	const (
		items   = 300
		pollers = 3
	)

	type order struct {
		ID int `json:"id"`
	}

	repo := newRepo(t, openFile(t))
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

func TestTxWrite_Atomic(t *testing.T) {
	db := openFile(t)
	repo := newRepo(t, db)

	_, err := db.ExecContext(t.Context(), "CREATE TRIGGER no_boom BEFORE INSERT ON "+table+
		" WHEN NEW.data = CAST('boom' AS BLOB) BEGIN SELECT RAISE(ABORT, 'boom'); END")
	require.NoError(t, err)

	// More rows than one statement holds, with a failing row in the second statement.
	data := make([][]byte, 1500)
	for i := range data {
		data[i] = gachatest.Bytes(i)
	}

	data[1200] = []byte("boom")

	tx, err := db.BeginTx(t.Context(), nil)
	require.NoError(t, err)

	sink := repo.Tx(tx)
	require.Error(t, sink.Write(t.Context(), data...))

	// The failed write left nothing behind, and the transaction is still usable.
	require.NoError(t, sink.Write(t.Context(), []byte("after")))
	require.NoError(t, tx.Commit())

	entries, err := repo.Claim(t.Context(), 2000, time.Minute)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Equal(t, []byte("after"), entries[0].Data)
}

func TestShortLease(t *testing.T) {
	db := openFile(t)
	repo := newRepo(t, db)
	require.NoError(t, repo.Write(t.Context(), []byte("a")))

	_, err := repo.Claim(t.Context(), 1, time.Nanosecond)
	require.NoError(t, err)

	var leased bool

	require.NoError(t, db.QueryRowContext(t.Context(),
		"SELECT available_at > updated_at FROM "+table).Scan(&leased))
	require.True(t, leased, "a sub-microsecond lease is rounded up, not truncated to zero")
}
