package cli

import (
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

type CarapaceCmd struct {
	Shell string `arg:"" optional:"" help:"Target shell (bash|zsh|fish|elvish|powershell|nushell|xonsh|tcsh)."`
}

func (c *CarapaceCmd) Run(kctx *kong.Context) error {
	shell := c.Shell
	if shell == "" {
		shell = "bash"
	}
	root := &cobra.Command{Use: kctx.Model.Name, Short: kctx.Model.Help}
	addNode(root, kctx.Model.Node)
	carapace.Gen(root)
	out, err := carapace.Gen(root).Snippet(shell)
	if err != nil {
		return fmt.Errorf("generate completion: %w", err)
	}
	_, _ = fmt.Fprintln(kctx.Stdout, out)
	return nil
}

func addNode(parent *cobra.Command, n *kong.Node) {
	for _, child := range n.Children {
		if child.Type != kong.CommandNode || child.Hidden {
			continue
		}
		sub := &cobra.Command{Use: child.Name, Short: child.Help, Run: func(*cobra.Command, []string) {}}
		addNode(sub, child)
		parent.AddCommand(sub)
	}
}
