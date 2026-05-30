package cli

import "github.com/alecthomas/kong"

const Description = "inject manages secrets and configuration for Injector projects across organizations, services, and environments."

type RootCmd struct {
	Project ProjectCmd `cmd:"" help:"Manage Injector projects."`

	Carapace CarapaceCmd `cmd:"" name:"_carapace" hidden:"" help:"Emit shell completion script."`

	Version kong.VersionFlag `help:"Show version and exit." short:"v"`
}
