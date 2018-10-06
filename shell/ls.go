package shell

import (
	"errors"

	"github.com/abiosoft/ishell"
)

func lsCmd(ctx *ShellCtxt) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "ls",
		Help: "list directory",
		Func: func(c *ishell.Context) {
			node := ctx.node

			if len(c.Args) == 1 {
				target := c.Args[0]

				argNode, err := ctx.api.Filetree.NodeByPath(target, ctx.node)

				if err != nil || node.IsFile() {
					c.Err(errors.New("directory doesn't exist"))
					return
				}

				node = argNode
			}

			for _, e := range node.Children {
				eType := "d"
				if e.IsFile() {
					eType = "f"
				}
				c.Printf("[%s]\t%s\n", eType, e.Name())
			}
		},
	}
}
