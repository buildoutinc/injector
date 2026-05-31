package cli

import (
	"context"

	"github.com/alecthomas/kong"
)

// HelpCmd implements `inject help`. It is a thin alias that asks kong
// to print the same help screen as `inject --help`.
type HelpCmd struct{}

func (c *HelpCmd) Run(ctx context.Context, kctx *kong.Context) error {
	_ = kctx.PrintUsage(false)
	return nil
}
