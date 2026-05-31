package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/buildoutinc/injector/internal/updatecheck"
)

// VersionCmd wires `inject version`.
type VersionCmd struct{}

func (c *VersionCmd) Run(ctx context.Context, ck updatecheck.Checker, stdout io.Writer) error {
	ck.Start(ctx)
	commit, date := BuildInfo()
	fmt.Fprintf(stdout, "inject %s (commit %s, built %s)\n", Version(), commit, date)
	RenderNotice(stdout, ck.Notice(), isTTY(stdout))
	return nil
}
