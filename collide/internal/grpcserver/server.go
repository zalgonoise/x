package grpcserver

import (
	"context"
	"github.com/zalgonoise/x/collide/internal/metrics"
	"net"
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/zalgonoise/x/collide/pkg/api/pb/collide/v1"
)

type Server struct {
	server        *grpc.Server
	serverMetrics *grpc_prometheus.ServerMetrics
}

type Metrics interface {
	IncListDistricts(ctx context.Context)
	IncListDistrictsFailed(ctx context.Context)
	ObserveListDistrictsLatency(ctx context.Context, duration time.Duration)
	IncListAllTracksByDistrict(ctx context.Context, district string)
	IncListAllTracksByDistrictFailed(ctx context.Context, district string)
	ObserveListAllTracksByDistrictLatency(ctx context.Context, duration time.Duration, district string)
	IncListDriftTracksByDistrict(ctx context.Context, district string)
	IncListDriftTracksByDistrictFailed(ctx context.Context, district string)
	ObserveListDriftTracksByDistrictLatency(ctx context.Context, duration time.Duration, district string)
	IncGetAlternativesByDistrictAndTrack(ctx context.Context, district string, track string)
	IncGetAlternativesByDistrictAndTrackFailed(ctx context.Context, district string, track string)
	ObserveGetAlternativesByDistrictAndTrackLatency(ctx context.Context, duration time.Duration, district string, track string)
	IncGetCollisionsByDistrictAndTrack(ctx context.Context, district string, track string)
	IncGetCollisionsByDistrictAndTrackFailed(ctx context.Context, district string, track string)
	ObserveGetCollisionsByDistrictAndTrackLatency(ctx context.Context, duration time.Duration, district string, track string)
}

func NewServer(
	unaryInterceptors []grpc.UnaryServerInterceptor,
	streamInterceptors []grpc.StreamServerInterceptor,
	m Metrics,
) *Server {
	var promMetrics *grpc_prometheus.ServerMetrics

	if prom, ok := m.(*metrics.Prometheus); ok {
		promMetrics = grpc_prometheus.NewServerMetrics(grpc_prometheus.WithServerHandlingTimeHistogram())
		unaryInterceptors = append(
			[]grpc.UnaryServerInterceptor{promMetrics.UnaryServerInterceptor()}, unaryInterceptors...)
		streamInterceptors = append(
			[]grpc.StreamServerInterceptor{promMetrics.StreamServerInterceptor()}, streamInterceptors...)
		metrics.RegisterCollector(prom, promMetrics)
	}

	s := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler(
			otelgrpc.WithMessageEvents(otelgrpc.SentEvents, otelgrpc.ReceivedEvents),
		)),
		grpc.ChainUnaryInterceptor(unaryInterceptors...), grpc.ChainStreamInterceptor(streamInterceptors...),
	)

	reflection.Register(s)

	return &Server{
		server:        s,
		serverMetrics: promMetrics,
	}
}

func (s *Server) Serve(l net.Listener) error {
	if s.serverMetrics != nil {
		s.serverMetrics.InitializeMetrics(s.server)
	}

	return s.server.Serve(l)
}

func (s *Server) Shutdown(ctx context.Context) {
	shutdownSuccess := make(chan struct{})

	go func() {
		s.server.GracefulStop()
		close(shutdownSuccess)
	}()

	select {
	case <-ctx.Done():
		s.server.Stop()
	case <-shutdownSuccess:
	}
}

func (s *Server) RegisterCollideServer(backend pb.CollideServiceServer) {
	pb.RegisterCollideServiceServer(s.server, backend)
}
