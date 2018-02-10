package shell

import "github.com/abiosoft/ishell"

func rmCmd(ctx *ShellCtxt) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "rm",
		Help: "delete entry",
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				return
			}

			target := c.Args[0]

			node, err := ctx.fileTree.NodeByPath(target, ctx.node)

			if err != nil {
				c.Println("entry doesn't exist")
				return
			}

			err = ctx.httpCtx.DeleteEntry(node)

			if err != nil {
				c.Println("failed to delete entry", err)
				return
			}

			c.Println("entry deleted")

			ctx.fileTree.DeleteNode(node)
		},
	}
}
