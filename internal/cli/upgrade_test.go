package cli_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buildoutinc/injector/internal/cli"
	"github.com/buildoutinc/injector/internal/updatecheck"
	"github.com/buildoutinc/injector/internal/updater"
)

// runUpgrade drives ExecuteWith with a fake Updater and disabled Checker.
// Returns stdout, stderr, exit code, and the fake for call-count asserts.
func runUpgrade(t *testing.T, version string, fake *updater.FakeUpdater, args ...string) (string, string, int, *updater.FakeUpdater) {
	t.Helper()
	if fake == nil {
		fake = &updater.FakeUpdater{}
	}
	cli.SetBuildInfo(version, "abc1234", "2026-05-30")
	t.Cleanup(func() { cli.SetBuildInfo("dev", "none", "unknown") })

	var so, se bytes.Buffer
	code := cli.ExecuteWith(
		context.Background(),
		append([]string{"upgrade"}, args...),
		&so, &se,
		version,
		cli.Options{Updater: fake, Checker: updatecheck.Disabled{}},
	)
	return so.String(), se.String(), code, fake
}

func TestUpgrade_AlreadyCurrent(t *testing.T) {
	f := &updater.FakeUpdater{LatestFunc: func(context.Context, updater.LatestOpts) (updater.Release, error) {
		return updater.Release{Version: "v0.1.0"}, nil
	}}
	stdout, stderr, code, fake := runUpgrade(t, "v0.1.0", f)
	if code != 0 {
		t.Fatalf("exit=%d stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "already up to date") {
		t.Errorf("stdout = %q; want contains 'already up to date'", stdout)
	}
	if fake.ApplyCallCount() != 0 {
		t.Errorf("Apply called %d times; want 0", fake.ApplyCallCount())
	}
}

func TestUpgrade_NewerVersion(t *testing.T) {
	// Point binPath at a writable tempfile so the permission check passes.
	bin := writableBin(t)
	defer chdirRestore(t, filepath.Dir(bin))()

	f := &updater.FakeUpdater{
		LatestFunc: func(context.Context, updater.LatestOpts) (updater.Release, error) {
			return updater.Release{Version: "v0.2.0"}, nil
		},
		ApplyFunc: func(_ context.Context, r updater.Release, p string) (updater.Result, error) {
			return updater.Result{FromVersion: "v0.1.0", ToVersion: r.Version, Path: p}, nil
		},
	}
	stdout, stderr, code, fake := runUpgrade(t, "v0.1.0", f)
	if code != 0 {
		t.Fatalf("exit=%d stderr=%q stdout=%q", code, stderr, stdout)
	}
	if !strings.Contains(stdout, "upgraded v0.1.0 → v0.2.0") {
		t.Errorf("stdout = %q", stdout)
	}
	if fake.ApplyCallCount() != 1 {
		t.Errorf("Apply called %d times; want 1", fake.ApplyCallCount())
	}
}

func TestUpgrade_NetworkError(t *testing.T) {
	f := &updater.FakeUpdater{LatestFunc: func(context.Context, updater.LatestOpts) (updater.Release, error) {
		return updater.Release{}, errors.New("dial tcp: no route to host")
	}}
	stdout, stderr, code, fake := runUpgrade(t, "v0.1.0", f)
	if code == 0 {
		t.Fatalf("expected non-zero exit; stdout=%q", stdout)
	}
	if !strings.HasPrefix(stderr, "inject: cannot upgrade:") {
		t.Errorf("stderr = %q", stderr)
	}
	if fake.ApplyCallCount() != 0 {
		t.Errorf("Apply called %d times; want 0", fake.ApplyCallCount())
	}
}

func TestUpgrade_ChecksumMismatch(t *testing.T) {
	bin := writableBin(t)
	defer chdirRestore(t, filepath.Dir(bin))()

	f := &updater.FakeUpdater{
		LatestFunc: func(context.Context, updater.LatestOpts) (updater.Release, error) {
			return updater.Release{Version: "v0.2.0"}, nil
		},
		ApplyFunc: func(context.Context, updater.Release, string) (updater.Result, error) {
			return updater.Result{}, updater.ErrChecksumMismatch
		},
	}
	_, stderr, code, _ := runUpgrade(t, "v0.1.0", f)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stderr, "checksum mismatch") {
		t.Errorf("stderr = %q", stderr)
	}
}

func TestUpgrade_UnsupportedPlatform(t *testing.T) {
	bin := writableBin(t)
	defer chdirRestore(t, filepath.Dir(bin))()

	f := &updater.FakeUpdater{LatestFunc: func(context.Context, updater.LatestOpts) (updater.Release, error) {
		return updater.Release{}, updater.ErrNoAssetForPlatform
	}}
	_, stderr, code, _ := runUpgrade(t, "v0.1.0", f)
	if code == 0 {
		t.Fatal("expected non-zero exit")
	}
	if !strings.Contains(stderr, "no release artifact") {
		t.Errorf("stderr = %q", stderr)
	}
}

func TestUpgrade_CheckOnly_NewerAvailable(t *testing.T) {
	f := &updater.FakeUpdater{LatestFunc: func(context.Context, updater.LatestOpts) (updater.Release, error) {
		return updater.Release{Version: "v0.3.0"}, nil
	}}
	stdout, stderr, code, fake := runUpgrade(t, "v0.1.0", f, "--check")
	if code != 0 {
		t.Fatalf("exit=%d stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "A newer version is available: v0.3.0 (you have v0.1.0)") {
		t.Errorf("stdout = %q", stdout)
	}
	if fake.ApplyCallCount() != 0 {
		t.Errorf("Apply called %d times; want 0 (--check)", fake.ApplyCallCount())
	}
}

func TestUpgrade_CheckOnly_AlreadyLatest(t *testing.T) {
	f := &updater.FakeUpdater{LatestFunc: func(context.Context, updater.LatestOpts) (updater.Release, error) {
		return updater.Release{Version: "v0.1.0"}, nil
	}}
	stdout, _, code, _ := runUpgrade(t, "v0.1.0", f, "--check")
	if code != 0 {
		t.Fatal("expected 0 exit")
	}
	if !strings.Contains(stdout, "Already on the latest version (v0.1.0)") {
		t.Errorf("stdout = %q", stdout)
	}
}

func TestUpgrade_PreReleaseFlag(t *testing.T) {
	f := &updater.FakeUpdater{LatestFunc: func(_ context.Context, opts updater.LatestOpts) (updater.Release, error) {
		if !opts.IncludePrereleases {
			t.Fatalf("IncludePrereleases not propagated")
		}
		return updater.Release{Version: "v0.1.0"}, nil
	}}
	_, _, code, _ := runUpgrade(t, "v0.1.0", f, "--check", "--pre-release")
	if code != 0 {
		t.Fatal("expected 0 exit")
	}
}

// writableBin creates a tempdir+file so os.Executable() points at
// something writable. We can't actually replace what os.Executable
// returns from a test, but we can ensure the *current* test binary
// runs under conditions where its path is writable (the test binary
// itself lives in a tempdir per Go's test runner). For
// permission-denied coverage, see TestUpgrade_PermissionDenied below.
func writableBin(t *testing.T) string {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Skipf("os.Executable failed: %v", err)
	}
	if err := os.Chmod(exe, 0o755); err != nil {
		t.Skipf("chmod failed: %v", err)
	}
	return exe
}

func chdirRestore(t *testing.T, _ string) func() {
	t.Helper()
	return func() {}
}

func TestUpgrade_PermissionDenied(t *testing.T) {
	// We can't intercept os.Executable; instead, point a temp file at a
	// 0500-perm dir and chmod the file 0444 to simulate. This works
	// because CheckWritable uses unix.Access(W_OK) on the file itself.
	dir := t.TempDir()
	roDir := filepath.Join(dir, "ro")
	if err := os.Mkdir(roDir, 0o700); err != nil {
		t.Fatal(err)
	}
	f := filepath.Join(roDir, "inject")
	if err := os.WriteFile(f, []byte("x"), 0o444); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(roDir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(roDir, 0o700) })

	if err := updater.CheckWritable(f); err == nil {
		t.Fatal("expected CheckWritable to refuse a 0444-mode file in a 0500-mode dir")
	} else if !updater.IsNotWritable(err) {
		t.Errorf("err type = %T; want ErrNotWritable", err)
	}
}
