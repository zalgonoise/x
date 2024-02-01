package grpcserver

import (
	"net"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/zalgonoise/x/authz/pb/authz/v1"
)

type Server struct {
	server        *grpc.Server
	serverMetrics *grpc_prometheus.ServerMetrics
}

//go:generate mockery --name=Metrics --with-expecter
type Metrics interface {
	RegisterCollector(collector prometheus.Collector)
}

func NewServer(
	metrics Metrics,
	unaryInterceptors []grpc.UnaryServerInterceptor,
	streamInterceptors []grpc.StreamServerInterceptor,
) *Server {
	serverMetrics := grpc_prometheus.NewServerMetrics(grpc_prometheus.WithServerHandlingTimeHistogram())

	s := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler(
			otelgrpc.WithMessageEvents(otelgrpc.SentEvents, otelgrpc.ReceivedEvents),
		)),
		grpc.ChainUnaryInterceptor(
			append([]grpc.UnaryServerInterceptor{serverMetrics.UnaryServerInterceptor()}, unaryInterceptors...)...,
		),
		grpc.ChainStreamInterceptor(
			append([]grpc.StreamServerInterceptor{serverMetrics.StreamServerInterceptor()}, streamInterceptors...)...,
		),
	)

	reflection.Register(s)
	metrics.RegisterCollector(serverMetrics)

	return &Server{
		server:        s,
		serverMetrics: serverMetrics,
	}
}

func (s *Server) Serve(l net.Listener) error {
	s.serverMetrics.InitializeMetrics(s.server)

	return s.server.Serve(l)
}

func (s *Server) Shutdown() {
	s.server.GracefulStop()
}

func (s *Server) RegisterAuthzServer(backend pb.AuthzServer) {
	pb.RegisterAuthzServer(s.server, backend)
}

func (s *Server) RegisterCertificateAuthorityServer(backend pb.CertificateAuthorityServer) {
	pb.RegisterCertificateAuthorityServer(s.server, backend)
}
