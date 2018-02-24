package shell

import (
	"github.com/abiosoft/ishell"
)

func statCmd(ctx *ShellCtxt) *ishell.Cmd {
	return &ishell.Cmd{
		Name:      "stat",
		Help:      "fetch entry metatada",
		Completer: createEntryCompleter(ctx),
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				c.Println("missing source file")
				return
			}

			srcName := c.Args[0]

			node, err := ctx.api.Filetree.NodeByPath(srcName, ctx.node)

			if err != nil {
				c.Println("file doesn't exist")
				return
			}

			c.Printf("%+v\n", node.Document)
		},
	}
}
