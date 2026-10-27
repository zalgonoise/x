// Command postgres-outbox places orders, writing each order and its outbox entry in the same transaction, and sends
// the order confirmations from a poller that PostgreSQL notifications wake up.
//
// If the business transaction rolls back, there is no entry to deliver. If it commits, the confirmation is delivered
// at least once, even if the process crashes right after the commit.
//
// It needs a PostgreSQL database:
//
//	GACHA_POSTGRES_DSN='postgres://user:password@localhost:5432/db?sslmode=disable' go run ./examples/postgres-outbox
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zalgonoise/x/gacha"
	"github.com/zalgonoise/x/gacha/poller"
	"github.com/zalgonoise/x/gacha/postgres"
	"github.com/zalgonoise/x/gacha/writer"
)

const (
	outboxTable = "orders_outbox"
	orders      = 3
	safetyNet   = 30 * time.Second
)

// OrderPlaced is the event delivered for every placed order.
type OrderPlaced struct {
	OrderID    string `json:"order_id"`
	Customer   string `json:"customer"`
	TotalCents int64  `json:"total_cents"`
}

var errNoDSN = errors.New("set GACHA_POSTGRES_DSN to a PostgreSQL connection string")

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	err := run(ctx)

	stop()

	if err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	dsn := os.Getenv("GACHA_POSTGRES_DSN")
	if dsn == "" {
		return errNoDSN
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return err
	}
	defer pool.Close()

	_, err = pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS orders (
		id          TEXT   PRIMARY KEY,
		customer    TEXT   NOT NULL,
		total_cents BIGINT NOT NULL
	)`)
	if err != nil {
		return err
	}

	// Every outbox write notifies the channel, on commit.
	repo, err := postgres.New(pool, outboxTable, postgres.WithNotify(outboxTable))
	if err != nil {
		return err
	}

	if err = repo.Migrate(ctx); err != nil {
		return err
	}

	listener, err := postgres.NewListener(pool, outboxTable)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	confirmed := make(chan string, orders)

	p := poller.New(repo, gacha.JSON[OrderPlaced](), func(ctx context.Context, event OrderPlaced) error {
		// The entry ID is stable across retries: use it as the idempotency key of the email provider.
		id, _ := gacha.IDFromContext(ctx)
		fmt.Printf("sending confirmation for %s to %s (idempotency key %s)\n", event.OrderID, event.Customer, id)

		confirmed <- event.OrderID

		return nil
	},
		// Wake up on notifications, and poll at least every 30 seconds for retries.
		poller.WithTrigger(poller.Any(listener, poller.Ticker(safetyNet))),
		poller.WithMaxAttempts(5),
		poller.WithRetryBackoff(poller.Jitter(poller.Exponential(time.Second, time.Minute), 0.2)),
	)

	var wg sync.WaitGroup

	wg.Go(func() {
		if err := listener.Run(ctx); err != nil {
			log.Println("listener:", err)
		}
	})

	wg.Go(func() {
		if err := p.Run(ctx); err != nil {
			log.Println("poller:", err)
		}
	})

	for i := range orders {
		order := OrderPlaced{
			OrderID:    fmt.Sprintf("order-%d-%d", time.Now().Unix(), i),
			Customer:   fmt.Sprintf("customer-%d@example.com", i),
			TotalCents: int64(1000 * (i + 1)),
		}

		if err = placeOrder(ctx, pool, repo, order); err != nil {
			return err
		}
	}

	for range orders {
		select {
		case id := <-confirmed:
			fmt.Println("confirmed", id)
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Stop gracefully: in-flight entries are settled before Run returns.
	cancel()
	wg.Wait()

	return nil
}

// placeOrder stores the order and its outbox entry atomically.
func placeOrder(ctx context.Context, pool *pgxpool.Pool, repo *postgres.Repository, order OrderPlaced) error {
	return pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, "INSERT INTO orders (id, customer, total_cents) VALUES ($1, $2, $3)",
			order.OrderID, order.Customer, order.TotalCents)
		if err != nil {
			return err
		}

		return writer.NewSync(repo.Tx(tx), gacha.JSON[OrderPlaced]()).Write(ctx, order)
	})
}
