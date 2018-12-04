package shell

import (
	"errors"
	"fmt"

	"github.com/abiosoft/ishell"
	"github.com/juruen/rmapi/util"
)

func putCmd(ctx *ShellCtxt) *ishell.Cmd {
	return &ishell.Cmd{
		Name:      "put",
		Help:      "copy a local document to cloud",
		Completer: createFsEntryCompleter(),
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				c.Println("missing source file")
				return
			}

			srcName := c.Args[0]

			docName := util.DocPathToName(srcName)

			_, err := ctx.api.Filetree.NodeByPath(docName, ctx.node)
			if err == nil {
				c.Err(errors.New("entry already exists"))
				return
			}

			dstDir := ctx.node.Id()
			if len(c.Args) == 2 {
				node, err := ctx.api.Filetree.NodeByPath(c.Args[1], ctx.node)

				if err != nil || node.IsFile() {
					c.Err(errors.New("directory doesn't exist"))
					return
				}

				dstDir = node.Id()
			}

			c.Printf("uploading: [%s]...", srcName)

			document, err := ctx.api.UploadDocument(dstDir, srcName)

			if err != nil {
				c.Err(errors.New(fmt.Sprint("Failed to upload file", srcName, err.Error())))
				return
			}

			c.Println("OK")

			ctx.api.Filetree.AddDocument(*document)
		},
	}
}
