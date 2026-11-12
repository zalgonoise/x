package wav_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"math"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/zalgonoise/x/audio/encoding/wav"
	"github.com/zalgonoise/x/audio/internal/fixtures"
	"github.com/zalgonoise/x/audio/internal/testutil"
)

type ctxKey struct{}

func TestStreamLifecycle(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	file := fixtures.WAV(t, "amen_kick/amen_kick_mono_16bit_44100hz.wav")

	t.Run("ReturnsAfterEOFWithCallerContext", func(t *testing.T) {
		var sawValue bool

		s := wav.NewStreamContext(func(ctx context.Context, _ *wav.Header, _ []float64) error {
			sawValue = ctx.Value(ctxKey{}) == "caller"

			_, hasID := wav.GetID(ctx)
			require.True(t, hasID, "the processor context carries a stream ID")

			return nil
		})

		errCh := make(chan error, 1)
		ctx := context.WithValue(context.Background(), ctxKey{}, "caller")

		done := make(chan struct{})

		go func() {
			defer close(done)

			s.Stream(ctx, bytes.NewReader(file), errCh)
		}()

		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Stream did not return after EOF")
		}

		require.True(t, sawValue, "the caller's context reaches the processor")

		select {
		case err, ok := <-errCh:
			t.Fatalf("unexpected error (or closed channel): %v %v", err, ok)
		default:
		}
	})

	t.Run("ProcessorErrorDeliveredOnceChannelNotClosed", func(t *testing.T) {
		boom := errors.New("boom")
		s := wav.NewStream(func(*wav.Header, []float64) error { return boom }, wav.WithSize(64))
		errCh := make(chan error, 4)

		s.Stream(context.Background(), bytes.NewReader(file), errCh)

		require.ErrorIs(t, <-errCh, boom)

		select {
		case err, ok := <-errCh:
			t.Fatalf("got a second value (%v) or a closed channel (ok=%v)", err, ok)
		default:
		}

		errCh <- nil // the caller still owns (and can use) the channel
	})

	t.Run("WaitsForInFlightProcessorOnCancel", func(t *testing.T) {
		// Stream must not return while a processor call is running: callers shut their exporters down right after
		var running atomic.Int32

		entered := make(chan struct{}, 1)

		s := wav.NewStreamContext(func(ctx context.Context, _ *wav.Header, _ []float64) error {
			running.Add(1)
			defer running.Add(-1)

			select {
			case entered <- struct{}{}:
			default:
			}

			<-ctx.Done() // a well-behaved processor returns once its context is done
			time.Sleep(20 * time.Millisecond)

			return ctx.Err()
		}, wav.WithSize(64))

		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan struct{})

		go func() {
			defer close(done)

			s.Stream(ctx, bytes.NewReader(file), nil)
		}()

		<-entered
		cancel()

		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Stream did not return after the processor honoured the cancellation")
		}

		require.Zero(t, running.Load(), "Stream returned while the processor was still running")
	})

	t.Run("ReturnsOnCancelWithBlockedSource", func(t *testing.T) {
		// a live source: the header and some audio, then a Read that blocks forever (it never sees EOF)
		pr, pw := io.Pipe()
		live := testutil.LiveHeader(file, math.MaxUint32)

		// the writer blocks once the payload is consumed (nobody closes the pipe until the end); its error is the
		// ErrClosedPipe of the cleanup below
		go func() {
			if _, err := pw.Write(live); err != nil {
				return
			}
		}()

		var calls, afterCancel atomic.Int32

		ctx, cancel := context.WithCancel(context.Background())
		canceled := make(chan struct{})

		s := wav.NewStream(func(*wav.Header, []float64) error {
			calls.Add(1)

			select {
			case <-canceled:
				afterCancel.Add(1)
			default:
			}

			return nil
		}, wav.WithSize(4096))

		errCh := make(chan error, 1)
		done := make(chan struct{})

		go func() {
			defer close(done)

			s.Stream(ctx, pr, errCh)
		}()

		require.Eventually(t, func() bool { return calls.Load() > 0 }, 5*time.Second, time.Millisecond)

		close(canceled)
		cancel()

		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Stream did not return after cancel while its source was blocked")
		}

		require.NoError(t, pw.CloseWithError(io.ErrClosedPipe)) // let the abandoned read goroutine exit
		require.Zero(t, afterCancel.Load(), "no processor call starts after cancel")

		select {
		case err := <-errCh:
			t.Fatalf("cancellation is not an error: %v", err)
		default:
		}
	})
}
