package stream

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/zalgonoise/gio"

	"github.com/zalgonoise/x/audio/fft"
	"github.com/zalgonoise/x/audio/fft/window"
	"github.com/zalgonoise/x/audio/wav"
	datah "github.com/zalgonoise/x/audio/wav/data/header"
)

// StreamFilter is a pluggable function that will scan, analyze, process
// or discard the input signal emitted by a ring buffer in an audio stream
type StreamFilter func(w *Wav, raw []byte) error

// MaxValues sends to int channel `ch` the highest value in the data buffer
func MaxValues(ch chan<- int) StreamFilter {
	return func(w *Wav, raw []byte) error {
		v := w.Data.Value()
		var max int
		for idx := range v {
			if v[idx] > max {
				max = v[idx]
			}
		}
		ch <- max
		return nil
	}
}

// MaxValuesTo writes the highest value in the data buffer to the input int writer
func MaxValuesTo(writer gio.Writer[int]) StreamFilter {
	return func(w *Wav, raw []byte) error {
		v := w.Data.Value()
		var max = make([]int, 1)
		for idx := range v {
			if v[idx] > max[0] {
				max[0] = v[idx]
			}
		}
		if _, err := writer.Write(max); err != nil {
			return err
		}
		return nil
	}
}

// MinValues sends to int channel `ch` the lowest value in the data buffer
func MinValues(ch chan<- int) StreamFilter {
	return func(w *Wav, raw []byte) error {
		v := w.Data.Value()
		var min int
		for idx := range v {
			if v[idx] < min {
				min = v[idx]
			}
		}
		ch <- min
		return nil
	}
}

// MinValuesTo writes the lowest value in the data buffer to the input int writer
func MinValuesTo(writer gio.Writer[int]) StreamFilter {
	return func(w *Wav, raw []byte) error {
		v := w.Data.Value()
		var min = make([]int, 1)
		for idx := range v {
			if v[idx] < min[0] {
				min[0] = v[idx]
			}
		}
		if _, err := writer.Write(min); err != nil {
			return err
		}
		return nil
	}
}

// LevelThreshold triggers the input StreamFilter `fn` whenever the signal is
// either equal and below or equal and above the threshold level `peak`.
//
// It will look for equal and below (item <= peak) if the peak is a negative value
// it will look for equal and above (item >= peak) if the peak is a positive value
func LevelThreshold(peak int, fn StreamFilter) StreamFilter {
	if peak < 0 {
		return func(w *Wav, raw []byte) error {
			v := w.Data.Value()
			for idx := range v {
				if v[idx] <= peak {
					if err := fn(w, raw); err != nil {
						return err
					}
					return nil
				}
			}
			return nil
		}
	}
	return func(w *Wav, raw []byte) error {
		v := w.Data.Value()
		for idx := range v {
			if v[idx] >= peak {
				if err := fn(w, raw); err != nil {
					return err
				}
				return nil
			}
		}
		return nil
	}
}

// LevelThresholdFn triggers the input StreamFilter `fn` whenever the signal is
// either equal and below or equal and above the threshold level `peak`.
//
// Before calling this StreamFilter function, it will first call the input `alertFn`
// that emits the item that surpasses the peak threshold value.
//
// It will look for equal and below (item <= peak) if the peak is a negative value
// it will look for equal and above (item >= peak) if the peak is a positive value
func LevelThresholdFn(peak int, alertFn func(int), fn StreamFilter) StreamFilter {
	if peak < 0 {
		return func(w *Wav, raw []byte) error {
			v := w.Data.Value()
			for idx := range v {
				if v[idx] <= peak {
					alertFn(v[idx])
					if err := fn(w, raw); err != nil {
						return err
					}
					return nil
				}
			}
			return nil
		}
	}
	return func(w *Wav, raw []byte) error {
		v := w.Data.Value()
		for idx := range v {
			if v[idx] >= peak {
				alertFn(v[idx])
				if err := fn(w, raw); err != nil {
					return err
				}
				return nil
			}
		}
		return nil
	}
}

// Flush writes the raw signal to the input buffer `dst`, returning an
// error if it is a short write
func Flush(dst []byte) StreamFilter {
	return func(w *Wav, raw []byte) error {
		n := copy(dst, raw)
		if n != len(raw) {
			return wav.ErrShortDataBuffer
		}
		return nil
	}
}

// FlushTo writes the raw signal to the input `writer`
func FlushTo(writer io.Writer) StreamFilter {
	return func(w *Wav, raw []byte) error {
		n, err := writer.Write(raw)
		if err != nil {
			return err
		}
		if n != len(raw) {
			return wav.ErrShortDataBuffer
		}
		return nil
	}
}

// FlushFor writes the raw signal to the input `writer`, then keeps recording
// from the WavBuffer reader for `dur` duration.
func FlushFor(writer io.Writer, dur time.Duration) StreamFilter {
	return func(w *Wav, raw []byte) error {
		var err error

		rate := (int64)(time.Second) / (int64)(w.Header.ByteRate)
		blockSize := (int64)(dur) / rate

		w.Header.ChunkSize = uint32(blockSize) + uint32(len(raw)) + 4
		if _, err = writer.Write(w.Header.Bytes()); err != nil {
			return err
		}

		dataHeader := datah.NewData()
		dataHeader.Subchunk2Size = uint32(blockSize) + uint32(len(raw))
		if _, err = writer.Write(dataHeader.Bytes()); err != nil {
			return err
		}

		if _, err = writer.Write(raw); err != nil {
			return err
		}

		rec := bytes.NewBuffer(make([]byte, 0, blockSize))
		r := io.LimitReader(w.Reader, blockSize)
		if _, err = rec.ReadFrom(r); err != nil && !errors.Is(err, io.EOF) {
			return err
		}

		if _, err = writer.Write(rec.Bytes()); err != nil {
			return err
		}

		if c, ok := (writer).(io.Closer); ok {
			err = c.Close()
			if err != nil {
				return err
			}
		}

		return nil
	}
}

// FlushToFileFor writes the raw signal to a file with the path and name pattern `name`,
// then keeps recording from the WavBuffer reader for `dur` duration.
func FlushToFileFor(name string, dur time.Duration) StreamFilter {
	return func(w *Wav, raw []byte) error {
		var err error
		writer, err := os.Create(fmt.Sprintf("%s_%s.wav", name, time.Now().Format(time.RFC3339)))

		rate := (int64)(time.Second) / (int64)(w.Header.ByteRate)
		blockSize := (int64)(dur) / rate

		w.Header.ChunkSize = uint32(blockSize) + uint32(len(raw)) + 4
		if _, err = writer.Write(w.Header.Bytes()); err != nil {
			return err
		}

		dataHeader := datah.NewData()
		dataHeader.Subchunk2Size = uint32(blockSize) + uint32(len(raw))
		if _, err = writer.Write(dataHeader.Bytes()); err != nil {
			return err
		}

		if _, err = writer.Write(raw); err != nil {
			return err
		}

		rec := bytes.NewBuffer(make([]byte, 0, blockSize))
		r := io.LimitReader(w.Reader, blockSize)
		if _, err = rec.ReadFrom(r); err != nil && !errors.Is(err, io.EOF) {
			return err
		}

		if _, err = writer.Write(rec.Bytes()); err != nil {
			return err
		}

		return writer.Close()
	}
}

// FlushCh creates a new Wav object for the input data and sends it to
// the input wav.Wav channel `ch`
func FlushCh(ch chan<- *wav.Wav) StreamFilter {
	return func(w *Wav, raw []byte) error {
		wav, err := wav.New(w.Header.SampleRate, w.Header.BitsPerSample, w.Header.NumChannels, w.Header.AudioFormat)
		if err != nil {
			return err
		}
		wav.Chunks[0] = w.Data
		wav.Data = wav.Chunks[0]
		ch <- wav
		return nil
	}
}

// FlushChFor creates a new Wav object for the input data, then keeps
// recording from the WavBuffer reader for `dur` duration.
//
// When done, it sends the created Wav to the input Wav channel `ch`
func FlushChFor(ch chan<- *wav.Wav, dur time.Duration) StreamFilter {
	return func(w *Wav, raw []byte) error {
		wav, err := wav.New(w.Header.SampleRate, w.Header.BitsPerSample, w.Header.NumChannels, w.Header.AudioFormat)
		if err != nil {
			return err
		}
		wav.Chunks[0] = w.Data
		wav.Data = wav.Chunks[0]

		rate := (int64)(time.Second) / (int64)(w.Header.ByteRate)
		blockSize := (int64)(dur) / rate
		r := io.LimitReader(w.Reader, blockSize)
		buf := bytes.NewBuffer(make([]byte, 0, blockSize))
		if _, err = io.Copy(buf, r); err != nil {
			return err
		}
		wav.Data.Parse(buf.Bytes())
		ch <- wav
		return nil
	}
}

// FFT analyzes the frequency spectrum on each data chunk emitted by the buffer,
// with precision according to the configured block size.
//
// Each pass will emit FrequencyPower items for each frequency that is over the
// default magnitude value
func FFT(blockSize fft.BlockSize, ch chan<- fft.FrequencyPower) StreamFilter {
	return func(w *Wav, raw []byte) error {
		v := w.Data.Float()
		for i := 0; i < len(v); i += int(blockSize) {
			if len(v) < i+int(blockSize) {
				break
			}

			// get precomputed window if it exists; with fallback to creating one
			var windowBlock = window.New(window.Blackman, int(blockSize))

			mag := fft.Apply(int(w.Header.SampleRate), v[i:i+int(blockSize)], windowBlock)
			for i := range mag {
				if mag[i].Mag > fft.DefaultMagnitudeThreshold {
					ch <- mag[i]
				}
			}
		}
		return nil
	}
}

// FFTOnThreshold analyzes the frequency spectrum on each data chunk emitted by
// the buffer, with precision according to the configured block size.
//
// Each pass will emit FrequencyPower items for each frequency that is over the
// input magnitude threshold value
func FFTOnThreshold(blockSize fft.BlockSize, thresh float64, ch chan<- fft.FrequencyPower) StreamFilter {
	return func(w *Wav, raw []byte) error {
		v := w.Data.Float()
		for i := 0; i < len(v); i += int(blockSize) {
			if len(v) < i+int(blockSize) {
				break
			}

			// get precomputed window if it exists; with fallback to creating one
			var windowBlock = window.New(window.Blackman, int(blockSize))

			mag := fft.Apply(int(w.Header.SampleRate), v[i:i+int(blockSize)], windowBlock)
			for idx := range mag {
				if mag[idx].Mag > thresh {
					ch <- mag[idx]
				}
			}
		}
		return nil
	}
}

// Spectrum analyzes the frequency spectrum on each data chunk emitted by
// the buffer, with precision according to the configured block size.
//
// Each pass will emit unfiltered FrequencyPower items
func Spectrum(blockSize fft.BlockSize, ch chan<- []fft.FrequencyPower) StreamFilter {
	return func(w *Wav, raw []byte) error {
		v := w.Data.Float()
		for i := 0; i < len(v); i += int(blockSize) {
			if len(v) < i+int(blockSize) {
				break
			}

			// get precomputed window if it exists; with fallback to creating one
			var windowBlock = window.New(window.Blackman, int(blockSize))

			mag := fft.Apply(int(w.Header.SampleRate), v[i:i+int(blockSize)], windowBlock)
			ch <- mag

		}
		return nil
	}
}
