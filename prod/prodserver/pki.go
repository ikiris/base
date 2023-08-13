package prodserver

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"os"

	"google.golang.org/grpc/credentials"
)

var (
	certFile = flag.String("certfile", "/etc/mTLS/cert.pem", "certificate PEM file")
	keyFile  = flag.String("keyfile", "/etc/mTLS/key.pem", "key PEM file")
	caFile   = flag.String("cafile", "/etc/mTLS/ca.crt", "ca file")
)

// loadKeyPair does all the work for setting up the CA pool chains for mTLS verification from the given inputs.
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
