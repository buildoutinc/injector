package cli

const Description = "inject manages secrets and configuration for Injector projects across organizations, services, and environments."

type RootCmd struct {
	Project ProjectCmd `cmd:"" help:"Manage Injector projects."`
	Upgrade UpgradeCmd `cmd:"" help:"Upgrade the inject binary to the latest GitHub release."`
	Version VersionCmd `cmd:"" help:"Show inject version."`
	Help    HelpCmd    `cmd:"" help:"Show help."`

	Carapace CarapaceCmd `cmd:"" name:"_carapace" hidden:"" help:"Emit shell completion script."`
}
