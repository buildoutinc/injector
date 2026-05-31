package main

import (
	"context"
	"io"
	"os"

	"github.com/buildoutinc/injector/internal/cli"
)

// Overridden at link time by GoReleaser via -ldflags -X main.version=…
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	cli.SetBuildInfo(version, commit, date)
	ctx, stop := cli.NotifyContext(context.Background())
	defer stop()
	return cli.Execute(ctx, args, stdout, stderr, version)
}
