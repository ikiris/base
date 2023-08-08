package server

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"golang.org/x/exp/slog"
)

var (
	certFile = flag.String("certfile", "cert.pem", "certificate PEM file")
	keyFile  = flag.String("keyfile", "key.pem", "key PEM file")
)

func init() {
	flag.Parse()
}

type config struct {
	port int
	certFile, keyFile *string
	clientAuthType tls.ClientAuthType
}

type sOpt func(*config)

func Port(p int) sOpt {
	return func(c *config) {
		c.port = p
	}
}

type server struct {
	conf              *config
	srv               *http.Server
	ctx *context.Context
}

// New returns a prod server.
func New(opts ...sOpt) (*server, error) {
	var (
		defaultPort = 8000
		defaultClientAuth = tls.VerifyClientCertIfGiven
	)
	if port, err := strconv.Atoi(os.Getenv("SERVER_PORT")); err == nil && port != 0 {
		defaultPort = port
	}
	conf := &config{
		port: defaultPort,
		certFile: certFile,
		keyFile: keyFile,
		clientAuthType: defaultClientAuth,
	}
	for _, opt := range opts {
		opt(conf)
	}
	return buildServer(conf)
}

func buildServer(c *config) (*server, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" {
			http.NotFound(w, req)
			return
		}
		fmt.Fprintf(w, "Something\n")
	})

	s := &server{
		srv: &http.Server{
			Handler:   mux,
			TLSConfig: &tls.Config{
				ClientAuth: tls.RequireAndVerifyClientCert,
			},
		},
		conf: c,
	}

	http.HandleFunc("/statusz", s.statusZHandler)

	return s, nil
}

func (s *server) statusZHandler(w http.ResponseWriter, r *http.Request) {

}

func (s *server) ListenAndServe(ctx context.Context) error {
	c := s.conf
	slog.Info("Prod server init on :%d", c.port)
	if err := s.srv.ListenAndServeTLS(*c.certFile, *c.keyFile); err != http.ErrServerClosed {
		return fmt.Errorf("prod server listen error: %w", err)
	}
	return nil
}

// RegisterHealthZHandler sets up a client function to serve as health reporting. Normally this would be smarter, but corner cut for speed.
func (s *server) RegisterHealthZHandler(h http.Handler) {
	http.Handle("/healthz", h)
}

func (s *server) Shutdown(ctx context.Context) error {
	slog.Info("Prod server shutdown initiated.")
	return s.srv.Shutdown(ctx)
}