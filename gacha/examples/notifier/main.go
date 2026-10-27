// Command notifier shows an event-driven poller: the writer notifies a poller.Notifier after every write, so the
// poller handles new entries right away instead of on its next tick. A ticker remains as a safety net, for retries and
// for entries written by other processes.
//
// It uses the in-memory repository, which stores the items themselves with gacha.Identity.
//
//	go run ./examples/notifier
package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/zalgonoise/x/gacha"
	"github.com/zalgonoise/x/gacha/memory"
	"github.com/zalgonoise/x/gacha/poller"
	"github.com/zalgonoise/x/gacha/writer"
)

const (
	messages  = 5
	interval  = 100 * time.Millisecond
	safetyNet = time.Minute
)

// Message records when it was written, to measure the delivery latency.
type Message struct {
	Text      string
	WrittenAt time.Time
}

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	repo := memory.New[Message]()
	notifier := poller.NewNotifier()

	w := writer.NewSync(repo, gacha.Identity[Message](), writer.WithNotify(notifier.Notify))

	handled := make(chan struct{}, messages)

	p := poller.New(repo, gacha.Identity[Message](), func(_ context.Context, m Message) error {
		fmt.Printf("handled %q %s after it was written\n", m.Text, time.Since(m.WrittenAt).Round(time.Microsecond))

		handled <- struct{}{}

		return nil
	}, poller.WithTrigger(poller.Any(notifier, poller.Ticker(safetyNet))))

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup

	wg.Go(func() {
		if err := p.Run(ctx); err != nil {
			log.Println("poller:", err)
		}
	})

	for i := range messages {
		time.Sleep(interval)

		if err := w.Write(ctx, Message{Text: fmt.Sprintf("message %d", i), WrittenAt: time.Now()}); err != nil {
			return err
		}

		<-handled
	}

	cancel()
	wg.Wait()

	return nil
}
