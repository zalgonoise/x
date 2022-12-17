package spanner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/zalgonoise/attr"
	"github.com/zalgonoise/logx"
)

const (
	maxBatchSize  = 2048
	maxExportSize = 1024
	defaultDelay  = 5 * time.Second
	maxTimeout    = 30 * time.Second
)

var (
	zeroTime time.Time
)

// SpanProcessor will handle routing ended Spans to an Exporter
type SpanProcessor interface {
	// Handle routes the input Span `span` to the SpanProcessor's Exporter
	Handle(span Span)
	// Shutdown gracefully stops the SpanProcessor, returning an error
	Shutdown(ctx context.Context) error
	// Flush will force-push the existing SpanData in the SpanProcessor's batch into the
	// Exporter, even if not yet scheduled to do so
	Flush(ctx context.Context) error
}

type processor struct {
	sync.Mutex
	stopOnce sync.Once
	rec      bool
	exporter Exporter
	stopCh   chan struct{}
	queue    chan Span
	timer    *time.Timer
	batch    []SpanData
}

// NewProcessor creates a new SpanProcessor configured with the input Exporter `e`
func NewProcessor(e Exporter) SpanProcessor {
	p := &processor{
		exporter: e,
		stopCh:   make(chan struct{}),
		queue:    make(chan Span),
		timer:    time.NewTimer(defaultDelay),
		batch:    make([]SpanData, 0, maxBatchSize),
	}

	go p.runtime()
	p.rec = true

	return p
}

// Handle routes the input Span `span` to the SpanProcessor's Exporter
func (p *processor) Handle(span Span) {
	if p.rec {
		p.queue <- span
	}
}

// Shutdown gracefully stops the SpanProcessor, returning an error
func (p *processor) Shutdown(ctx context.Context) error {
	p.Lock()
	p.rec = false
	p.Unlock()

	var err error
	p.stopOnce.Do(func() {
		wait := make(chan struct{})
		go func() {
			p.stopCh <- struct{}{}
			err := p.export(ctx)
			exporterErr := p.exporter.Shutdown(ctx)
			if exporterErr != nil {
				if err != nil {
					err = fmt.Errorf("%w -- %v", exporterErr, err)
				}
			}
			close(wait)
		}()
		select {
		case <-wait:
		case <-ctx.Done():
			if cErr := ctx.Err(); cErr != nil {
				if err != nil {
					err = fmt.Errorf("%w -- %v", cErr, err)
				} else {
					err = cErr
				}
			}
		}
	})

	return err
}

// Flush will force-push the existing SpanData in the SpanProcessor's batch into the
// Exporter, even if not yet scheduled to do so
func (p *processor) Flush(ctx context.Context) error {
	return p.export(ctx)
}

func (p *processor) runtime() {
	defer p.timer.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		select {
		// shutdown signal
		case <-p.stopCh:
			return
		// export triggered
		case <-p.timer.C:
			err := p.export(ctx)
			if err != nil {
				logx.Error("[processor] spanner export failed", attr.String("error", err.Error()))
			}
		// span enqueued
		case span := <-p.queue:
			sd := span.Extract()
			p.Lock()
			p.batch = append(p.batch, sd)
			toExport := len(p.batch) >= maxExportSize
			p.Unlock()

			if toExport {
				if !p.timer.Stop() {
					<-p.timer.C
				}
				err := p.export(ctx)
				if err != nil {
					logx.Error("[processor] spanner export failed", attr.String("error", err.Error()))
				}
			}
		}
	}
}

func (p *processor) export(ctx context.Context) error {
	p.timer.Reset(defaultDelay)

	ctx, cancel := context.WithTimeout(ctx, maxTimeout)
	defer cancel()

	if len(p.batch) > 0 {
		err := p.exporter.Export(ctx, p.batch)
		p.batch = p.batch[:0]
		if err != nil {
			return err
		}
	}

	return nil
}
