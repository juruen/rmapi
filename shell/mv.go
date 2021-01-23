package shell

import (
	"errors"
	"fmt"
	"path"

	"github.com/abiosoft/ishell"
)

func mvCmd(ctx *ShellCtxt) *ishell.Cmd {
	return &ishell.Cmd{
		Name:      "mv",
		Help:      "mv file or directory",
		Completer: createEntryCompleter(ctx),
		Func: func(c *ishell.Context) {
			if len(c.Args) < 2 {
				c.Err(errors.New("missing source and/or destination"))
				return
			}

			src := c.Args[0]

			srcNode, err := ctx.api.Filetree.NodeByPath(src, ctx.node)

			if err != nil {
				c.Err(errors.New("source entry doesn't exist"))
				return
			}

			dst := c.Args[1]

			dstNode, err := ctx.api.Filetree.NodeByPath(dst, ctx.node)

			if dstNode != nil && dstNode.IsFile() {
				c.Err(errors.New("destination entry already exists"))
				return
			}

			// We are moving the node to antoher directory
			if dstNode != nil && dstNode.IsDirectory() {
				n, err := ctx.api.MoveEntry(srcNode, dstNode, srcNode.Name())

				if err != nil {
					c.Err(errors.New(fmt.Sprint("failed to move entry", err)))
					return
				}

				ctx.api.Filetree.MoveNode(srcNode, n)
				return
			}

			// We are renaming the node
			parentDir := path.Dir(dst)
			newEntry := path.Base(dst)

			parentNode, err := ctx.api.Filetree.NodeByPath(parentDir, ctx.node)

			if err != nil || parentNode.IsFile() {
				c.Err(errors.New("directory doesn't exist"))
				return
			}

			n, err := ctx.api.MoveEntry(srcNode, parentNode, newEntry)

			if err != nil {
				c.Err(errors.New(fmt.Sprint("failed to move entry", err)))
				return
			}

			ctx.api.Filetree.MoveNode(srcNode, n)
		},
	}
}
