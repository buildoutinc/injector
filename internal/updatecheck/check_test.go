package updatecheck_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/buildoutinc/injector/internal/updatecheck"
	"github.com/buildoutinc/injector/internal/updater"
)

func newChecker(t *testing.T, u updater.Updater, binVersion string) *updatecheck.BackgroundChecker {
	t.Helper()
	return &updatecheck.BackgroundChecker{
		Updater:       u,
		CacheDir:      t.TempDir(),
		BinaryVersion: binVersion,
	}
}

func TestCheck_RespectsOptOut(t *testing.T) {
	t.Setenv(updatecheck.OptOutEnv, "1")
	f := &updater.FakeUpdater{LatestFunc: func(context.Context, updater.LatestOpts) (updater.Release, error) {
		t.Fatal("Latest should not be called when opted out")
		return updater.Release{}, nil
	}}
	c := newChecker(t, f, "v0.1.0")
	c.Start(context.Background())
	if n := c.Notice(); n != nil {
		t.Fatalf("expected nil notice, got %+v", n)
	}
	if f.LatestCallCount() != 0 {
		t.Fatalf("expected 0 Latest calls, got %d", f.LatestCallCount())
	}
}

func TestCheck_DevVersionSkipsCheck(t *testing.T) {
	f := &updater.FakeUpdater{LatestFunc: func(context.Context, updater.LatestOpts) (updater.Release, error) {
		t.Fatal("Latest should not be called for dev builds")
		return updater.Release{}, nil
	}}
	c := newChecker(t, f, "dev")
	c.Start(context.Background())
	if n := c.Notice(); n != nil {
		t.Fatalf("expected nil notice for dev, got %+v", n)
	}
}

func TestCheck_FreshCacheSkipsNetwork(t *testing.T) {
	f := &updater.FakeUpdater{LatestFunc: func(context.Context, updater.LatestOpts) (updater.Release, error) {
		t.Fatal("Latest should not be called when cache is fresh")
		return updater.Release{}, nil
	}}
	c := newChecker(t, f, "v0.1.0")
	_ = updatecheck.WriteCache(c.CacheDir, updatecheck.Record{
		CheckedAt:     time.Now().UTC().Add(-1 * time.Hour),
		LatestVersion: "v0.3.0",
		BinaryVersion: "v0.1.0",
	})
	c.Start(context.Background())
	n := c.Notice()
	if n == nil || n.Latest != "v0.3.0" || n.Current != "v0.1.0" {
		t.Fatalf("notice = %+v", n)
	}
}

func TestCheck_StaleCacheTriggersNetwork(t *testing.T) {
	f := &updater.FakeUpdater{LatestFunc: func(context.Context, updater.LatestOpts) (updater.Release, error) {
		return updater.Release{Version: "v0.4.0"}, nil
	}}
	c := newChecker(t, f, "v0.1.0")
	_ = updatecheck.WriteCache(c.CacheDir, updatecheck.Record{
		CheckedAt:     time.Now().UTC().Add(-25 * time.Hour),
		LatestVersion: "v0.2.0",
		BinaryVersion: "v0.1.0",
	})
	c.Start(context.Background())
	n := c.Notice()
	if n == nil || n.Latest != "v0.4.0" {
		t.Fatalf("notice = %+v", n)
	}
	if f.LatestCallCount() != 1 {
		t.Fatalf("expected 1 Latest call, got %d", f.LatestCallCount())
	}
}

func TestCheck_NetworkErrorIsSwallowed(t *testing.T) {
	f := &updater.FakeUpdater{LatestFunc: func(context.Context, updater.LatestOpts) (updater.Release, error) {
		return updater.Release{}, errors.New("connection refused")
	}}
	c := newChecker(t, f, "v0.1.0")
	c.Start(context.Background())
	if n := c.Notice(); n != nil {
		t.Fatalf("expected nil on error, got %+v", n)
	}
}

func TestCheck_Timeout(t *testing.T) {
	block := make(chan struct{})
	defer close(block)
	f := &updater.FakeUpdater{LatestFunc: func(ctx context.Context, _ updater.LatestOpts) (updater.Release, error) {
		select {
		case <-ctx.Done():
			return updater.Release{}, ctx.Err()
		case <-block:
			return updater.Release{}, nil
		}
	}}
	c := newChecker(t, f, "v0.1.0")
	start := time.Now()
	c.Start(context.Background())
	n := c.Notice()
	if elapsed := time.Since(start); elapsed > 300*time.Millisecond {
		t.Fatalf("Notice() took %v; budget is %v", elapsed, updatecheck.MaxCheckWait)
	}
	if n != nil {
		t.Fatalf("expected nil notice on timeout, got %+v", n)
	}
}

func TestCheck_UpToDateNoNotice(t *testing.T) {
	f := &updater.FakeUpdater{LatestFunc: func(context.Context, updater.LatestOpts) (updater.Release, error) {
		return updater.Release{Version: "v0.1.0"}, nil
	}}
	c := newChecker(t, f, "v0.1.0")
	c.Start(context.Background())
	if n := c.Notice(); n != nil {
		t.Fatalf("expected nil when up to date, got %+v", n)
	}
}
