package shell

import (
	"errors"
	"fmt"
	"os/exec"
	"github.com/abiosoft/ishell"
)

func getCmdA(ctx *ShellCtxt) *ishell.Cmd {
	return &ishell.Cmd{
		Name:      "geta",
		Help:      "copy remote file annotated to local",
		Completer: createEntryCompleter(ctx),
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				c.Err(errors.New("missing source file"))
				return
			}

			srcName := c.Args[0]

			node, err := ctx.api.Filetree.NodeByPath(srcName, ctx.node)

			if err != nil || node.IsDirectory() {
				c.Err(errors.New("file doesn't exist"))
				return
			}

			c.Println(fmt.Sprintf("downlading: [%s]...", srcName))

			err = ctx.api.FetchDocument(node.Document.ID, fmt.Sprintf("%s.zip", node.Name()))

			if err != nil {
				c.Err(errors.New(fmt.Sprintf("Failed to download file %s with %s", srcName, err.Error())))
				return
			}
			

			_, err = exec.Command(fmt.Sprintf("/bin/bash/rm %s.zip", node.Name())).Output()
			if err != nil {
				c.Err(err)
			}

			c.Println("OK")
		},
	}
}
