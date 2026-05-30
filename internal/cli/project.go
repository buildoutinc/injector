package cli

type ProjectCmd struct {
	Init ProjectInitCmd `cmd:"" help:"Scaffold a new Injector project."`
}
