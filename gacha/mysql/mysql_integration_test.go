//go:build integration

package mysql_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcmariadb "github.com/testcontainers/testcontainers-go/modules/mariadb"
	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/zalgonoise/cfg"

	"github.com/zalgonoise/x/gacha"
	"github.com/zalgonoise/x/gacha/gachatest"
	"github.com/zalgonoise/x/gacha/mysql"
	"github.com/zalgonoise/x/gacha/poller"
	"github.com/zalgonoise/x/gacha/writer"
)

const (
	mysqlImage   = "mysql:8.4"
	mariadbImage = "mariadb:11.8.6"
)

// backend is a database server, with a pool that interpolates parameters (the recommended DSN) and one that uses
// server-side prepared statements (the driver's default).
type backend struct {
	name     string
	db       *sql.DB
	prepared *sql.DB
	dsn      func(ctx context.Context, args ...string) (string, error)
}

var (
	backends []*backend
	tableSeq atomic.Int64
)

type container interface {
	testcontainers.Container
	ConnectionString(ctx context.Context, args ...string) (string, error)
}

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	ctx := context.Background()

	starters := []struct {
		name  string
		start func() (container, error)
	}{
		{name: "MySQL", start: func() (container, error) {
			return tcmysql.Run(ctx, mysqlImage,
				tcmysql.WithDatabase("gacha"), tcmysql.WithUsername("gacha"), tcmysql.WithPassword("gacha"))
		}},
		{name: "MariaDB", start: func() (container, error) {
			return tcmariadb.Run(ctx, mariadbImage,
				tcmariadb.WithDatabase("gacha"), tcmariadb.WithUsername("gacha"), tcmariadb.WithPassword("gacha"))
		}},
	}

	containers := make([]container, len(starters))
	errs := make([]error, len(starters))

	var wg sync.WaitGroup

	for i, starter := range starters {
		wg.Go(func() {
			containers[i], errs[i] = starter.start()
		})
	}

	wg.Wait()

	defer func() {
		for _, c := range containers {
			if c != nil {
				if err := testcontainers.TerminateContainer(c); err != nil {
					fmt.Fprintln(os.Stderr, "terminate container:", err)
				}
			}
		}
	}()

	for i, starter := range starters {
		if errs[i] != nil {
			fmt.Fprintln(os.Stderr, "start", starter.name+":", errs[i])

			return 1
		}

		b, err := connect(ctx, starter.name, containers[i])
		if err != nil {
			fmt.Fprintln(os.Stderr, "connect to", starter.name+":", err)

			return 1
		}

		defer func() {
			if err := errors.Join(b.db.Close(), b.prepared.Close()); err != nil {
				fmt.Fprintln(os.Stderr, "close", b.name+":", err)
			}
		}()

		backends = append(backends, b)
	}

	return m.Run()
}

func connect(ctx context.Context, name string, c container) (*backend, error) {
	b := &backend{name: name, dsn: c.ConnectionString}

	for _, conn := range []struct {
		db   **sql.DB
		args []string
	}{
		{db: &b.db, args: []string{"interpolateParams=true", "transaction_isolation=%27READ-COMMITTED%27"}},
		{db: &b.prepared},
	} {
		dsn, err := c.ConnectionString(ctx, conn.args...)
		if err != nil {
			return nil, err
		}

		if *conn.db, err = sql.Open("mysql", dsn); err != nil {
			return nil, err
		}

		if err = (*conn.db).PingContext(ctx); err != nil {
			return nil, err
		}
	}

	return b, nil
}

// forEach runs test as a subtest for each backend.
func forEach(t *testing.T, test func(t *testing.T, b *backend)) {
	t.Helper()

	for _, b := range backends {
		t.Run(b.name, func(t *testing.T) {
			test(t, b)
		})
	}
}

// newTable returns a unique table name, whose tables are dropped when the test ends.
func newTable(t testing.TB, db *sql.DB) string {
	t.Helper()

	table := "outbox_" + strconv.FormatInt(tableSeq.Add(1), 10)

	t.Cleanup(func() {
		_, err := db.ExecContext(context.Background(), "DROP TABLE IF EXISTS "+table+", "+table+"_dlq")
		require.NoError(t, err)
	})

	return table
}

func newRepo(t testing.TB, db *sql.DB, opts ...cfg.Option[mysql.Config]) (*mysql.Repository, string) {
	t.Helper()

	table := newTable(t, db)

	repo, err := mysql.New(db, table, opts...)
	require.NoError(t, err)
	require.NoError(t, repo.Migrate(context.Background()))

	return repo, table
}

func TestRepository(t *testing.T) {
	forEach(t, func(t *testing.T, b *backend) {
		for _, mode := range []struct {
			name string
			db   *sql.DB
		}{
			{name: "Interpolated", db: b.db},
			{name: "Prepared", db: b.prepared},
		} {
			t.Run(mode.name, func(t *testing.T) {
				gachatest.Suite[[]byte]{
					New: func(t *testing.T) gacha.Repository[[]byte] {
						repo, _ := newRepo(t, mode.db)

						return repo
					},
					NewDeleteOnComplete: func(t *testing.T) gacha.Repository[[]byte] {
						repo, _ := newRepo(t, mode.db, mysql.WithDeleteOnComplete())

						return repo
					},
					Value: gachatest.Bytes,
					Tx: func(t *testing.T, repo gacha.Repository[[]byte]) (gacha.Sink[[]byte], func(bool) error) {
						tx, err := mode.db.BeginTx(t.Context(), nil)
						require.NoError(t, err)

						t.Cleanup(func() {
							_ = tx.Rollback()
						})

						return repo.(*mysql.Repository).Tx(tx), func(commit bool) error {
							if commit {
								return tx.Commit()
							}

							return tx.Rollback()
						}
					},
				}.Run(t)
			})
		}
	})
}

func TestMigrate_Concurrent(t *testing.T) {
	forEach(t, func(t *testing.T, b *backend) {
		repo, err := mysql.New(b.db, newTable(t, b.db))
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
	})
}

func TestQualifiedTables(t *testing.T) {
	forEach(t, func(t *testing.T, b *backend) {
		table := newTable(t, b.db)
		dlq := table + "_dead"

		t.Cleanup(func() {
			_, err := b.db.ExecContext(context.Background(), "DROP TABLE IF EXISTS "+dlq)
			require.NoError(t, err)
		})

		repo, err := mysql.New(b.db, "gacha."+table, mysql.WithDeadLetterTable("gacha."+dlq))
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

		err = b.db.QueryRowContext(t.Context(), "SELECT data, attempts, last_error FROM gacha."+dlq+" WHERE id = ?",
			entries[0].ID[:]).Scan(&data, &attempts, &lastError)
		require.NoError(t, err)
		require.Equal(t, []byte("a"), data)
		require.Equal(t, 1, attempts)
		require.Equal(t, "boom", lastError)
	})
}

func TestLastError(t *testing.T) {
	forEach(t, func(t *testing.T, b *backend) {
		repo, table := newRepo(t, b.db)
		require.NoError(t, repo.Write(t.Context(), []byte("a")))

		lastError := func(id []byte) sql.NullString {
			var err sql.NullString

			require.NoError(t, b.db.QueryRowContext(t.Context(),
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
		require.Equal(t, long, lastError(entries[0].ID[:]).String, "long, non-ASCII errors are kept whole")

		entries, err = repo.Claim(t.Context(), 1, time.Minute)
		require.NoError(t, err)

		// An outcome without an error keeps the previous one.
		require.NoError(t, repo.Settle(t.Context(), gacha.Outcome{
			ID: entries[0].ID, Attempts: 2, Status: gacha.StatusCompleted,
		}))
		require.Equal(t, long, lastError(entries[0].ID[:]).String)
	})
}

func TestLargeWrite(t *testing.T) {
	forEach(t, func(t *testing.T, b *backend) {
		repo, _ := newRepo(t, b.db, mysql.WithDeleteOnComplete())

		// More rows than one statement holds, and more bytes than one statement carries.
		const rows, blobs = 2500, 6

		data := make([][]byte, rows, rows+blobs)
		for i := range data {
			data[i] = gachatest.Bytes(i)
		}

		blob := make([]byte, 1<<20)
		for range blobs {
			data = append(data, blob)
		}

		require.NoError(t, repo.Write(t.Context(), data...))

		stats, err := repo.Stats(t.Context())
		require.NoError(t, err)
		require.Equal(t, int64(len(data)), stats.Pending)

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
	})
}

func TestNilData(t *testing.T) {
	forEach(t, func(t *testing.T, b *backend) {
		repo, _ := newRepo(t, b.db)
		require.NoError(t, repo.Write(t.Context(), nil, []byte{}, []byte("x")))

		entries, err := repo.Claim(t.Context(), 10, time.Minute)
		require.NoError(t, err)
		require.Len(t, entries, 3)
		require.Empty(t, entries[0].Data)
		require.Empty(t, entries[1].Data)
		require.Equal(t, []byte("x"), entries[2].Data)
	})
}

// explain returns the rows of EXPLAIN query as maps of column names to values.
func explain(t *testing.T, db *sql.DB, query string, args ...any) []map[string]string {
	t.Helper()

	rows, err := db.QueryContext(t.Context(), "EXPLAIN "+query, args...)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, rows.Close())
	}()

	columns, err := rows.Columns()
	require.NoError(t, err)

	var plan []map[string]string

	for rows.Next() {
		values := make([]sql.NullString, len(columns))
		dest := make([]any, len(columns))

		for i := range values {
			dest[i] = &values[i]
		}

		require.NoError(t, rows.Scan(dest...))

		row := make(map[string]string, len(columns))
		for i, column := range columns {
			row[column] = values[i].String
		}

		plan = append(plan, row)
	}

	require.NoError(t, rows.Err())

	return plan
}

func TestQueryPlans(t *testing.T) {
	forEach(t, func(t *testing.T, b *backend) {
		repo, table := newRepo(t, b.db)

		data := make([][]byte, 2000)
		for i := range data {
			data[i] = gachatest.Bytes(i)
		}

		require.NoError(t, repo.Write(t.Context(), data...))

		// Complete half of the entries, so claims have completed rows to skip.
		entries, err := repo.Claim(t.Context(), 1000, time.Minute)
		require.NoError(t, err)

		outcomes := make([]gacha.Outcome, len(entries))
		for i := range entries {
			outcomes[i] = gacha.Outcome{ID: entries[i].ID, Attempts: 1, Status: gacha.StatusCompleted}
		}

		require.NoError(t, repo.Settle(t.Context(), outcomes...))

		_, err = b.db.ExecContext(t.Context(), "ANALYZE TABLE "+table)
		require.NoError(t, err)

		t.Run("Claim", func(t *testing.T) {
			plan := explain(t, b.db, mysql.ClaimQuery(repo), 100)
			t.Log(plan)
			require.Len(t, plan, 1)
			require.Equal(t, "claim_idx", plan[0]["key"], "%v", plan)
			require.NotContains(t, plan[0]["Extra"], "filesort", "%v", plan)
		})

		t.Run("Settle", func(t *testing.T) {
			query, args := mysql.SettleCompletedQuery(repo, outcomes[:10])
			plan := explain(t, b.db, query, args...)
			t.Log(plan)

			for _, row := range plan {
				if row["table"] == "t" {
					require.Equal(t, "PRIMARY", row["key"], "%v", plan)
				}
			}
		})
	})
}

func TestPollers(t *testing.T) {
	forEach(t, func(t *testing.T, b *backend) {
		const (
			items   = 300
			pollers = 3
		)

		type order struct {
			ID int `json:"id"`
		}

		repo, _ := newRepo(t, b.db)
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
	})
}

func TestTxWrite_Atomic(t *testing.T) {
	forEach(t, func(t *testing.T, b *backend) {
		repo, table := newRepo(t, b.db)

		_, err := b.db.ExecContext(t.Context(), "ALTER TABLE "+table+" ADD CONSTRAINT "+table+
			"_no_boom CHECK (data <> 'boom')")
		require.NoError(t, err)

		// More rows than one statement holds, with a failing row in the second statement.
		data := make([][]byte, 1500)
		for i := range data {
			data[i] = gachatest.Bytes(i)
		}

		data[1200] = []byte("boom")

		tx, err := b.db.BeginTx(t.Context(), nil)
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
	})
}

func TestShortLease(t *testing.T) {
	forEach(t, func(t *testing.T, b *backend) {
		repo, table := newRepo(t, b.db)
		require.NoError(t, repo.Write(t.Context(), []byte("a")))

		_, err := repo.Claim(t.Context(), 1, time.Nanosecond)
		require.NoError(t, err)

		var leased bool

		require.NoError(t, b.db.QueryRowContext(t.Context(),
			"SELECT available_at > updated_at FROM "+table).Scan(&leased))
		require.True(t, leased, "a sub-microsecond lease is rounded up, not truncated to zero")
	})
}

func TestCreatedAt_ParseTimeLocation(t *testing.T) {
	forEach(t, func(t *testing.T, b *backend) {
		// Times are stored as UTC: a DSN parsing them in another location must not shift them.
		dsn, err := b.dsn(t.Context(), "interpolateParams=true", "parseTime=true", "loc=Asia%2FTokyo")
		require.NoError(t, err)

		db, err := sql.Open("mysql", dsn)
		require.NoError(t, err)

		t.Cleanup(func() {
			require.NoError(t, db.Close())
		})

		repo, _ := newRepo(t, db)

		before := time.Now().Add(-time.Minute)

		require.NoError(t, repo.Write(t.Context(), []byte("a")))

		entries, err := repo.Claim(t.Context(), 1, time.Minute)
		require.NoError(t, err)
		require.Len(t, entries, 1)
		require.WithinRange(t, entries[0].CreatedAt, before, time.Now().Add(time.Minute))
	})
}
