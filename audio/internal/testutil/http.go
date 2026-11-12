package testutil

import (
	"bytes"
	"encoding/binary"
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

// ServeOption configures a WAVServer.
type ServeOption func(*serveConfig)

type serveConfig struct {
	status          int
	contentType     string
	chunk           int
	flush           bool
	delay           time.Duration
	endless         bool
	endlessSize     uint32
	disconnectAfter int
}

// Chunked writes the body in chunks of n bytes, flushing after each one (chunked transfer encoding).
func Chunked(n int) ServeOption {
	return func(c *serveConfig) { c.chunk, c.flush = n, true }
}

// FlushEvery is an alias of Chunked, for readability at call sites that care about flush boundaries.
func FlushEvery(n int) ServeOption { return Chunked(n) }

// Delay sleeps for d between chunks.
func Delay(d time.Duration) ServeOption {
	return func(c *serveConfig) { c.delay = d }
}

// Endless serves the file's header once, with the data chunk size rewritten to dataSize (0 or 0xFFFFFFFF, as live
// streaming servers send), and then repeats the data payload forever. The payload is cut at the data chunk's real
// size, so its pad byte and any chunks after it are never replayed as audio. Pass a normal file, not a LiveHeader one.
func Endless(dataSize uint32) ServeOption {
	return func(c *serveConfig) { c.endless, c.endlessSize = true, dataSize }
}

// DisconnectAfter aborts the response (closing the connection) after n bytes of body.
func DisconnectAfter(n int) ServeOption {
	return func(c *serveConfig) { c.disconnectAfter = n }
}

// Status sets the response status code (default 200).
func Status(code int) ServeOption {
	return func(c *serveConfig) { c.status = code }
}

// ContentType sets the Content-Type header (default "audio/wav").
func ContentType(ct string) ServeOption {
	return func(c *serveConfig) { c.contentType = ct }
}

// WAVServer is an httptest.Server serving a WAV file, with a counter of served requests (useful for reconnect tests).
type WAVServer struct {
	*httptest.Server

	requests atomic.Int64
}

// Requests returns how many requests the server has received.
func (s *WAVServer) Requests() int64 { return s.requests.Load() }

// NewWAVServer starts a server that serves body at every path. It is closed when the test ends.
func NewWAVServer(tb testing.TB, body []byte, opts ...ServeOption) *WAVServer {
	tb.Helper()

	cfg := serveConfig{status: http.StatusOK, contentType: "audio/wav"}
	for _, o := range opts {
		o(&cfg)
	}

	srv := &WAVServer{}
	srv.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		srv.requests.Add(1)

		w.Header().Set("Content-Type", cfg.contentType)

		if !cfg.flush && !cfg.endless && cfg.disconnectAfter == 0 {
			w.Header().Set("Content-Length", strconv.Itoa(len(body)))
		}

		w.WriteHeader(cfg.status)

		payload := body
		header, data := endlessParts(body, cfg)

		written := 0
		write := func(p []byte) bool {
			for len(p) > 0 {
				n := len(p)
				if cfg.chunk > 0 {
					n = min(n, cfg.chunk)
				}

				if cfg.disconnectAfter > 0 && written+n >= cfg.disconnectAfter {
					// send the headers and the partial body first, then close the connection mid-body
					if _, err := w.Write(p[:cfg.disconnectAfter-written]); err == nil {
						flush(w)
					}

					panic(http.ErrAbortHandler)
				}

				if _, err := w.Write(p[:n]); err != nil {
					return false
				}

				written += n
				p = p[n:]

				if cfg.flush {
					flush(w)
				}

				if cfg.delay > 0 {
					select {
					case <-r.Context().Done():
						return false
					case <-time.After(cfg.delay):
					}
				}
			}

			return true
		}

		if !cfg.endless {
			write(payload)

			return
		}

		if !write(header) {
			return
		}

		for r.Context().Err() == nil {
			if !write(data) {
				return
			}
		}
	}))

	tb.Cleanup(srv.Close)

	return srv
}

// endlessParts returns the header (with the data size rewritten for Endless) and the payload to repeat.
func endlessParts(body []byte, cfg serveConfig) (header, data []byte) {
	header, data, found := splitAtData(body)

	if cfg.endless && found {
		header = bytes.Clone(header)
		setLiveSizes(header, cfg.endlessSize)
	}

	return header, data
}

// flush sends buffered response data to the client. Errors are ignored: a client that went away ends the handler on
// its next Write.
func flush(w http.ResponseWriter) {
	http.NewResponseController(w).Flush()
}

// splitAtData splits a RIFF/WAVE file into everything up to and including the data chunk header, and the data
// payload (bounded by its declared size; a size of 0 or 0xFFFFFFFF runs to the end). found is false, and the whole
// body is returned as the header, when no complete data chunk header exists.
func splitAtData(b []byte) (header, data []byte, found bool) {
	for off := 12; off+8 <= len(b); {
		id := string(b[off : off+4])
		size := int(binary.LittleEndian.Uint32(b[off+4 : off+8]))

		if id == chunkData {
			end := off + 8 + size
			if size == 0 || size == int(^uint32(0)) || end > len(b) || end < off {
				end = len(b)
			}

			return b[:off+8], b[off+8 : end], true
		}

		off += 8 + size + size&1
	}

	return b, nil, false
}

// LiveHeader returns a copy of a WAV file whose data chunk size is replaced with size, as served by live streaming
// servers. For the live-stream markers 0 and 0xFFFFFFFF the RIFF size gets the same value, as those servers write it
// (a definite RIFF size declaring chunks after a 0-size data chunk would make that data empty). Input without a
// complete data chunk header is returned unchanged.
func LiveHeader(b []byte, size uint32) []byte {
	out := bytes.Clone(b)

	if header, _, found := splitAtData(out); found {
		setLiveSizes(out[:len(header)], size)
	}

	return out
}

// setLiveSizes writes size into the data chunk size field at the end of header, and into the RIFF size field too when
// size is a live-stream marker.
func setLiveSizes(header []byte, size uint32) {
	binary.LittleEndian.PutUint32(header[len(header)-4:], size)

	if size == 0 || size == math.MaxUint32 {
		binary.LittleEndian.PutUint32(header[4:8], size)
	}
}
