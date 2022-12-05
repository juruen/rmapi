package shell

import (
	"errors"
	"fmt"

	"github.com/abiosoft/ishell"
	flag "github.com/ogier/pflag"
)

func rmCmd(ctx *ShellCtxt) *ishell.Cmd {
	return &ishell.Cmd{
		Name:      "rm",
		Help:      "delete entry",
		Completer: createEntryCompleter(ctx),
		Func: func(c *ishell.Context) {
			flagSet := flag.NewFlagSet("rm", flag.ContinueOnError)
			recursive := flagSet.BoolP("recursive", "r", false, "remove non empty folders")
			if err := flagSet.Parse(c.Args); err != nil {
				if err != flag.ErrHelp {
					c.Err(err)
				}
				return
			}
			argRest := flagSet.Args()
			if len(argRest) < 1 {
				c.Err(errors.New("missing param"))
				return
			}

			for _, target := range argRest {
				nodes, err := ctx.api.Filetree().NodesByPath(target, ctx.node, false)

				if err != nil {
					c.Err(err)
					return
				}
				for _, node := range nodes {
					c.Println("deleting: ", node.Name())
					err = ctx.api.DeleteEntry(node, *recursive, false)

					if err != nil {
						c.Err(fmt.Errorf("failed to delete entry, %v", err))
						return
					}

					ctx.api.Filetree().DeleteNode(node)
				}
			}

			err := ctx.api.SyncComplete()
			if err != nil {
				c.Err(err)
			}
		},
	}
}
