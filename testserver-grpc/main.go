package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/teraptra/base/prod/prodserver"
	pb "github.com/teraptra/base/testserver-grpc/proto"
	"github.com/teraptra/base/testserver-grpc/server"
)

func main() {
	if err := run(); err != nil {
		slog.Error("Error: %v", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()
	homedir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	var kubeconfig string
	flag.StringVar(&kubeconfig, "kubeconfig", filepath.Join(homedir, ".kube", "config"), "path to the kubeconfig file")
	flag.Parse()
	gs, err := server.New(kubeconfig)
	if err != nil {
		return err
	}

	s, err := prodserver.New()
	if err != nil {
		return err
	}

	pb.RegisterDepManServiceServer(s, gs)

	return s.ListenAndServe(ctx)
}
