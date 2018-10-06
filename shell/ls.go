package shell

import (
	"errors"

	"github.com/abiosoft/ishell"
)

func lsCmd(ctx *ShellCtxt) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "ls",
		Help: "list directory content of given path, defaults to current directory",
		Func: func(c *ishell.Context) {
			// Default to current node.
			node := ctx.node

			// Optional node path to list.
			if len(c.Args) > 0 {
				var err error
				node, err = ctx.api.Filetree.NodeByPath(c.Args[0], node)
				if err != nil || node.IsFile() {
					c.Err(errors.New("path entry doesn't exist"))
					return
				}
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
