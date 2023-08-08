package prodserver

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/exp/slog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

var (
	certFile = flag.String("certfile", "cert.pem", "certificate PEM file")
	keyFile  = flag.String("keyfile", "key.pem", "key PEM file")
	port     = flag.Int("port", 6000, "listen port")
	proto    = flag.String("proto", "tcp", "listen protocol")
)

func init() {
	flag.Parse()
}

type prodServer struct {
	gs *grpc.Server
	cf context.CancelFunc
}

func New() (*prodServer, error) {
	creds, err := loadKeyPair()
	if err != nil {
		return nil, err
	}
	gs := grpc.NewServer(
		grpc.Creds(creds),
		grpc.UnaryInterceptor(middlefunc),
	)
	return &prodServer{
		gs: gs,
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

func loadKeyPair() (credentials.TransportCredentials, error) {
	certificate, err := tls.LoadX509KeyPair("certs/server.crt", "certs/server.key")
	if err != nil {
		return nil, fmt.Errorf("failed to load server certification: %w", err)
	}

	data, err := os.ReadFile("certs/ca.crt")
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
