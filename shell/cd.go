package shell

import (
	"errors"

	"github.com/abiosoft/ishell"
)

func cdCmd(ctx *ShellCtxt) *ishell.Cmd {
	return &ishell.Cmd{
		Name:      "cd",
		Help:      "change directory",
		Completer: createDirCompleter(ctx),
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				return
			}

			target := c.Args[0]

			node, err := ctx.api.Filetree().NodeByPath(target, ctx.node)

			if err != nil || node.IsFile() {
				c.Err(errors.New("directory doesn't exist"))
				return
			}

			path, err := ctx.api.Filetree().NodeToPath(node)

			if err != nil || node.IsFile() {
				c.Err(errors.New("directory doesn't exist"))
				return
			}

			ctx.path = path
			ctx.node = node

			c.Println()
			c.SetPrompt(ctx.prompt())
		},
	}
}
