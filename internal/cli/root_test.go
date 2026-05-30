package cli_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/buildoutinc/injector/internal/cli"
)

func runCLI(t *testing.T, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	var so, se bytes.Buffer
	code = cli.Execute(context.Background(), args, &so, &se, "test")
	return so.String(), se.String(), code
}

func TestRootHelp_NoArgs(t *testing.T) {
	stdout, stderr, code := runCLI(t)
	if code != 0 {
		t.Fatalf("exit=%d stderr=%q", code, stderr)
	}
	for _, want := range []string{"inject", "project", "Commands", "Flags"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout missing %q\n--- stdout ---\n%s", want, stdout)
		}
	}
}

func TestRootHelp_HelpFlag(t *testing.T) {
	for _, flag := range []string{"--help", "-h"} {
		t.Run(flag, func(t *testing.T) {
			stdout, stderr, code := runCLI(t, flag)
			if code != 0 {
				t.Fatalf("exit=%d stderr=%q", code, stderr)
			}
			if !strings.Contains(stdout, "project") {
				t.Errorf("stdout missing project listing:\n%s", stdout)
			}
		})
	}
}

func TestRootUnknownSubcommand(t *testing.T) {
	stdout, stderr, code := runCLI(t, "does-not-exist")
	if code == 0 {
		t.Fatalf("expected non-zero exit, got 0 (stdout=%q stderr=%q)", stdout, stderr)
	}
	if stderr == "" {
		t.Errorf("expected error on stderr, got empty")
	}
	if !strings.Contains(stderr, "does-not-exist") {
		t.Errorf("expected stderr to mention the unknown command, got: %q", stderr)
	}
}
