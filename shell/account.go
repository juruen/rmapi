package shell

import (
	"github.com/abiosoft/ishell"
)

func accountCmd(ctx *ShellCtxt) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "account",
		Help: "account info",
		Func: func(c *ishell.Context) {
			c.Printf("User: %s, SyncVersion: %v\n", ctx.UserInfo.User, ctx.UserInfo.SyncVersion)
		},
	}
}
