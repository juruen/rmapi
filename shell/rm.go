package shell

import "github.com/abiosoft/ishell"

func rmCmd(ctx *ShellCtxt) *ishell.Cmd {
	return &ishell.Cmd{
		Name:      "rm",
		Help:      "delete entry",
		Completer: createEntryCompleter(ctx),
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				return
			}

			target := c.Args[0]

			node, err := ctx.api.Filetree.NodeByPath(target, ctx.node)

			if err != nil {
				c.Println("entry doesn't exist")
				return
			}

			err = ctx.api.DeleteEntry(node)

			if err != nil {
				c.Println("failed to delete entry", err)
				return
			}

			c.Println("entry deleted")

			ctx.api.Filetree.DeleteNode(node)
		},
	}
}
