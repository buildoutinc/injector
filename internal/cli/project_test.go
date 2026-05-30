package cli_test

import (
	"strings"
	"testing"
)

func TestProjectHelp(t *testing.T) {
	stdout, stderr, code := runCLI(t, "project")
	if code != 0 {
		t.Fatalf("exit=%d stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "init") {
		t.Errorf("project help should list 'init' subcommand, got:\n%s", stdout)
	}
}

func TestProjectHelpFlag(t *testing.T) {
	stdout, stderr, code := runCLI(t, "project", "--help")
	if code != 0 {
		t.Fatalf("exit=%d stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "init") {
		t.Errorf("project --help should list 'init' subcommand, got:\n%s", stdout)
	}
}
