package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/teraptra/base/prod/prodserver"
	pb "github.com/teraptra/base/testserver-grpc/proto"
	"github.com/teraptra/base/testserver-grpc/server"
	"google.golang.org/grpc/health/grpc_health_v1"
)

const defaultCheckTimeout = 30 * time.Second

func main() {
	if err := run(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

// run is a simple func wrapper for the basic startup seq
func run() error {
	ctx := context.Background()
	homedir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	var kubeconfig string
	var pollFreq time.Duration
	flag.StringVar(&kubeconfig, "kubeconfig", filepath.Join(homedir, ".kube", "config"), "path to the kubeconfig file")
	flag.DurationVar(&pollFreq, "pollFreq", 30*time.Second, "Poll inteval for backend health check")
	flag.Parse()
	s, err := server.New(kubeconfig)
	if err != nil {
		return err
	}

	ps, err := prodserver.New()
	if err != nil {
		return err
	}

	pb.RegisterDepManServiceServer(ps, s)

	// health watch poller
	updateHealthz := func(ctx context.Context) {
		tctx, cf := context.WithTimeout(ctx, defaultCheckTimeout)
		defer cf()
		stat := grpc_health_v1.HealthCheckResponse_SERVING
		if err := s.Check(tctx); err != nil {
			stat = grpc_health_v1.HealthCheckResponse_NOT_SERVING
		}
		ps.SetServingStatus(pb.DepManService_ServiceDesc.ServiceName, stat)
	}
	go func() {
		tik := time.NewTicker(pollFreq)
		defer tik.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-tik.C:
				updateHealthz(ctx)
			}
		}
	}()

	return ps.ListenAndServe(ctx)
}
