package stream_test

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"io"
	"testing"
	"time"

	wav2 "github.com/zalgonoise/x/audio/encoding/wav"
	stream2 "github.com/zalgonoise/x/audio/encoding/wav/stream"
	"github.com/zalgonoise/x/audio/fft"
	"github.com/zalgonoise/x/audio/osc"
)

//go:embed internal/testdata/2khz.wav
var sine2 []byte

//go:embed internal/testdata/4khz.wav
var sine4 []byte

//go:embed internal/testdata/8khz.wav
var sine8 []byte

//go:embed internal/testdata/16khz.wav
var sine16 []byte

func newSine(freq int) (*wav2.Wav, error) {
	// create a sine wave 16 bit depth, 44.1kHz rate, mono,
	// 5 second duration. Pass audio bytes into a new bytes.Buffer
	sine, err := wav2.New(44100, 16, 1, 1)
	if err != nil {
		return nil, err
	}
	sine.Generate(osc.SineWave, freq, 5*time.Second)
	return sine, nil
}

func TestWavBuffer(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("MaxValues", func(t *testing.T) {
			// expect test to be faster than the actual length of the generated audio
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*300)
			defer cancel()

			buf := bytes.NewReader(sine8)

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
			w := stream2.New(buf).
				Ratio(0.5).
				WithFilter(
					stream2.MaxValues(maxCh),
				)

			// stream the audio using the context and an err channel
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

		t.Run("FFTOnThreshold", func(t *testing.T) {
			for _, testdata := range []struct {
				name string
				data []byte
				freq int
			}{
				{
					name: "2kHz",
					data: sine2,
					freq: 2000,
				}, {
					name: "4kHz",
					data: sine4,
					freq: 4000,
				}, {
					name: "8kHz",
					data: sine8,
					freq: 8000,
				}, {
					name: "16kHz",
					data: sine16,
					freq: 16000,
				},
			} {
				t.Run(testdata.name, func(t *testing.T) {
					// expect test to be faster than the actual length of the generated audio
					// create a channel to read the max values emitted by the filter

					var (
						ctx, cancel      = context.WithTimeout(context.Background(), time.Second*30)
						maxCh            = make(chan fft.FrequencyPower, 5)
						detectionCounter int
						drift            = 200
					)

					defer cancel()
					defer close(maxCh)

					buf := bytes.NewReader(testdata.data)

					// goroutine to verify emitted FrequencyPower objects
					go func() {
						for i := range maxCh {
							if i.Freq < testdata.freq-drift || i.Freq > testdata.freq+drift {
								t.Errorf("frequency is off: emitted %dHz ; got %dHz", testdata.freq, i.Freq)
							}
							detectionCounter++
						}
					}()

					// create a new stream using the bytes.Buffer as an io.Reader
					w := stream2.New(buf).
						WithFilter(
							stream2.FFTOnThreshold(fft.Block1024, 10, maxCh),
						)

					// stream the audio using the context and an err channel
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
						if !errors.Is(err, io.EOF) {
							t.Errorf("unexpected error: wanted %v ; got %v", io.EOF, err)
							return
						}
						if detectionCounter == 0 {
							t.Errorf("expected detector to increase in value during the test")
						}
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

		})
	})

	t.Run("FailNoHeader", func(t *testing.T) {
		// expect test to be faster than the actual length of the generated audio
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()

		// sine wave stream for testing
		sine, err := newSine(2000)
		if err != nil {
			t.Error(err)
			return
		}
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
		w := stream2.New(buf)

		// stream the audio using the context and an err channel
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
			if errors.Is(err, wav2.ErrInvalid) && errors.Is(err, wav2.ErrHeader) {
				return
			}
			t.Errorf("unexpected error: wanted %v ; got %v", wav2.ErrInvalidHeader, err)
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
