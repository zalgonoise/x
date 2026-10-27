// Command redrive walks through a dead letter's life: a payment in an unsupported currency fails permanently and is
// moved to the dead-letter table, an operator inspects it, a fix adds the currency, and the payment is redriven into
// the outbox, where it succeeds.
//
//	go run ./examples/redrive
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
	"uuid"

	_ "modernc.org/sqlite"

	"github.com/zalgonoise/x/gacha"
	"github.com/zalgonoise/x/gacha/poller"
	"github.com/zalgonoise/x/gacha/sqlite"
	"github.com/zalgonoise/x/gacha/writer"
)

const inspectLimit = 100

// Payment is a payment to execute.
type Payment struct {
	ID          string `json:"id"`
	Currency    string `json:"currency"`
	AmountCents int64  `json:"amount_cents"`
}

var errUnsupportedCurrency = errors.New("unsupported currency")

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) (err error) {
	dir, err := os.MkdirTemp("", "gacha-redrive")
	if err != nil {
		return err
	}

	defer func() {
		err = errors.Join(err, os.RemoveAll(dir))
	}()

	db, err := sql.Open("sqlite", "file:"+filepath.Join(dir, "payments.db")+"?_pragma=journal_mode(WAL)")
	if err != nil {
		return err
	}

	defer func() {
		err = errors.Join(err, db.Close())
	}()

	repo, err := sqlite.New(db, "payments_outbox")
	if err != nil {
		return err
	}

	if err = repo.Migrate(ctx); err != nil {
		return err
	}

	err = writer.NewSync(repo, gacha.JSON[Payment]()).Write(ctx,
		Payment{ID: "pay-1", Currency: "EUR", AmountCents: 1999},
		Payment{ID: "pay-2", Currency: "JPY", AmountCents: 120000},
		Payment{ID: "pay-3", Currency: "USD", AmountCents: 500},
	)
	if err != nil {
		return err
	}

	supported := map[string]bool{"EUR": true, "USD": true}

	p := poller.New(repo, gacha.JSON[Payment](), func(_ context.Context, payment Payment) error {
		if !supported[payment.Currency] {
			// Retrying cannot help until the code changes: dead-letter it now.
			return gacha.Permanent(fmt.Errorf("%w %s", errUnsupportedCurrency, payment.Currency))
		}

		fmt.Printf("executed %s: %d %s cents\n", payment.ID, payment.AmountCents, payment.Currency)

		return nil
	})

	res, err := p.Poll(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("first run: %d completed, %d dead-lettered\n", res.Completed, res.Failed)

	ids, err := inspect(ctx, repo)
	if err != nil {
		return err
	}

	// The fix is deployed: the currency is now supported.
	supported["JPY"] = true

	if err = repo.Redrive(ctx, ids...); err != nil {
		return err
	}

	if res, err = p.Poll(ctx); err != nil {
		return err
	}

	fmt.Printf("after the redrive: %d completed, %d dead-lettered\n", res.Completed, res.Failed)

	return nil
}

// inspect lists the dead letters through the DeadLetters view, and returns their IDs.
func inspect(ctx context.Context, repo *sqlite.Repository) ([]uuid.UUID, error) {
	view := repo.DeadLetters(0)

	entries, err := view.Claim(ctx, inspectLimit, time.Minute)
	if err != nil {
		return nil, err
	}

	ids := make([]uuid.UUID, 0, len(entries))
	release := make([]gacha.Outcome, 0, len(entries))

	for _, entry := range entries {
		fmt.Printf("dead letter %s after %d attempt(s): %s\n", entry.ID, entry.Attempts, entry.Data)

		ids = append(ids, entry.ID)
		// Hand the view's lease back right away: inspecting is not handling.
		release = append(release, gacha.Outcome{ID: entry.ID, Status: gacha.StatusPending})
	}

	return ids, view.Settle(ctx, release...)
}
