package httpserver

import (
	"context"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"

	pb "github.com/zalgonoise/x/collide/pkg/api/pb/collide/v1"
)

type Server struct {
	server http.Server
	mux    *runtime.ServeMux
}

const (
	defaultReadTimeout  = 2500 * time.Millisecond
	defaultWriteTimeout = 5000 * time.Millisecond
)

func NewServer(addr string) (*Server, error) {
	mux := runtime.NewServeMux(
		runtime.WithMetadata(func(ctx context.Context, _ *http.Request) metadata.MD {
			md := metadata.MD{}

			otelgrpc.Inject(ctx, &md)

			return md
		}),
	)

	err := mux.HandlePath(http.MethodGet, "/ready", ready)
	if err != nil {
		return nil, err
	}

	tracingMiddleware := otelhttp.NewMiddleware("grpc-gateway",
		otelhttp.WithFilter(func(request *http.Request) bool {
			return request.URL.Path != "/ready" && request.URL.Path != "/metrics"
		}),
	)

	return &Server{
		server: http.Server{
			Handler:      tracingMiddleware(urlAttributesMiddleware(mux)),
			Addr:         addr,
			ReadTimeout:  defaultReadTimeout,
			WriteTimeout: defaultWriteTimeout,
		},
		mux: mux,
	}, nil
}

func (s *Server) RegisterCollideService(ctx context.Context, client pb.CollideServiceClient) error {
	return pb.RegisterCollideServiceHandlerClient(ctx, s.mux, client)
}

func (s *Server) RegisterHTTP(method, path string, handler http.Handler) error {
	return s.mux.HandlePath(method, path, func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		handler.ServeHTTP(w, r)
	})
}

func (s *Server) ListenAndServe() error {
	s.server.Handler = corsMiddleware(s.server.Handler)

	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func ready(w http.ResponseWriter, _ *http.Request, _ map[string]string) {
	w.WriteHeader(http.StatusOK)
}

func urlAttributesMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		span := trace.SpanFromContext(request.Context())
		span.SetAttributes(semconv.URLPath(request.URL.Path))

		if request.URL.RawQuery != "" {
			span.SetAttributes(semconv.URLQuery(request.URL.RawQuery))
		}

		h.ServeHTTP(writer, request)
	})
}

// corsMiddleware adds the necessary CORS headers to the response.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: limit to the actual target domain
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// allow common methods
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		// allow common headers
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// browsers sometimes send a pre-flight OPTIONS request to check
		// if the server allows the actual request.
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// pass the request to the next handler in the chain
		next.ServeHTTP(w, r)
	})
}
