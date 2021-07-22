package shell

import (
	"errors"
	"fmt"

	"github.com/abiosoft/ishell"
)

func rmCmd(ctx *ShellCtxt) *ishell.Cmd {
	return &ishell.Cmd{
		Name:      "rm",
		Help:      "delete entry",
		Completer: createEntryCompleter(ctx),
		Func: func(c *ishell.Context) {
			for _, target := range c.Args {
				node, err := ctx.api.Filetree().NodeByPath(target, ctx.node)

				if err != nil {
					c.Err(errors.New("entry doesn't exist"))
					return
				}

				err = ctx.api.DeleteEntry(node)

				if err != nil {
					c.Err(errors.New(fmt.Sprint("failed to delete entry", err)))
					return
				}

				ctx.api.Filetree().DeleteNode(node)
			}

			c.Println("entry(s) deleted")
		},
	}
}
