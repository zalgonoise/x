// Command batch-writer compares the Sync and Batch writers under concurrent writes, on a SQLite outbox.
//
// Each Sync write is its own transaction, so concurrent producers queue behind SQLite's single writer lock and pay
// for one durable commit each. The Batch writer coalesces their writes into shared transactions (group commit), so
// the same events take a fraction of the commits.
//
//	go run ./examples/batch-writer
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	_ "modernc.org/sqlite"

	"github.com/zalgonoise/x/gacha"
	"github.com/zalgonoise/x/gacha/sqlite"
	"github.com/zalgonoise/x/gacha/writer"
)

const (
	producers = 32
	events    = 50
	linger    = 2 * time.Millisecond
)

// Event is a domain event produced by the application.
type Event struct {
	Producer int `json:"producer"`
	Sequence int `json:"sequence"`
}

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) (err error) {
	dir, err := os.MkdirTemp("", "gacha-batch-writer")
	if err != nil {
		return err
	}

	defer func() {
		err = errors.Join(err, os.RemoveAll(dir))
	}()

	db, err := sql.Open("sqlite",
		"file:"+filepath.Join(dir, "outbox.db")+"?_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)")
	if err != nil {
		return err
	}

	defer func() {
		err = errors.Join(err, db.Close())
	}()

	for _, name := range []string{"sync", "batch"} {
		repo, err := sqlite.New(db, name+"_outbox")
		if err != nil {
			return err
		}

		if err = repo.Migrate(ctx); err != nil {
			return err
		}

		// Count the repository writes, to show the coalescing.
		var writes atomic.Int64

		sink := gacha.SinkFunc[[]byte](func(ctx context.Context, data ...[]byte) error {
			writes.Add(1)

			return repo.Write(ctx, data...)
		})

		var w gacha.Writer[Event]

		closeWriter := func() error { return nil }

		if name == "sync" {
			w = writer.NewSync(sink, gacha.JSON[Event]())
		} else {
			batch := writer.NewBatch(sink, gacha.JSON[Event](), writer.WithLinger(linger))
			w, closeWriter = batch, func() error { return batch.Close(ctx) }
		}

		elapsed, err := produce(ctx, w)
		if err != nil {
			return err
		}

		if err = closeWriter(); err != nil {
			return err
		}

		stats, err := repo.Stats(ctx)
		if err != nil {
			return err
		}

		fmt.Printf("%-5s wrote %d events in %d repository writes, in %s\n",
			name, stats.Pending, writes.Load(), elapsed.Round(time.Millisecond))
	}

	return nil
}

// produce writes events from concurrent producers, each one waiting for its write to be durable.
func produce(ctx context.Context, w gacha.Writer[Event]) (time.Duration, error) {
	var (
		wg       sync.WaitGroup
		firstErr atomic.Pointer[error]
	)

	start := time.Now()

	for producer := range producers {
		wg.Go(func() {
			for sequence := range events {
				if err := w.Write(ctx, Event{Producer: producer, Sequence: sequence}); err != nil {
					firstErr.CompareAndSwap(nil, &err)

					return
				}
			}
		})
	}

	wg.Wait()

	if err := firstErr.Load(); err != nil {
		return 0, *err
	}

	return time.Since(start), nil
}
