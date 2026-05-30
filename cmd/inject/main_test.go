package main_test

import (
	"bytes"
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func buildBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "inject")
	cmd := exec.Command("go", "build", "-o", bin, "github.com/buildoutinc/injector/cmd/inject")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, stderr.String())
	}
	return bin
}

func TestBinarySmoke_HelpAndInit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping binary smoke test in -short mode")
	}
	bin := buildBinary(t)

	t.Run("no-args prints help and exits 0", func(t *testing.T) {
		var so, se bytes.Buffer
		cmd := exec.Command(bin)
		cmd.Stdout, cmd.Stderr = &so, &se
		if err := cmd.Run(); err != nil {
			t.Fatalf("exit=%v stderr=%q", err, se.String())
		}
		if !strings.Contains(so.String(), "project") {
			t.Errorf("help should list project subcommand:\n%s", so.String())
		}
	})

	t.Run("project init prints TODO: init!", func(t *testing.T) {
		var so, se bytes.Buffer
		cmd := exec.Command(bin, "project", "init")
		cmd.Stdout, cmd.Stderr = &so, &se
		if err := cmd.Run(); err != nil {
			t.Fatalf("exit=%v stderr=%q", err, se.String())
		}
		if so.String() != "TODO: init!\n" {
			t.Errorf("stdout = %q; want %q", so.String(), "TODO: init!\n")
		}
		if se.String() != "" {
			t.Errorf("stderr should be empty, got %q", se.String())
		}
	})

	t.Run("unknown command exits non-zero with stderr", func(t *testing.T) {
		var so, se bytes.Buffer
		cmd := exec.Command(bin, "does-not-exist")
		cmd.Stdout, cmd.Stderr = &so, &se
		err := cmd.Run()
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) || exitErr.ExitCode() == 0 {
			t.Fatalf("expected non-zero exit, got err=%v", err)
		}
		if se.String() == "" {
			t.Errorf("expected error on stderr, got empty")
		}
	})
}
