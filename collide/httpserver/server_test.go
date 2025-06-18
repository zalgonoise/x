package httpserver

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestServer_RegisterHTTP_Serve(t *testing.T) {
	s, err := NewServer(":0")
	require.NoError(t, err)

	testListener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	err = s.RegisterHTTP(
		http.MethodGet,
		"/test",
		http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			_, err := writer.Write([]byte("Hello World"))
			require.NoError(t, err)
		}),
	)
	require.NoError(t, err)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		err := s.server.Serve(testListener)
		require.ErrorIs(t, err, http.ErrServerClosed)
	}()

	client := http.Client{}

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		fmt.Sprintf("http://%s/test", testListener.Addr().String()),
		http.NoBody,
	)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "Hello World", string(respBody))

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = s.Shutdown(ctx)
	require.NoError(t, err)

	wg.Wait()
}
