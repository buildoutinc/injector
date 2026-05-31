package cli_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/buildoutinc/injector/internal/cli"
	"github.com/buildoutinc/injector/internal/updatecheck"
)

type fakeChecker struct {
	startDelay time.Duration
	notice     *updatecheck.Notice
}

func (f *fakeChecker) Start(ctx context.Context) {
	if f.startDelay > 0 {
		select {
		case <-time.After(f.startDelay):
		case <-ctx.Done():
		}
	}
}
func (f *fakeChecker) Notice() *updatecheck.Notice { return f.notice }

func runVersion(t *testing.T, ck updatecheck.Checker, version string) (string, string, int) {
	t.Helper()
	cli.SetBuildInfo(version, "abc1234", "2026-05-30")
	t.Cleanup(func() { cli.SetBuildInfo("dev", "none", "unknown") })
	var so, se bytes.Buffer
	code := cli.ExecuteWith(
		context.Background(),
		[]string{"version"},
		&so, &se,
		version,
		cli.Options{Checker: ck, Updater: nilUpdater{}},
	)
	return so.String(), se.String(), code
}

func TestVersionCommand_NoticeRendered(t *testing.T) {
	ck := &fakeChecker{notice: &updatecheck.Notice{Current: "v0.1.0", Latest: "v0.3.0"}}
	stdout, _, code := runVersion(t, ck, "v0.1.0")
	if code != 0 {
		t.Fatal("expected 0 exit")
	}
	if !strings.Contains(stdout, "inject v0.1.0") {
		t.Errorf("missing version line: %q", stdout)
	}
	if !strings.Contains(stdout, "==> A newer version is available: v0.3.0") {
		t.Errorf("missing notice: %q", stdout)
	}
}

func TestVersionCommand_NoNoticeWhenUpToDate(t *testing.T) {
	stdout, _, code := runVersion(t, &fakeChecker{}, "v0.1.0")
	if code != 0 {
		t.Fatal("expected 0 exit")
	}
	if strings.Contains(stdout, "==>") {
		t.Errorf("unexpected notice: %q", stdout)
	}
}
