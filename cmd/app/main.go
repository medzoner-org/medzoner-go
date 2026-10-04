package main

import (
	"context"

	wire "github.com/Medzoner/medzoner-go/internal/wire"
	"github.com/medzoner-org/gomedz/pkg/logger"
)

func main() {
	ctx := context.Background()
	srv, err := wire.InitServer(ctx)
	if err != nil {
		logger.Fatal(ctx, "Failed to initialize server", err)
		return
	}

	if err := srv.Serve(ctx); err != nil {
		logger.Fatal(ctx, "Server encountered an error", err)
	}
}
