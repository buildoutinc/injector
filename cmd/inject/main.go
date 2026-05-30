package main

import (
	"context"
	"io"
	"os"

	"github.com/buildoutinc/injector/internal/cli"
)

var version = "dev"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	ctx, stop := cli.NotifyContext(context.Background())
	defer stop()
	return cli.Execute(ctx, args, stdout, stderr, version)
}
