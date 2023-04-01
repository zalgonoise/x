package wav

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/zalgonoise/x/audio/wav/osc"
)

func TestWavBuffer(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// expect test to be faster than the actual length of the generated audio
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()

		// create a sine wave 16 bit depth, 44.1kHz rate, mono,
		// 5 second duration. Pass audio bytes into a new bytes.Buffer
		sine, err := New(44100, 16, 1)
		if err != nil {
			t.Error(err)
			return
		}
		sine.Data.Generate(osc.SineWave, 2000, 44100, 5*time.Second)
		buf := bytes.NewBuffer(sine.Bytes())

		// create a channel to read the max values emitted by the filter
		var maxCh = make(chan int, 2)
		defer close(maxCh)
		go func() {
			for i := range maxCh {
				t.Log(i)
			}
		}()

		// create a new stream using the bytes.Buffer as an io.Reader
		// half a second ratio (expect 11 entries), with a max values filter
		w := NewStream(buf).
			Ratio(0.5).
			WithFilter(
				MaxValues(maxCh),
			)

		// stream the audio using the context and an errors channel
		errCh := make(chan error)
		defer close(errCh)
		go w.Stream(ctx, errCh)

		// wait for the stream processing to end
		// expect an error (io.EOF) when the stream is consumed
		//
		// in case the context is done before an error is received,
		// it's surely a deadline reached error, as the test took too long
		select {
		case err := <-errCh:
			if errors.Is(err, io.EOF) {
				return
			}
			t.Errorf("unexpected error: wanted %v ; got %v", io.EOF, err)
			return
		case <-ctx.Done():
			err := ctx.Err()
			if err != nil {
				t.Error(err)
				return
			}
		}
	})

	t.Run("FailNoHeader", func(t *testing.T) {
		// expect test to be faster than the actual length of the generated audio
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()

		// create a sine wave 16 bit depth, 44.1kHz rate, mono,
		// 5 second duration. Pass audio bytes into a new bytes.Buffer
		// with no WAV header
		sine, err := New(44100, 16, 1)
		if err != nil {
			t.Error(err)
			return
		}
		sine.Data.Generate(osc.SineWave, 2000, 44100, 5*time.Second)
		buf := bytes.NewBuffer(sine.Data.Bytes())

		// create a channel to read the max values emitted by the filter
		var maxCh = make(chan int, 2)
		defer close(maxCh)
		go func() {
			for i := range maxCh {
				t.Log(i)
			}
		}()

		// create a new stream using the bytes.Buffer as an io.Reader
		w := NewStream(buf)

		// stream the audio using the context and an errors channel
		errCh := make(chan error)
		defer close(errCh)
		go w.Stream(ctx, errCh)

		// wait for the stream processing to end
		// expect an error (io.EOF) when the stream is consumed
		//
		// in case the context is done before an error is received,
		// it's surely a deadline reached error, as the test took too long
		select {
		case err := <-errCh:
			if errors.Is(err, ErrInvalid) && errors.Is(err, ErrHeader) {
				return
			}
			t.Errorf("unexpected error: wanted %v ; got %v", ErrInvalidHeader, err)
			return
		case <-ctx.Done():
			err := ctx.Err()
			if err != nil {
				t.Error(err)
				return
			}
		}
	})
}
