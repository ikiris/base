package prodserver

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// I would expect this to be defaulted using opt mods on New so it can be individualized for multiple active runners, but this is the quick and dirty to save time for std defaults.
var (
	port     = flag.Int("port", 6000, "listen port")
	proto    = flag.String("proto", "tcp", "listen protocol")
)

type prodServer struct {
	gs      *grpc.Server
	cf      context.CancelFunc
	healthz *health.Server
}

// RegisterService is used to register a grpc service on the prodserver.
func (s *prodServer) RegisterService(sd *grpc.ServiceDesc, ss interface{}) {
	s.gs.RegisterService(sd, ss)
}

// New returns a prodServer for use.
func New() (*prodServer, error) {
	creds, err := loadKeyPair()
	if err != nil {
		return nil, err
	}
	s := grpc.NewServer(
		grpc.Creds(creds),
		grpc.UnaryInterceptor(middlefunc),
	)
	healthz := health.NewServer()
	healthpb.RegisterHealthServer(s, healthz)
	return &prodServer{
		gs:      s,
		healthz: healthz,
	}, nil
}

// ListenAndServe sets up a standard server for X service on the configured port for the default servemux.
func (s *prodServer) ListenAndServe(ctx context.Context) error {
	slog.Info("prod server init on :%d", *port)
	l, err := net.Listen(*proto, fmt.Sprintf(":%d", *port))
	if err != nil {
		return err
	}

	sctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	s.cf = stop

	go func() {
		<-sctx.Done()
		slog.Info("prod stop requested")
		s.healthz.Shutdown()
		s.gs.GracefulStop()
	}()

	if err := s.gs.Serve(l); err != nil {
		return fmt.Errorf("prod server listen error: %w", err)
	}
	slog.Info("prod server bye")
	return nil
}

// GracefulStop sets health checks to non serving, haults inbound acceptance, and instructs the server to shutdown once existing serving complete.
func (s *prodServer) GracefulStop() {
	s.cf()
}

// SetServingStatus updates the healthcheck table.
func (s *prodServer) SetServingStatus(service string, status healthpb.HealthCheckResponse_ServingStatus) {
	// This should be smarter to be safe in terms of service name to prevent overlap, but this is a single service example.
	s.healthz.SetServingStatus(fmt.Sprintf("grpc.health.v1.%s", service), status)
}
