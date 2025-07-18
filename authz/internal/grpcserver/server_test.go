package grpcserver

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/zalgonoise/x/authz/internal/grpcserver/mocks"
)

func TestServer_Serve(t *testing.T) {
	s := Server{server: grpc.NewServer()}

	testListener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		err := s.Serve(testListener)
		require.NoError(t, err)
	}()

	conn, err := grpc.Dial(testListener.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	for state := conn.GetState(); state != connectivity.Ready; state = conn.GetState() {
		require.True(t, conn.WaitForStateChange(ctx, state), "timeout waiting for conn to get ready")
	}

	s.Shutdown(context.Background())
	wg.Wait()
}

func TestNewServer(t *testing.T) {
	metrics := mocks.NewMetrics(t)

	metrics.EXPECT().RegisterCollector(mock.Anything).Return()

	s := NewServer(metrics, nil, nil)
	require.NotNil(t, s.server)
}
