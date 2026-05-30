package cli_test

import (
	"strings"
	"testing"
)

func TestProjectInit_Output(t *testing.T) {
	stdout, stderr, code := runCLI(t, "project", "init")
	if code != 0 {
		t.Fatalf("exit=%d stderr=%q", code, stderr)
	}
	if stdout != "TODO: init!\n" {
		t.Errorf("stdout = %q; want %q", stdout, "TODO: init!\n")
	}
	if stderr != "" {
		t.Errorf("stderr should be empty, got %q", stderr)
	}
}

func TestProjectInitHelpFlag(t *testing.T) {
	stdout, stderr, code := runCLI(t, "project", "init", "--help")
	if code != 0 {
		t.Fatalf("exit=%d stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "init") {
		t.Errorf("init help should mention 'init', got:\n%s", stdout)
	}
}
