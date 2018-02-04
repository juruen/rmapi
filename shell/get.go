package shell

import (
	"fmt"

	"github.com/abiosoft/ishell"
)

func getCmd(ctx *ShellCtxt) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "get",
		Help: "copy remote file to local",
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				c.Println("missing source file")
				return
			}

			srcName := c.Args[0]

			node, err := ctx.fileTree.NodeByPath(srcName, ctx.node)

			if err != nil || node.IsDirectory() {
				c.Println("file doesn't exist")
				return
			}

			c.Println(fmt.Sprintf("downlading: [%s]...", srcName))

			err = ctx.httpCtx.FetchDocument(node.Document.ID, fmt.Sprintf("%s.zip", node.Name()))

			if err == nil {
				c.Println("OK")
				return
			}

			c.Println("Failed to downlaod file: %s", err.Error())
		},
	}
}
