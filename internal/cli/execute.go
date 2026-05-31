package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/alecthomas/kong"

	"github.com/buildoutinc/injector/internal/updatecheck"
	"github.com/buildoutinc/injector/internal/updater"
)

// Options lets tests substitute the Updater and Checker. nil fields are
// filled in with production defaults (real GitHub Updater and a
// BackgroundChecker rooted at the user cache dir).
type Options struct {
	Updater updater.Updater
	Checker updatecheck.Checker
}

// Execute parses args and runs the matched subcommand with production
// defaults. Returns a POSIX-style exit code.
func Execute(ctx context.Context, args []string, stdout, stderr io.Writer, version string) int {
	return ExecuteWith(ctx, args, stdout, stderr, version, Options{})
}

// ExecuteWith is Execute plus the ability to inject Updater/Checker.
func ExecuteWith(ctx context.Context, args []string, stdout, stderr io.Writer, version string, opts Options) int {
	resetNoticeOnce()
	u := opts.Updater
	if u == nil {
		u = updater.NewGithub("buildoutinc/injector", Version())
	}
	ck := opts.Checker
	if ck == nil {
		dir, err := updatecheck.DefaultCacheDir()
		if err != nil {
			ck = updatecheck.Disabled{}
		} else {
			ck = &updatecheck.BackgroundChecker{
				Updater:       u,
				CacheDir:      dir,
				BinaryVersion: Version(),
			}
		}
	}
	ck.Start(ctx)

	var helpRequested bool
	var root RootCmd
	parser, err := kong.New(&root,
		kong.Name("inject"),
		kong.Description(Description),
		kong.UsageOnError(),
		kong.Vars{"version": version},
		kong.Writers(stdout, stderr),
		kong.Exit(func(int) { helpRequested = true }),
		kong.BindTo(ctx, (*context.Context)(nil)),
		kong.BindTo(stdout, (*io.Writer)(nil)),
		kong.Bind(Stderr{stderr}),
		kong.BindTo(u, (*updater.Updater)(nil)),
		kong.BindTo(ck, (*updatecheck.Checker)(nil)),
	)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 1
	}

	if len(args) == 0 {
		args = []string{"--help"}
	}

	kctx, err := parser.Parse(args)
	if helpRequested {
		RenderNotice(stdout, ck.Notice(), isTTY(stdout))
		return 0
	}
	if err != nil {
		var pe *kong.ParseError
		if errors.As(err, &pe) && strings.HasPrefix(err.Error(), "expected ") {
			_, _ = parser.Parse(append(args, "--help"))
			if helpRequested {
				RenderNotice(stdout, ck.Notice(), isTTY(stdout))
				return 0
			}
		}
		_, _ = fmt.Fprintf(stderr, "inject: %s\n", err)
		_, _ = fmt.Fprintln(stderr, `Run "inject --help" for usage.`)
		return 1
	}

	if helpRequested {
		RenderNotice(stdout, ck.Notice(), isTTY(stdout))
		return 0
	}

	if err := kctx.Run(); err != nil {
		if errors.Is(err, context.Canceled) {
			return 130
		}
		if err.Error() != "" {
			_, _ = fmt.Fprintln(stderr, err)
		}
		return 1
	}
	return 0
}
