package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/teraptra/base/prod/prodserver"
	"github.com/teraptra/base/testserver-http/server"
)

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		slog.Error("Error: %v", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	var kubeconfig string
	flag.StringVar(&kubeconfig, "kubeconfig", filepath.Join(homedir, ".kube", "config"), "path to the kubeconfig file")
	flag.Parse()

	pSrv, err := prodserver.New()
	if err != nil {
		return err
	}

	s, err := server.New(*&kubeconfig)
	if err != nil {
		return err
	}

	s.ProdRegister(pSrv)
	return pSrv.ListenAndServe(ctx)
}
