package prodserver

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/peer"
)

// I would expect this to be defaulted using opt mods on New so it can be individualized for multiple active runners, but this is the quick and dirty to save time for std defaults.
var (
	certFile = flag.String("certfile", "/etc/mTLS/cert.pem", "certificate PEM file")
	keyFile  = flag.String("keyfile", "/etc/mTLS/key.pem", "key PEM file")
	caFile   = flag.String("cafile", "/etc/mTLS/ca.crt", "key PEM file")
	port     = flag.Int("port", 6000, "listen port")
	proto    = flag.String("proto", "tcp", "listen protocol")
)

type prodServer struct {
	gs      *grpc.Server
	cf      context.CancelFunc
	healthz *health.Server
}

func (s *prodServer) RegisterService(sd *grpc.ServiceDesc, ss interface{}) {
	s.gs.RegisterService(sd, ss)
}

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

func (s *prodServer) GracefulStop() {
	s.cf()
}

func (s *prodServer) SetServingStatus(service string, status healthpb.HealthCheckResponse_ServingStatus) {
	// This should be smarter to be safe in terms of service name to prevent overlap, but this is a single service example.
	s.healthz.SetServingStatus(fmt.Sprintf("grpc.health.v1.%s", service), status)
}

func loadKeyPair() (credentials.TransportCredentials, error) {
	certificate, err := tls.LoadX509KeyPair(*certFile, *keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load server certification: %w", err)
	}

	data, err := os.ReadFile(*caFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load CA file: %w", err)
	}

	capool := x509.NewCertPool()
	if !capool.AppendCertsFromPEM(data) {
		return nil, fmt.Errorf("failed to create ca pool")
	}

	tlsConfig := &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    capool,
	}
	return credentials.NewTLS(tlsConfig), nil
}

// auth stuff goes here. Mostly so the healthchecks on mux can work, also any user auth/z for service.
func middlefunc(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	// get client tls info
	if p, ok := peer.FromContext(ctx); ok {
		if mtls, ok := p.AuthInfo.(credentials.TLSInfo); ok {
			for _, item := range mtls.State.PeerCertificates {
				slog.Debug("request certificate subject:", item.Subject)
			}
		}
	}
	return handler(ctx, req)
}
