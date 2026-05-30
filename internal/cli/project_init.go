package cli

import (
	"context"
	"io"
)

type ProjectInitCmd struct{}

func (c *ProjectInitCmd) Run(ctx context.Context, w io.Writer) error {
	_, err := io.WriteString(w, "TODO: init!\n")
	return err
}
