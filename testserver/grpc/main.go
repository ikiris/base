package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/teraptra/base/prod/prodserver"
)

func main() {
	if err := run(); err != nil {
		slog.Error("Error: %v", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()
	s, err := prodserver.New()
	if err != nil {
		return err
	}
	gs := toolserver.New()

	greet.RegisterGreetingServer(s, new(gs))

	return s.ListenAndServe(ctx)
}
