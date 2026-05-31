package updatecheck

import (
	"context"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/blang/semver"

	"github.com/buildoutinc/injector/internal/updater"
)

// MaxCheckWait is the hard ceiling on how long Notice() will block.
const MaxCheckWait = 200 * time.Millisecond

// CacheTTL is the rate-limit window.
const CacheTTL = 24 * time.Hour

// OptOutEnv disables the passive check entirely.
const OptOutEnv = "INJECT_NO_UPDATE_CHECK"

// Notice carries everything the renderer needs.
type Notice struct {
	Current string
	Latest  string
}

// Checker is the interface the CLI consumes; tests substitute fakes.
type Checker interface {
	Start(ctx context.Context)
	Notice() *Notice
}

// BackgroundChecker is the production Checker. It honors opt-out and
// "dev" version, reads/writes the cache, and races a network probe.
type BackgroundChecker struct {
	Updater       updater.Updater
	CacheDir      string
	BinaryVersion string

	once   sync.Once
	doneCh chan struct{}
	notice *Notice
	mu     sync.Mutex
}

// Start kicks off the check. Safe to call multiple times; only the first
// call has effect. Returns immediately; Notice() blocks (up to MaxCheckWait)
// when called.
func (c *BackgroundChecker) Start(ctx context.Context) {
	c.once.Do(func() {
		c.doneCh = make(chan struct{})
		if c.skip() {
			close(c.doneCh)
			return
		}
		go c.run(ctx)
	})
}

func (c *BackgroundChecker) skip() bool {
	if os.Getenv(OptOutEnv) != "" {
		return true
	}
	if c.BinaryVersion == "" || c.BinaryVersion == "dev" {
		return true
	}
	return false
}

func (c *BackgroundChecker) run(ctx context.Context) {
	defer close(c.doneCh)

	rec, ok, _ := ReadCache(c.CacheDir)
	if ok && time.Since(rec.CheckedAt) < CacheTTL {
		c.setNoticeIfNewer(rec.LatestVersion)
		return
	}

	probeCtx, cancel := context.WithTimeout(ctx, MaxCheckWait)
	defer cancel()
	if c.Updater == nil {
		return
	}
	rel, err := c.Updater.Latest(probeCtx, updater.LatestOpts{})
	if err != nil {
		return
	}
	c.setNoticeIfNewer(rel.Version)
	_ = WriteCache(c.CacheDir, Record{
		CheckedAt:     time.Now().UTC(),
		LatestVersion: rel.Version,
		BinaryVersion: c.BinaryVersion,
	})
}

func (c *BackgroundChecker) setNoticeIfNewer(latest string) {
	if latest == "" {
		return
	}
	if !isStrictlyNewer(c.BinaryVersion, latest) {
		return
	}
	c.mu.Lock()
	c.notice = &Notice{Current: c.BinaryVersion, Latest: latest}
	c.mu.Unlock()
}

// Notice blocks for up to MaxCheckWait waiting for Start's goroutine to
// finish, then returns whatever notice (or nil) it produced. Safe to
// call without Start (returns nil).
func (c *BackgroundChecker) Notice() *Notice {
	if c.doneCh == nil {
		return nil
	}
	select {
	case <-c.doneCh:
	case <-time.After(MaxCheckWait):
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.notice
}

func isStrictlyNewer(current, latest string) bool {
	cur, err := semver.ParseTolerant(strings.TrimPrefix(current, "v"))
	if err != nil {
		return false
	}
	lat, err := semver.ParseTolerant(strings.TrimPrefix(latest, "v"))
	if err != nil {
		return false
	}
	return lat.GT(cur)
}

// Disabled is a Checker that always returns nil; used in tests and as a
// stand-in when the env var is set before Start ran.
type Disabled struct{}

func (Disabled) Start(context.Context) {}
func (Disabled) Notice() *Notice       { return nil }
