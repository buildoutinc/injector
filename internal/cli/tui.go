package cli

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
)

// Run starts a Bubble Tea program with the given model, cancelling on ctx.
// Reserved for future interactive subcommands; not exercised by the
// current command tree.
func Run(ctx context.Context, m tea.Model) error {
	_, err := tea.NewProgram(m, tea.WithContext(ctx)).Run()
	return err
}
