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
	maxBatchSize  = 1024
	maxExportSize = 512
	defaultDelay  = 5 * time.Second
	maxTimeout    = 30 * time.Second
)

var (
	zeroTime time.Time
)

type SpanProcessor interface {
	Handle(span SpanData)
	Shutdown(ctx context.Context) error
	Flush(ctx context.Context) error
}

type processor struct {
	sync.Mutex
	stopOnce sync.Once
	rec      bool
	exporter Exporter
	stopCh   chan struct{}
	queue    chan SpanData
	timer    *time.Timer
	batch    []SpanData
}

func NewProcessor(e Exporter) SpanProcessor {
	p := &processor{
		exporter: e,
		stopCh:   make(chan struct{}),
		queue:    make(chan SpanData),
		timer:    time.NewTimer(defaultDelay),
		batch:    make([]SpanData, 0, maxBatchSize),
	}

	go p.runtime()
	p.rec = true

	return p
}

func (p *processor) Handle(span SpanData) {
	if span.StartTime == zeroTime {
		return
	}
	if p.rec {
		p.queue <- span
	}
}

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
			p.Lock()
			p.batch = append(p.batch, span)
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
