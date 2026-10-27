// Command janitor handles a few jobs from a SQLite outbox, then purges the completed jobs and the dead letters,
// archiving each one before deleting it.
//
// A service would run the janitors in the background, with a retention period, such as
// janitor.New(repo.Completed(7*24*time.Hour), archive).Run(ctx). This example runs a single pass of each, with no
// retention.
//
//	go run ./examples/janitor
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"

	"github.com/zalgonoise/x/gacha"
	"github.com/zalgonoise/x/gacha/janitor"
	"github.com/zalgonoise/x/gacha/poller"
	"github.com/zalgonoise/x/gacha/sqlite"
	"github.com/zalgonoise/x/gacha/writer"
)

const jobs = 5

// Job is a unit of background work.
type Job struct {
	ID   int    `json:"id"`
	Kind string `json:"kind"`
}

var errUnknownKind = errors.New("unknown job kind")

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) (err error) {
	dir, err := os.MkdirTemp("", "gacha-janitor")
	if err != nil {
		return err
	}

	defer func() {
		err = errors.Join(err, os.RemoveAll(dir))
	}()

	db, err := sql.Open("sqlite", "file:"+filepath.Join(dir, "jobs.db")+"?_pragma=journal_mode(WAL)")
	if err != nil {
		return err
	}

	defer func() {
		err = errors.Join(err, db.Close())
	}()

	repo, err := sqlite.New(db, "jobs_outbox")
	if err != nil {
		return err
	}

	if err = repo.Migrate(ctx); err != nil {
		return err
	}

	batch := make([]Job, jobs)
	for i := range batch {
		batch[i] = Job{ID: i, Kind: "resize"}
	}

	batch[2].Kind = "teleport"

	if err = writer.NewSync(repo, gacha.JSON[Job]()).Write(ctx, batch...); err != nil {
		return err
	}

	// Unknown kinds can never succeed: move them to the dead-letter table right away.
	p := poller.New(repo, gacha.JSON[Job](), func(_ context.Context, job Job) error {
		if job.Kind != "resize" {
			return gacha.Permanent(fmt.Errorf("job %d: %w %q", job.ID, errUnknownKind, job.Kind))
		}

		return nil
	})

	if _, err = p.Poll(ctx); err != nil {
		return err
	}

	if err = printStats(ctx, repo, "after handling"); err != nil {
		return err
	}

	// The archive hook runs before each deletion. Returning an error keeps the entry for the next run.
	archive := func(prefix string) gacha.Handler[[]byte] {
		return func(_ context.Context, data []byte) error {
			fmt.Printf("%s %s\n", prefix, data)

			return nil
		}
	}

	for _, j := range []*poller.Poller[[]byte, []byte]{
		janitor.New(repo.Completed(0), archive("archived completed job"), janitor.WithInterval(time.Hour)),
		janitor.New(repo.DeadLetters(0), archive("archived dead letter"), janitor.WithInterval(time.Hour)),
	} {
		if _, err = j.Poll(ctx); err != nil {
			return err
		}
	}

	return printStats(ctx, repo, "after the janitors")
}

func printStats(ctx context.Context, repo *sqlite.Repository, when string) error {
	stats, err := repo.Stats(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("%s: %d completed, %d dead letters\n", when, stats.Completed, stats.DeadLetters)

	return nil
}
