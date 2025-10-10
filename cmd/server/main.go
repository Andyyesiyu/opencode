package main

import (
	"context"
	"log"

	"github.com/opencodehq/opencode/pkg/server"
)

func main() {
	srv := server.New()
	ctx := context.Background()
	if err := srv.Start(ctx, ":8080"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
