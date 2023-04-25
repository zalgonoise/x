package data

import (
	"bytes"
	"errors"
	"testing"
)

func TestChunkHeader(t *testing.T) {
	t.Run("HeaderFrom", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			_, err := HeaderFrom(test8bitHeader)
			if err != nil {
				t.Error(err)
				return
			}
		})

		t.Run("InvalidID", func(t *testing.T) {
			byteHeader := []byte("nope")                // add ID bytes
			byteHeader = append(byteHeader, 0, 0, 0, 0) // add size bytes

			_, err := HeaderFrom(byteHeader)
			if err == nil || !errors.Is(err, ErrInvalidSubChunkHeader) {
				t.Error("expected an error to be raised")
				return
			}
		})
	})

	t.Run("Bytes", func(t *testing.T) {
		t.Run("Data", func(t *testing.T) {
			var (
				header = NewDataHeader()
				output = header.Bytes()
			)

			wants := []byte(dataSubchunkIDString) // add ID bytes
			wants = append(wants, 0, 0, 0, 0)     // add size bytes

			if len(output) != len(wants) {
				t.Errorf("output length mismatch error: wanted %d ; got %d", len(wants), len(output))
			}

			if !bytes.Equal(wants, output) {
				t.Errorf("output mismatch error: wanted %v ; got %v", wants, output)
			}
		})
		t.Run("Junk", func(t *testing.T) {
			var (
				header = NewJunkHeader()
				output = header.Bytes()
			)

			wants := []byte(junkSubchunkIDString) // add ID bytes
			wants = append(wants, 0, 0, 0, 0)     // add size bytes

			if len(output) != len(wants) {
				t.Errorf("output length mismatch error: wanted %d ; got %d", len(wants), len(output))
			}

			if !bytes.Equal(wants, output) {
				t.Errorf("output mismatch error: wanted %v ; got %v", wants, output)
			}
		})
	})
}

func TestErr(t *testing.T) {
	wants := (string)(ErrInvalidSubChunkHeader)
	errString := ErrInvalidSubChunkHeader.Error()

	if errString != wants {
		t.Errorf("output mismatch error: wanted %s ; got %s", wants, errString)
	}
}
