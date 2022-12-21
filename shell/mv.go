package shell

import (
	"errors"
	"fmt"
	"path"

	"github.com/abiosoft/ishell"
	"github.com/juruen/rmapi/model"
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
			dst := c.Args[1]

			srcNodes, err := ctx.api.Filetree().NodesByPath(src, ctx.node, false)

			if err != nil {
				c.Err(err)
				return
			}
			if len(srcNodes) < 1 {
				c.Err(errors.New("no nodes found"))
				return
			}

			dstNode, _ := ctx.api.Filetree().NodeByPath(dst, ctx.node)

			if dstNode != nil && dstNode.IsFile() {
				c.Err(errors.New("destination entry already exists"))
				return
			}

			// We are moving the node to another directory
			if dstNode != nil && dstNode.IsDirectory() {
				for _, node := range srcNodes {
					if isSubdir(node, dstNode) {
						c.Err(fmt.Errorf("cannot move: %s in itself", node.Name()))
						return
					}

					n, err := ctx.api.MoveEntry(node, dstNode, node.Name())

					if err != nil {
						c.Err(fmt.Errorf("failed to move entry %w", err))
						return
					}

					ctx.api.Filetree().MoveNode(node, n)
				}
				err = ctx.api.SyncComplete()
				if err != nil {
					c.Err(fmt.Errorf("cannot notify, %w", err))
				}
				return
			}

			if len(srcNodes) > 1 {
				c.Err(errors.New("cannot rename multiple nodes, only first match will be renamed"))
			}

			srcNode := srcNodes[0]

			// We are renaming the node
			parentDir := path.Dir(dst)
			newEntry := path.Base(dst)

			parentNode, err := ctx.api.Filetree().NodeByPath(parentDir, ctx.node)

			if err != nil || parentNode.IsFile() {
				c.Err(fmt.Errorf("cannot move, %w", err))
				return
			}

			n, err := ctx.api.MoveEntry(srcNode, parentNode, newEntry)

			if err != nil {
				c.Err(fmt.Errorf("failed to move entry, %w", err))
				return
			}
			err = ctx.api.SyncComplete()
			if err != nil {
				c.Err(fmt.Errorf("cannot notify, %w", err))
			}

			ctx.api.Filetree().MoveNode(srcNode, n)
		},
	}
}

// isSubdir check for moves e.g. a in a/sub1 which result in data loss
func isSubdir(parent *model.Node, child *model.Node) bool {
	for child != nil {
		if parent.Id() == child.Id() {
			return true
		}
		child = child.Parent
	}
	return false
}
