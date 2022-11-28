package shell

import (
	"errors"

	"github.com/abiosoft/ishell"
)

func refreshCmd(ctx *ShellCtxt) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "refresh",
		Help: "refreshes the tree with remote changes",
		Func: func(c *ishell.Context) {
			err := ctx.api.Refresh()
			if err != nil {
				c.Err(err)
				return
			}
			n, err := ctx.api.Filetree().NodeByPath(ctx.path, nil)
			if err != nil {
				c.Err(errors.New("current path is invalid"))

				ctx.node = ctx.api.Filetree().Root()
				ctx.path = ctx.node.Name()
				c.SetPrompt(ctx.prompt())
				return
			}
			ctx.node = n
		},
	}
}
