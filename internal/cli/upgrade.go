package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/blang/semver"

	"github.com/buildoutinc/injector/internal/updater"
)

// UpgradeCmd wires `inject upgrade`.
type UpgradeCmd struct {
	Check      bool `help:"Only report whether an upgrade is available; do not modify anything."`
	PreRelease bool `name:"pre-release" help:"Include pre-release tags."`
}

// Run is invoked by kong. `u` and `stderr` are injected via kong.Bind.
func (c *UpgradeCmd) Run(ctx context.Context, u updater.Updater, stdout io.Writer, stderrW Stderr) error {
	stderr := stderrW.Writer
	binPath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(stderr, "inject: cannot upgrade: locating binary: %v\n", err)
		return errExit
	}

	if !c.Check {
		if perr := updater.CheckWritable(binPath); perr != nil {
			fmt.Fprintf(stderr,
				"inject: cannot upgrade: %s is not writable by your user — install via your package manager (e.g. brew upgrade injector)\n",
				binPath)
			return errExit
		}
	}

	current := Version()
	rel, err := u.Latest(ctx, updater.LatestOpts{IncludePrereleases: c.PreRelease})
	if err != nil {
		if errors.Is(err, updater.ErrNoAssetForPlatform) {
			fmt.Fprintf(stderr, "inject: cannot upgrade: no release artifact for %s/%s\n", runtime.GOOS, runtime.GOARCH)
			return errExit
		}
		fmt.Fprintf(stderr, "inject: cannot upgrade: %v\n", err)
		return errExit
	}

	newer, err := isStrictlyNewer(current, rel.Version)
	if err != nil {
		// Treat unparsable current version (e.g. "dev") as "not up to date".
		newer = true
	}

	if c.Check {
		if newer {
			fmt.Fprintf(stdout, "A newer version is available: %s (you have %s)\n", rel.Version, current)
		} else {
			fmt.Fprintf(stdout, "Already on the latest version (%s)\n", rel.Version)
		}
		return nil
	}

	if !newer {
		fmt.Fprintf(stdout, "inject: already up to date (%s)\n", rel.Version)
		return nil
	}

	if _, err := u.Apply(ctx, rel, binPath); err != nil {
		if errors.Is(err, updater.ErrChecksumMismatch) {
			fmt.Fprintln(stderr, "inject: cannot upgrade: checksum mismatch — release tampered or transfer corrupt")
			return errExit
		}
		fmt.Fprintf(stderr, "inject: cannot upgrade: %v\n", err)
		return errExit
	}
	fmt.Fprintf(stdout, "inject: upgraded %s → %s\n", current, rel.Version)
	return nil
}

// errExit is returned to kong to signal a non-zero exit. The Execute
// function maps any non-nil, non-context-canceled error from Run to
// exit code 1, and we don't want a second error string printed.
var errExit = errors.New("")

func isStrictlyNewer(current, latest string) (bool, error) {
	cur, err := semver.ParseTolerant(strings.TrimPrefix(current, "v"))
	if err != nil {
		return false, err
	}
	lat, err := semver.ParseTolerant(strings.TrimPrefix(latest, "v"))
	if err != nil {
		return false, err
	}
	return lat.GT(cur), nil
}
