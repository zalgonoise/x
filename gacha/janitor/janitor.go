// Package janitor purges completed and dead-lettered entries on a schedule, optionally archiving them first.
//
// A janitor is a regular poller.Poller over one of a repository's views (see gacha.Viewer): the Completed view for
// the outbox's audit trail, or the DeadLetters view for the dead-letter table. Handling an entry archives it, and
// completing it deletes it from the view.
//
//	purge := janitor.New(repo.Completed(7*24*time.Hour), nil, janitor.WithInterval(time.Hour))
//
//	go purge.Run(ctx)
package janitor

import (
	"context"

	"github.com/zalgonoise/cfg"

	"github.com/zalgonoise/x/gacha"
	"github.com/zalgonoise/x/gacha/poller"
)

// New returns a poller that archives the entries of view with archive, then deletes them. A nil archive deletes the
// entries without archiving them.
//
// An archive error retries the entry at the next run, and an error marked with gacha.Permanent leaves the entry in
// place until its lease expires. Either way, nothing is deleted before it is archived.
//
// The poller runs every WithInterval, draining the view in batches of WithBatchSize, and settles each batch in one
// call (a single bulk delete). It applies no maximum attempt count, since view entries keep their final attempt
// count from the outbox.
func New[R any](view gacha.Source[R], archive gacha.Handler[R], opts ...cfg.Option[Config]) *poller.Poller[R, R] {
	config := cfg.Set(defaultConfig(), opts...)

	if archive == nil {
		archive = func(context.Context, R) error {
			return nil
		}
	}

	pollerOpts := append([]cfg.Option[poller.Config]{
		poller.WithTrigger(poller.Ticker(config.interval)),
		poller.WithBatchSize(config.batchSize),
		poller.WithLease(config.lease),
		poller.WithMaxAttempts(0),
		poller.WithRetryBackoff(poller.Constant(config.interval)),
		poller.WithBatchSettle(),
	}, config.extra...)

	return poller.New(view, gacha.Identity[R](), archive, pollerOpts...)
}
