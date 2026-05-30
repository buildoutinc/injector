package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/alecthomas/kong"
)

// Execute parses args and runs the matched subcommand.
// Returns a POSIX-style exit code (0 success, 130 SIGINT, 1 other error).
func Execute(ctx context.Context, args []string, stdout, stderr io.Writer, version string) int {
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
	)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 1
	}

	// No args → print help to stdout and exit 0 (FR-002).
	if len(args) == 0 {
		args = []string{"--help"}
	}

	kctx, err := parser.Parse(args)
	if helpRequested {
		return 0
	}
	if err != nil {
		// A command group invoked without a subcommand should render
		// its help screen and exit 0 (FR-007). Detect kong's
		// "expected one of …" error and retry with --help.
		var pe *kong.ParseError
		if errors.As(err, &pe) && strings.HasPrefix(err.Error(), "expected ") {
			_, _ = parser.Parse(append(args, "--help"))
			if helpRequested {
				return 0
			}
		}
		_, _ = fmt.Fprintf(stderr, "inject: %s\n", err)
		_, _ = fmt.Fprintln(stderr, `Run "inject --help" for usage.`)
		return 1
	}

	if helpRequested {
		return 0
	}

	if err := kctx.Run(); err != nil {
		if errors.Is(err, context.Canceled) {
			return 130
		}
		_, _ = fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}
