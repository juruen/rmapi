package shell

import (
	"errors"
	"fmt"
	"path"

	"github.com/abiosoft/ishell"
)

func mkdirCmd(ctx *ShellCtxt) *ishell.Cmd {
	return &ishell.Cmd{
		Name:      "mkdir",
		Help:      "create a directory",
		Completer: createDirCompleter(ctx),
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				c.Err(errors.New("missing directory"))
				return
			}

			target := c.Args[0]

			_, err := ctx.api.Filetree.NodeByPath(target, ctx.node)

			if err == nil {
				c.Println("entry already exists")
				return
			}

			parentDir := path.Dir(target)
			newDir := path.Base(target)

			if newDir == "/" || newDir == "." {
				c.Err(errors.New("invalid directory name"))
				return
			}

			parentNode, err := ctx.api.Filetree.NodeByPath(parentDir, ctx.node)

			if err != nil || parentNode.IsFile() {
				c.Err(errors.New("directory doesn't exist"))
				return
			}

			parentId := parentNode.Id()
			if parentNode.IsRoot() {
				parentId = ""
			}

			document, err := ctx.api.CreateDir(parentId, newDir)

			if err != nil {
				c.Err(errors.New(fmt.Sprint("failed to create directory", err)))
				return
			}

			ctx.api.Filetree.AddDocument(document)
		},
	}
}
