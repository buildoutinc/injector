package cli_test

import (
	"context"
	"syscall"
	"testing"
	"time"

	"github.com/buildoutinc/injector/internal/cli"
)

func TestNotifyContext_CancelsOnSignal(t *testing.T) {
	ctx, stop := cli.NotifyContext(context.Background())
	defer stop()

	if err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM); err != nil {
		t.Fatalf("kill: %v", err)
	}

	select {
	case <-ctx.Done():
		// ok
	case <-time.After(2 * time.Second):
		t.Fatal("context not cancelled within 2s after SIGTERM")
	}
}
