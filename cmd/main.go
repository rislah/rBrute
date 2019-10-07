package main

import (
	"context"

	rbrute "github.com/rislah/rBrute"
	"github.com/rislah/rBrute/config"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	cfg := config.NewConfig("/home/rsl/go/src/github.com/rislah/rBrute/config.yaml")
	rb := rbrute.NewRBrute(ctx, cfg)
	rb.Start("/home/rsl/go/src/github.com/rislah/rBrute/out.txt", "/home/rsl/go/src/github.com/rislah/rBrute/creds.txt")
	cancel()
}
