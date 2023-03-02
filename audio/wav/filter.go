package wav

import (
	"bytes"
	"errors"
	"io"
	"time"

	"github.com/zalgonoise/gio"
)

// StreamFilter is a pluggable function that will scan, analyze, process
// or discard the input signal emitted by a ring buffer in an audio stream
type StreamFilter func(w *WavBuffer, data []int, raw []byte) error

// MaxValues sends to int channel `ch` the highest value in the data buffer
func MaxValues(ch chan<- int) StreamFilter {
	return func(w *WavBuffer, data []int, raw []byte) error {
		var max int
		for idx := range data {
			if data[idx] > max {
				max = data[idx]
			}
		}
		ch <- max
		return nil
	}
}

// MaxValuesTo writes the highest value in the data buffer to the input int writer
func MaxValuesTo(writer gio.Writer[int]) StreamFilter {
	return func(w *WavBuffer, data []int, raw []byte) error {
		var max = make([]int, 1)
		for idx := range data {
			if data[idx] > max[0] {
				max[0] = data[idx]
			}
		}
		_, err := writer.Write(max)
		if err != nil {
			return err
		}
		return nil
	}
}

// MinValues sends to int channel `ch` the lowest value in the data buffer
func MinValues(ch chan<- int) StreamFilter {
	return func(w *WavBuffer, data []int, raw []byte) error {
		var min int
		for idx := range data {
			if data[idx] < min {
				min = data[idx]
			}
		}
		ch <- min
		return nil
	}
}

// MinValuesTo writes the lowest value in the data buffer to the input int writer
func MinValuesTo(writer gio.Writer[int]) StreamFilter {
	return func(w *WavBuffer, data []int, raw []byte) error {
		var min = make([]int, 1)
		for idx := range data {
			if data[idx] < min[0] {
				min[0] = data[idx]
			}
		}
		_, err := writer.Write(min)
		if err != nil {
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
		return func(w *WavBuffer, data []int, raw []byte) error {
			for idx := range data {
				if data[idx] <= peak {
					err := fn(w, data, raw)
					if err != nil {
						return err
					}
					return nil
				}
			}
			return nil
		}
	}
	return func(w *WavBuffer, data []int, raw []byte) error {
		for idx := range data {
			if data[idx] >= peak {
				err := fn(w, data, raw)
				if err != nil {
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
	return func(w *WavBuffer, data []int, raw []byte) error {
		n := copy(dst, raw)
		if n != len(raw) {
			return ErrShortDataBuffer
		}
		return nil
	}
}

// FlushTo writes the raw signal to the input `writer`
func FlushTo(writer io.Writer) StreamFilter {
	return func(w *WavBuffer, data []int, raw []byte) error {
		n, err := writer.Write(raw)
		if err != nil {
			return err
		}
		if n != len(raw) {
			return ErrShortDataBuffer
		}
		return nil
	}
}

// FlushFor writes the raw signal to the input `writer`, then keeps recording
// from the WavBuffer reader for `dur` duration.
func FlushFor(writer io.Writer, dur time.Duration) StreamFilter {
	return func(w *WavBuffer, data []int, raw []byte) error {
		_, err := writer.Write(raw)
		if err != nil {
			return err
		}

		rate := (int64)(time.Second) / (int64)(w.Header.ByteRate)
		blockSize := (int64)(dur) / rate
		r := io.LimitReader(w.Reader, blockSize)
		_, err = io.Copy(writer, r)
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}
		return nil
	}
}

// FlushCh creates a new Wav object for the input data and sends it to
// the input Wav channel `ch`
func FlushCh(ch chan<- *Wav) StreamFilter {
	return func(w *WavBuffer, data []int, raw []byte) error {
		wav, err := New(w.Header.SampleRate, w.Header.BitsPerSample, w.Header.NumChannels)
		if err != nil {
			return err
		}
		wav.Chunks[0] = w.Data
		wav.Data = wav.Chunks[0]
		ch <- wav
		return nil
	}
}

// FlushFor creates a new Wav object for the input data, then keeps
// recording from the WavBuffer reader for `dur` duration.
//
// When done, it sends the created Wav to the input Wav channel `ch`
func FlushChFor(ch chan<- *Wav, dur time.Duration) StreamFilter {
	return func(w *WavBuffer, data []int, raw []byte) error {
		wav, err := New(w.Header.SampleRate, w.Header.BitsPerSample, w.Header.NumChannels)
		if err != nil {
			return err
		}
		wav.Chunks[0] = w.Data
		wav.Data = wav.Chunks[0]

		rate := (int64)(time.Second) / (int64)(w.Header.ByteRate)
		blockSize := (int64)(dur) / rate
		r := io.LimitReader(w.Reader, blockSize)
		buf := bytes.NewBuffer(make([]byte, 0, blockSize))
		_, err = io.Copy(buf, r)
		wav.Data.Parse(buf.Bytes())
		ch <- wav
		return nil
	}
}
