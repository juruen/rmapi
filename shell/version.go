package shell

import (
	"github.com/abiosoft/ishell"
	"github.com/peerdavid/rmapi/version"
)

func versionCmd(ctx *ShellCtxt) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "version",
		Help: "show rmapi version",
		Func: func(c *ishell.Context) {
			c.Println("rmapi version:", version.Version)
		},
	}
}
