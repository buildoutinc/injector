package cli_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/buildoutinc/injector/internal/cli"
	"github.com/buildoutinc/injector/internal/updatecheck"
	"github.com/buildoutinc/injector/internal/updater"
)

func TestRenderNotice_NoTTY(t *testing.T) {
	var buf bytes.Buffer
	cli.RenderNotice(&buf, &updatecheck.Notice{Current: "v0.1.0", Latest: "v0.3.0"}, false)
	want := "\n==> A newer version is available: v0.3.0\n==> Run \"inject upgrade\" to install it.\n"
	if buf.String() != want {
		t.Fatalf("got %q\nwant %q", buf.String(), want)
	}
}

func TestRenderNotice_NilSilent(t *testing.T) {
	var buf bytes.Buffer
	cli.RenderNotice(&buf, nil, false)
	if buf.Len() != 0 {
		t.Fatalf("expected empty, got %q", buf.String())
	}
}

func TestNoticeAppendedToHelp(t *testing.T) {
	ck := &fakeChecker{notice: &updatecheck.Notice{Current: "v0.1.0", Latest: "v0.3.0"}}
	cli.SetBuildInfo("v0.1.0", "abc1234", "2026-05-30")
	t.Cleanup(func() { cli.SetBuildInfo("dev", "none", "unknown") })

	var so, se bytes.Buffer
	code := cli.ExecuteWith(
		context.Background(),
		[]string{"--help"},
		&so, &se,
		"v0.1.0",
		cli.Options{Checker: ck, Updater: nilUpdater{}},
	)
	if code != 0 {
		t.Fatalf("exit=%d stderr=%q", code, se.String())
	}
	if !strings.Contains(so.String(), "Commands:") {
		t.Errorf("missing help screen: %q", so.String())
	}
	if !strings.Contains(so.String(), "==> A newer version is available: v0.3.0") {
		t.Errorf("missing notice after help: %q", so.String())
	}
}

// nilUpdater is an Updater that's never called (because Checker is fake).
type nilUpdater struct{}

func (nilUpdater) Latest(context.Context, updater.LatestOpts) (updater.Release, error) {
	return updater.Release{}, nil
}
func (nilUpdater) Apply(context.Context, updater.Release, string) (updater.Result, error) {
	return updater.Result{}, nil
}
