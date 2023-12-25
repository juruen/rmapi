package shell

import (
	"errors"
	"fmt"

	"github.com/abiosoft/ishell"
)

func getCmd(ctx *ShellCtxt) *ishell.Cmd {
	return &ishell.Cmd{
		Name:      "get",
		Help:      "copy remote file to local",
		Completer: createEntryCompleter(ctx),
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				c.Err(errors.New("missing source file"))
				return
			}

			srcName := c.Args[0]

			node, err := ctx.api.Filetree().NodeByPath(srcName, ctx.node)

			if err != nil || node.IsDirectory() {
				c.Err(errors.New("file doesn't exist"))
				return
			}

			c.Println(fmt.Sprintf("downloading: [%s]...", srcName))

			if len(c.Args) > 1 {
				dest := c.Args[1]
				err = ctx.api.FetchDocument(node.Document.ID, fmt.Sprintf("%s/%s.zip", dest, node.Name()))
			} else {
				err = ctx.api.FetchDocument(node.Document.ID, fmt.Sprintf("%s.zip", node.Name()))
			}

			if err == nil {
				c.Println("OK")
				return
			}

			c.Err(errors.New(fmt.Sprintf("Failed to download file %s with %s", srcName, err.Error())))
		},
	}
}
