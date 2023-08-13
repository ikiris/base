package server

import (
	"context"
	"testing"
)

func TestServer(t *testing.T) {
	ctx := context.Background()

	server := New("something")

	_ := map[string]struct {
		name string
	}
}
