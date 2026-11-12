package sources_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/zalgonoise/x/audio/sdk/sources"
)

func TestFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "a.wav")
	require.NoError(t, os.WriteFile(path, []byte("RIFF"), 0o600))

	src := sources.NewFile(path)

	for range 2 { // every Open reads the file from the start
		rc, err := src.Open(t.Context())
		require.NoError(t, err)

		b, err := io.ReadAll(rc)
		require.NoError(t, err)
		require.Equal(t, []byte("RIFF"), b)
		require.NoError(t, rc.Close())
	}

	_, err := sources.NewFile(filepath.Join(t.TempDir(), "missing")).Open(t.Context())
	require.ErrorIs(t, err, os.ErrNotExist)
}

type closer struct {
	io.Reader

	closed bool
}

func (c *closer) Close() error {
	c.closed = true

	return nil
}

func TestReader(t *testing.T) {
	src := sources.NewReader(bytes.NewReader([]byte("abc")))

	rc, err := src.Open(t.Context())
	require.NoError(t, err)

	b, err := io.ReadAll(rc)
	require.NoError(t, err)
	require.Equal(t, []byte("abc"), b)
	require.NoError(t, rc.Close())

	_, err = src.Open(t.Context())
	require.ErrorIs(t, err, sources.ErrConsumed)

	c := &closer{Reader: bytes.NewReader(nil)}
	rc, err = sources.NewReader(c).Open(t.Context())
	require.NoError(t, err)
	require.NoError(t, rc.Close())
	require.True(t, c.closed, "an io.ReadCloser is closed, which unblocks a pending read")
}

func TestHTTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/missing":
			http.NotFound(w, r)
		case "/header":
			if _, err := io.WriteString(w, r.Header.Get("X-Test")+" "+r.Header.Get("Accept")); err != nil {
				t.Error(err)
			}
		default:
			if _, err := io.WriteString(w, "audio"); err != nil {
				t.Error(err)
			}
		}
	}))
	defer srv.Close()

	src, err := sources.NewHTTP(srv.URL + "/ok")
	require.NoError(t, err)

	rc, err := src.Open(t.Context())
	require.NoError(t, err)

	b, err := io.ReadAll(rc)
	require.NoError(t, err)
	require.Equal(t, "audio", string(b))
	require.NoError(t, rc.Close())

	src, err = sources.NewHTTP(srv.URL+"/header", sources.WithHeader("X-Test", "yes"), sources.WithClient(srv.Client()))
	require.NoError(t, err)

	rc, err = src.Open(t.Context())
	require.NoError(t, err)

	b, err = io.ReadAll(rc)
	require.NoError(t, err)
	require.Equal(t, "yes audio/*, */*;q=0.5", string(b))
	require.NoError(t, rc.Close())

	src, err = sources.NewHTTP(srv.URL + "/missing")
	require.NoError(t, err)

	_, err = src.Open(t.Context())
	require.ErrorIs(t, err, sources.ErrUnexpectedStatus)

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err = src.Open(ctx)
	require.ErrorIs(t, err, context.Canceled)

	for _, bad := range []string{"", "ftp://host/a.wav", "/relative", "http://", "::"} {
		_, err := sources.NewHTTP(bad)
		require.True(t, errors.Is(err, sources.ErrInvalidURL), "%q: %v", bad, err)
	}
}

// blockingReader blocks every Read until release is closed, like stdin on a terminal: closing it is not possible.
type blockingReader struct{ release chan struct{} }

func (b blockingReader) Read([]byte) (int, error) {
	<-b.release

	return 0, io.EOF
}

// Review round 3: closing a Reader source's stream unblocks a pending Read even when the source cannot be interrupted.
func TestReaderCloseUnblocks(t *testing.T) {
	src := blockingReader{release: make(chan struct{})}

	rc, err := sources.NewReader(src).Open(t.Context())
	require.NoError(t, err)

	read := make(chan error, 1)

	go func() {
		_, err := rc.Read(make([]byte, 8))
		read <- err
	}()

	require.NoError(t, rc.Close())

	select {
	case err := <-read:
		require.ErrorIs(t, err, io.ErrClosedPipe)
	case <-time.After(5 * time.Second):
		t.Fatal("Read still blocked after Close")
	}

	close(src.release) // the copying goroutine ends when its read returns
}

// Review round 3: the default HTTP client keeps no idle connections behind.
func TestHTTPLeavesNoConnections(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, err := io.WriteString(w, "audio"); err != nil {
			t.Error(err)
		}
	}))
	defer srv.Close()

	src, err := sources.NewHTTP(srv.URL)
	require.NoError(t, err)

	for range 3 {
		rc, err := src.Open(t.Context())
		require.NoError(t, err)

		_, err = io.ReadAll(rc)
		require.NoError(t, err)
		require.NoError(t, rc.Close())
	}
}
